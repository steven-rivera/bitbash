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
	Name string
	Args []string
	IOStream
	PipedInto *Command
}

func NewCommand(input string) (*Command, error) {
	splitInput, err := SplitInput(input)
	if err != nil {
		return nil, err
	}

	command, err := ParsePipes(splitInput)
	if err != nil {
		return nil, err
	}

	currCmd := command
	for currCmd != nil {
		args, IOStream, err := ParseRedirection(currCmd.Args)
		if err != nil {
			return nil, err
		}
		currCmd.Args = args
		currCmd.IOStream = IOStream
		currCmd = currCmd.PipedInto
	}

	currCmd = command
	for currCmd.PipedInto != nil {
		r, w, err := os.Pipe()
		if err != nil {
			return nil, err
		}

		currCmd.Stdout = w
		currCmd.PipedInto.Stdin = r
		currCmd = currCmd.PipedInto
	}

	// currCmd = command
	// for currCmd != nil {
	// 	fmt.Printf("Name:   %#v\r\n", currCmd.Name)
	// 	fmt.Printf("Args:   %#v\r\n", currCmd.Args)
	// 	fmt.Printf("Stdin:  %#v\r\n", currCmd.Stdin)
	// 	fmt.Printf("Stdout: %#v\r\n", currCmd.Stdout)
	// 	fmt.Printf("Stderr: %#v\r\n\r\n", currCmd.Stderr)
	// 	currCmd = currCmd.PipedInto
	// }

	return command, nil

}

func (c *Command) Exec(cfg *config) {
	curr := c
	for curr != nil {
		cfg.running.Add(1)
		if builtin, ok := GetBuiltInCommands()[curr.Name]; ok {
			go builtin.Handler(curr, cfg)
		} else {
			go HandlerExec(curr, cfg)
		}
		curr = curr.PipedInto
	}
	cfg.running.Wait()
}

// Close pipes and/or redirection files after execution
func (c *Command) Close() {
	if c.Stdin != os.Stdin {
		c.Stdin.Close()
	}
	if c.Stdout != os.Stdout {
		c.Stdout.Close()
	}
	if c.Stderr != os.Stderr {
		c.Stderr.Close()
	}
}

func HandlerExec(cmd *Command, cfg *config) {
	defer cfg.running.Done()
	defer cmd.Close()

	cmdExec := exec.Command(cmd.Name, cmd.Args...)
	// Cmd.Err is non-nil when command is not found in PATH
	if cmdExec.Err != nil {
		fmt.Fprintf(os.Stderr, "%s: command not found\r\n", cmd.Name)
		return
	}

	cmdExec.Stdin = cmd.Stdin

	stdoutPipe, err := cmdExec.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "shell: error creating stdout pipe: %s", err)
		return
	}
	stderrPipe, err := cmdExec.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "shell: error creating stderr pipe: %s", err)
		return
	}

	if err := cmdExec.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: error starting command: %s\r\n", cmd.Name, err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	// Handle stdout and stderr conversion in separate goroutines
	go interceptOutput(wg, stdoutPipe, cmd.Stdout)
	go interceptOutput(wg, stderrPipe, cmd.Stderr)
	// Finish processing stdout and stderr before calling Cmd.Wait()
	wg.Wait()

	cmdExec.Wait()
}

// Raw mode requires replacing "\n" with "\r\n" so output is displayed correctly
func interceptOutput(wg *sync.WaitGroup, output io.Reader, writer io.Writer) {
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprint(writer, line+"\r\n")
	}
	wg.Done()
}
