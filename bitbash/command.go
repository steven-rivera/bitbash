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

func NewCommand(rawInput string) (*Command, error) {
	splitInput, err := SplitInput(rawInput)
	if err != nil {
		return nil, err
	}

	command, err := ParsePipes(splitInput)
	if err != nil {
		return nil, err
	}

	err = ParseRedirection(command)
	if err != nil {
		return nil, err
	}

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

// Raw mode requires replacing "\n" with "\r\n" when outputting to terminal.
// If not text will be printed on a new line but at column of the previous
// line instead of column 0.
func interceptOutput(wg *sync.WaitGroup, output io.Reader, writer io.Writer) {
	reader := bufio.NewReader(output)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}
		if b == '\n' && (writer == os.Stdout || writer == os.Stderr) {
			fmt.Fprint(writer, "\r\n")
		} else {
			fmt.Fprint(writer, string(b))
		}
	}
	wg.Done()
}
