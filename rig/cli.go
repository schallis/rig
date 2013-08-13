package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gocardless/rig"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Cli struct {
	client *http.Client
	proto  string
	addr   string
	out    io.Writer
	err    io.Writer
}

func NewCli(proto, addr string) *Cli {
	cli := &Cli{
		client: &http.Client{},
		proto:  proto,
		addr:   addr,
		out:    os.Stdout,
		err:    os.Stderr,
	}
	return cli
}

func (c *Cli) ParseCommand(args ...string) error {
	cmds := map[string]func(args ...string) error{
		"help":    c.CmdHelp,
		"list":    c.CmdList,
		"ps":      c.CmdPs,
		"restart": c.CmdRestart,
		"start":   c.CmdStart,
		"stop":    c.CmdStop,
		"tail":    c.CmdTail,
		"version": c.CmdVersion,
	}

	if len(args) > 0 {
		if cmd, exists := cmds[args[0]]; exists {
			return cmd(args[1:]...)
		}
	}

	return cmds["help"](args...)
}

func (c *Cli) Subcmd(name, signature, description string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Printf("\nUsage: rig %s %s\n\n%s\n\n", name, signature, description)
		flags.PrintDefaults()
	}
	return flags
}

func (c *Cli) CmdHelp(args ...string) error {
	help := "Usage: rig [OPTIONS] COMMAND DESCRIPTOR \n\nCommands:\n"
	for _, cmd := range [][]string{
		{"help", "Show rig help"},
		{"list", "List stacks, services and processes"},
		{"ps", "Show running processes status"},
		{"restart", "Restart a stack, a service or a process"},
		{"start", "Start a stack, a service or a process"},
		{"stop", "Stop a stack, a service or a process"},
		{"tail", "Tail logs of a stack, a service or a process"},
		{"version", "Show the rig version"},
	} {
		help += fmt.Sprintf("    %-10.10s%s\n", cmd[0], cmd[1])
	}
	fmt.Println(help)
	return nil
}

func (c *Cli) CmdList(args ...string) error {
	return nil
}

func (c *Cli) CmdPs(args ...string) error {
	return nil
}

func (c *Cli) CmdRestart(args ...string) error {
	return nil
}

func (c *Cli) CmdStart(args ...string) error {
	cmd := c.Subcmd("start", "DESCRIPTOR", "Start a stack, a service or a process")
	if err := cmd.Parse(args); err != nil {
		return nil
	}

	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	descriptor := strings.Split(cmd.Arg(0), ":")
	l := len(descriptor)

	var stack, service, process string
	var path string

	stack = descriptor[0]
	path = fmt.Sprintf("/%s/start", stack)

	if l > 1 {
		service = descriptor[1]
		path = fmt.Sprintf("/%s/%s/start", stack, service)
	}
	if l > 2 {
		process = descriptor[2]
		path = fmt.Sprintf("/%s/%s/%s/start", stack, service, process)
	}

	// resolveBody, _, err := c.call("POST", "/resolve", nil)
	// if err != nil {
	// 	return err
	// }

	body, _, err := c.call("POST", path, nil)
	if err != nil {
		return err
	}

	fmt.Println(body)

	return nil
}

func (c *Cli) CmdStop(args ...string) error {
	cmd := c.Subcmd("stop", "DESCRIPTOR", "Stop a stack, a service or a process")
	if err := cmd.Parse(args); err != nil {
		return nil
	}

	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	descriptor := strings.Split(cmd.Arg(0), ":")
	l := len(descriptor)

	var stack, service, process string
	var path string

	stack = descriptor[0]
	path = fmt.Sprintf("/%s/stop", stack)

	if l > 1 {
		service = descriptor[1]
		path = fmt.Sprintf("/%s/%s/stop", stack, service)
	}
	if l > 2 {
		process = descriptor[2]
		path = fmt.Sprintf("/%s/%s/%s/stop", stack, service, process)
	}

	// resolveBody, _, err := c.call("POST", "/resolve", nil)
	// if err != nil {
	// 	return err
	// }

	body, _, err := c.call("POST", path, nil)
	if err != nil {
		return err
	}

	fmt.Println(body)

	return nil
}

func (c *Cli) CmdTail(args ...string) error {
	cmd := c.Subcmd("stop", "DESCRIPTOR", "Stop a stack, a service or a process")
	if err := cmd.Parse(args); err != nil {
		return nil
	}

	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	descriptor := strings.Split(cmd.Arg(0), ":")
	l := len(descriptor)

	var stack, service, process string
	var path string

	stack = descriptor[0]
	path = fmt.Sprintf("/%s/tail", stack)

	if l > 1 {
		service = descriptor[1]
		path = fmt.Sprintf("/%s/%s/tail", stack, service)
	}
	if l > 2 {
		process = descriptor[2]
		path = fmt.Sprintf("/%s/%s/%s/tail", stack, service, process)
	}

	// resolveBody, _, err := c.call("POST", "/resolve", nil)
	// if err != nil {
	// 	return err
	// }

	err := c.stream("POST", path, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cli) CmdVersion(args ...string) error {
	cmd := c.Subcmd("version", "", "Show the rig version")
	if err := cmd.Parse(args); err != nil {
		return nil
	}

	if cmd.NArg() > 0 {
		cmd.Usage()
		return nil
	}

	body, _, err := c.call("GET", "/version", nil)
	if err != nil {
		return err
	}

	var out rig.ApiVersion
	err = json.Unmarshal(body, &out)
	if err != nil {
		fmt.Printf("Error unmarshal: body: %s, err: %s\n", body, err)
		return err
	}
	fmt.Println("Version:", out.Version)

	return nil
}

func (c *Cli) call(method, path string, data interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if data != nil {
		buf, err := json.Marshal(data)
		if err != nil {
			return nil, -1, err
		}
		reqBody = bytes.NewBuffer(buf)
	}

	urlStr := fmt.Sprintf("%s://%s%s", c.proto, c.addr, path)
	req, err := http.NewRequest(method, urlStr, reqBody)
	if err != nil {
		return nil, -1, err
	}
	req.Header.Set("User-Agent", "Rig-Client/"+rig.Version)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, -1, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, -1, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		if len(body) == 0 {
			return nil, resp.StatusCode, fmt.Errorf("Error: %s", http.StatusText(resp.StatusCode))
		}
		return nil, resp.StatusCode, fmt.Errorf("Error: %s", body)
	}

	return body, resp.StatusCode, nil
}

func (c *Cli) stream(method, path string, data interface{}) error {
	var reqBody io.Reader
	if data != nil {
		buf, err := json.Marshal(data)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(buf)
	}

	urlStr := fmt.Sprintf("%s://%s%s", c.proto, c.addr, path)
	req, err := http.NewRequest(method, urlStr, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Rig-Client/"+rig.Version)

	resp, err := c.client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if len(body) == 0 {
			return fmt.Errorf("Error :%s", http.StatusText(resp.StatusCode))
		}
		return fmt.Errorf("Error: %s", body)
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	for {
		m := rig.ProcessOutputMessage{}
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		fmt.Fprintf(c.out, "%+v\r\n", m)
	}
	return nil
}
