package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Command struct {
	CommandName string
	Args        []string
	Stdout      *os.File
	Stderr      *os.File
}

func NewCommand(input string) (*Command, error) {
	splitInput, err := CoalesceQuotes(input)
	if err != nil {
		return nil, err
	}

	command, args := splitInput[0], splitInput[1:]

	args, outputFile, errFile, err := ParseRedirection(args)
	if err != nil {
		return nil, err
	}

	return &Command{
		CommandName: command,
		Args:        args,
		Stdout:      outputFile,
		Stderr:      errFile,
	}, nil

}

func (c *Command) Exec() error {
	if handler, ok := GetBuiltInCommands()[c.CommandName]; ok {
		return handler(c)
	}
	return HandlerExec(c)
}

func (c *Command) Close() {
	if c.Stdout != os.Stdout {
		c.Stdout.Close()
	}
	if c.Stderr != os.Stderr {
		c.Stderr.Close()
	}
}

func HandlerExec(cmd *Command) error {
	cmdExec := exec.Command(cmd.CommandName, cmd.Args...)
	if cmdExec.Err != nil {
		return fmt.Errorf("%s: command not found", cmd.CommandName)
	}
	//fmt.Printf("Path: '%s', Dir: '%s'\r\n", cmd.Path, cmd.Dir)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmdExec.Stdin = os.Stdin
	cmdExec.Stdout = &stdout
	cmdExec.Stderr = &stderr

	cmdExec.Run()
	// if err != nil {
	// 	return fmt.Errorf("error: %s", err)
	// }

	if stdout.Len() != 0 {
		fmt.Fprint(cmd.Stdout, strings.ReplaceAll(stdout.String(), "\n", "\r\n"))
	}
	if stderr.Len() != 0 {
		fmt.Fprint(cmd.Stderr, strings.ReplaceAll(stderr.String(), "\n", "\r\n"))
	}

	return nil
}
