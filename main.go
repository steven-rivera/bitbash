package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"golang.org/x/term"
)

type config struct {
	history          []string
	oldTerminalState *term.State
	userName         string
	currDirectory    string
	homeDirectory    string
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

	u, err := user.Current()
	if err != nil {
		log.Fatalf("shell: %s", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("shell: %s", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("shell: %s", err)
	}

	cfg := &config{
		history:          make([]string, 0),
		oldTerminalState: oldState,
		userName:         u.Username,
		currDirectory:    dir,
		homeDirectory:    home,
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

		command, err := NewCommand(trimmedInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "shell: %s\r\n", err)
			continue
		}

		err = command.Exec(cfg)
		if err != nil {
			fmt.Fprintf(command.Stderr, "%s\r\n", err)
		}

		command.Close()
	}
}
