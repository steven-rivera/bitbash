package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/term"
)

type config struct {
	history           []string
	oldTerminalState *term.State
}

func (cfg *config) CleanUp() {
	// Restore terminal to previouse state before exiting
	term.Restore(int(os.Stdin.Fd()), cfg.oldTerminalState)
}

func main() {
	// Switch terminal from cooked to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("shell: %s", err)
	}
	
	cfg := &config{
		history: make([]string, 0),
		oldTerminalState: oldState,
	}
	defer cfg.CleanUp()

	startREPL(cfg)
}

func startREPL(cfg *config) {
	stdin := bufio.NewReader(os.Stdin)

	for {
		input, err := ReadLine(cfg, stdin)
		if err != nil {
			return
		}

		trimmedInput := strings.TrimSpace(input)
		if len(trimmedInput) == 0 {
			continue
		}

		command, err := NewCommand(cfg, trimmedInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "shell: %s\r\n", err)
			continue
		}

		err = command.Exec()
		if err != nil {
			fmt.Fprintf(command.Stderr, "%s\r\n", err)
		}

		command.Close()
	}
}
