package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type IOStream struct {
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}

type Command struct {
	CommandName string
	Args        []string
	IOStream
}

func NewCommand(input string) (*Command, error) {
	splitInput, err := CoalesceQuotes(input)
	if err != nil {
		return nil, err
	}

	command, args := splitInput[0], splitInput[1:]

	args, IOStream, err := ParseRedirection(args)
	if err != nil {
		return nil, err
	}

	return &Command{
		CommandName: command,
		Args:        args,
		IOStream:    IOStream,
	}, nil

}

func (c *Command) Exec(cfg *config) error {
	if handler, ok := GetBuiltInCommands()[c.CommandName]; ok {
		return handler(c, cfg)
	}
	return HandlerExec(c, cfg)
}

func (c *Command) Close() {
	if c.Stdout != os.Stdout {
		c.Stdout.Close()
	}
	if c.Stderr != os.Stderr {
		c.Stderr.Close()
	}
}

func HandlerExec(cmd *Command, cfg *config) error {
	cmdExec := exec.Command(cmd.CommandName, cmd.Args...)
	// Cmd.Err is non-nil when command is not found in PATH
	if cmdExec.Err != nil {
		return fmt.Errorf("%s: command not found", cmd.CommandName)
	}

	cmdExec.Stdin = cmd.Stdin

	stdoutPipe, err := cmdExec.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}
	stderrPipe, err := cmdExec.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}

	
	if err := cmdExec.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	// Handle stdout and stderr conversion in separate goroutines
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go processOutput(wg, stdoutPipe, cmd.Stdout)
	go processOutput(wg, stderrPipe, cmd.Stderr)
	wg.Wait()

	cmdExec.Wait()
	return nil
}

// Raw mode requires replacing "\n" with "\r\n" so output is display correctly
func processOutput(wg *sync.WaitGroup, output io.Reader, writer io.Writer) {
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprint(writer, line+"\r\n")
	}
	wg.Done()
}
