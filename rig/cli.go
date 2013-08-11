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
	"strings"
)

type Cli struct {
	client *http.Client
	proto  string
	addr   string
}

func NewCli(proto, addr string) *Cli {
	cli := &Cli{
		client: &http.Client{},
		proto:  proto,
		addr:   addr,
	}
	return cli
}

func (c *Cli) ParseCommand(args ...string) error {
	cmds := map[string]func(args ...string) error{
		"help":    c.CmdHelp,
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
		{"restart", "Restart a stack, a service or a process"},
		{"start", "Start a stack, a service or a process"},
		{"stop", "Stop a stack, a service or a process"},
		{"tail", "Tail logs of a stack, a service or a process"},
		{"help", "Show rig help"},
		{"version", "Show the rig version"},
	} {
		help += fmt.Sprintf("    %-10.10s%s\n", cmd[0], cmd[1])
	}
	fmt.Println(help)
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

	return body, resp.StatusCode, nil
}
