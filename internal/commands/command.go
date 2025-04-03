package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Command struct {
	commandName string
	args        []string
	stdout      *os.File
	stderr      *os.File
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
		commandName: command,
		args:        args,
		stdout:      outputFile,
		stderr:      errFile,
	}, nil

}

func (c *Command) Exec() error {
	if handler, ok := GetBuiltInCommands()[c.commandName]; ok {
		return handler(c)
	}
	return HandlerExec(c)
}

func (c *Command) Close() {
	if c.stdout != os.Stdout {
		c.stdout.Close()
	}
	if c.stderr != os.Stderr {
		c.stdout.Close()
	}
}

func HandlerExec(cmd *Command) error {
	cmdExec := exec.Command(cmd.commandName, cmd.args...)
	if cmdExec.Err != nil {
		return fmt.Errorf("%s: command not found", cmd.commandName)
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
		fmt.Fprint(cmd.stdout, strings.ReplaceAll(stdout.String(), "\n", "\r\n"))
	}
	if stderr.Len() != 0 {
		fmt.Fprint(cmd.stderr, strings.ReplaceAll(stderr.String(), "\n", "\r\n"))
	}

	return nil
}
