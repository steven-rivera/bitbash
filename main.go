package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/term"
)

func main() {
	// Switch terminal from cooked to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("shell: %s", err)
	}
	// Restore terminal to previouse state before exiting
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	startREPL()
}

func startREPL() {
	stdin := bufio.NewReader(os.Stdin)

	for {
		input, err := ReadLine(stdin)
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

		err = command.Exec()
		if err != nil {
			fmt.Fprintf(command.Stderr, "%s\r\n", err)
		}

		command.Close()
	}
}
