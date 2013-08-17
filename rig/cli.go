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
	"net/url"
	"os"
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
		{"ps", "Show running processes"},
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
	cmd := c.Subcmd("list", "", "List stacks, services and processes")
	if err := cmd.Parse(args); err != nil {
		return nil
	}

	if cmd.NArg() > 0 {
		cmd.Usage()
		return nil
	}

	body, _, err := c.call("GET", "/list", nil)
	if err != nil {
		return err
	}

	var stacks map[string]map[string][]string
	err = json.Unmarshal(body, &stacks)
	if err != nil {
		fmt.Printf("Error unmarshal: body: %s, err: %s\n", body, err)
		return err
	}

	fmt.Println("Stack list:\n")
	for stackName, s := range stacks {
		fmt.Printf("\x1b[1m- %s :\x1b[0m\n", stackName)
		for serviceName, svc := range s {
			fmt.Printf("      %s :\n", serviceName)
			for _, processName := range svc {
				fmt.Printf("        %s\n", processName)
			}
		}
	}
	fmt.Println("")

	return nil
}

func (c *Cli) CmdPs(args ...string) error {
	cmd := c.Subcmd("ps", "", "Show running processes")
	if err := cmd.Parse(args); err != nil {
		return nil
	}

	if cmd.NArg() > 0 {
		cmd.Usage()
		return nil
	}

	body, _, err := c.call("GET", "/ps", nil)
	if err != nil {
		return err
	}

	var stacks map[string]map[string][]*rig.ApiProcess
	err = json.Unmarshal(body, &stacks)
	if err != nil {
		fmt.Printf("Error unmarshal: body: %s, err: %s\n", body, err)
		return err
	}

	fmt.Println("PID    Name           Status")
	for stackName, s := range stacks {
		for serviceName, svc := range s {
			for _, process := range svc {
				var status string
				if process.Status == 1 {
					status = "Running"
				} else {
					status = "Not running"
				}
				fmt.Printf("%d  %s:%s:%s       %s\n", process.Pid, stackName, serviceName, process.Name, status)
			}
		}
	}

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

	path, err := c.resolve(cmd.Arg(0))
	if err != nil {
		return err
	}
	path += "/start"

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

	path, err := c.resolve(cmd.Arg(0))
	if err != nil {
		return err
	}
	path += "/stop"

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

	path, err := c.resolve(cmd.Arg(0))
	if err != nil {
		return err
	}
	path += "/tail"

	err = c.stream("POST", path, nil)
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

func (c *Cli) resolve(descriptor string) (string, error) {
	v := url.Values{}
	v.Set("descriptor", descriptor)
	if pwd, err := os.Getwd(); err == nil {
		v.Set("pwd", pwd)
	}

	resolveBody, _, err := c.call("GET", "/resolve?"+v.Encode(), nil)
	if err != nil {
		return "", err
	}

	var d rig.Descriptor
	err = json.Unmarshal(resolveBody, &d)
	if err != nil {
		return "", fmt.Errorf("Error unmarshal: body: %s, err: %s\n", resolveBody, err)
	}

	if d.Stack == "" {
		return "", fmt.Errorf("Error : resolver couldn't find stack")
	}

	path := fmt.Sprintf("/%s", d.Stack)

	if d.Service != "" {
		path += fmt.Sprintf("/%s", d.Service)
	}

	if d.Process != "" {
		path += fmt.Sprintf("/%s", d.Process)
	}

	return path, nil
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

	logger := NewProcessLogger()
	dec := json.NewDecoder(resp.Body)
	for {
		m := rig.ProcessOutputMessage{}
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		logger.Println(m)
	}
	return nil
}
