package main

import (
	"log"
	"os"

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
