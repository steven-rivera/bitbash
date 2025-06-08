package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"
	"sync"

	"golang.org/x/term"
)

type config struct {
	history          []string
	savedUpToIndex   int
	oldTerminalState *term.State
	userName         string
	currDirectory    string
	homeDirectory    string
	running          *sync.WaitGroup
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
		running:          &sync.WaitGroup{},
	}
	defer cfg.CleanUp()

	printWelcomeMessage()
	startREPL(cfg)
}

func startREPL(cfg *config) {
	stdin := bufio.NewReader(os.Stdin)

	for {
		input, err := ReadLine(cfg, stdin)
		if err != nil {
			fmt.Print("\r\n")
			return
		}

		trimmedInput := strings.TrimSpace(input)
		if len(trimmedInput) == 0 {
			continue
		}
		cfg.history = append(cfg.history, trimmedInput)

		command, err := NewCommand(trimmedInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "shell: %s\r\n", err)
			continue
		}

		command.Exec(cfg)
	}
}

func printWelcomeMessage() {
	fmt.Print(GREEN)
	fmt.Print(`________   ___   _________   ________   ________   ________   ___  ___      `, "\r\n")
	fmt.Print(`|\   __  \ |\  \ |\___   ___\|\   __  \ |\   __  \ |\   ____\ |\  \|\  \    `, "\r\n")
	fmt.Print(`\ \  \|\ /_\ \  \\|___ \  \_|\ \  \|\ /_\ \  \|\  \\ \  \___|_\ \  \\\  \   `, "\r\n")
	fmt.Print(` \ \   __  \\ \  \    \ \  \  \ \   __  \\ \   __  \\ \_____  \\ \   __  \  `, "\r\n")
	fmt.Print(`  \ \  \|\  \\ \  \    \ \  \  \ \  \|\  \\ \  \ \  \\|____|\  \\ \  \ \  \ `, "\r\n")
	fmt.Print(`   \ \_______\\ \__\    \ \__\  \ \_______\\ \__\ \__\ ____\_\  \\ \__\ \__\`, "\r\n")
	fmt.Print(`    \|_______| \|__|     \|__|   \|_______| \|__|\|__||\_________\\|__|\|__|`, "\r\n")
	fmt.Print(`                                                      \|_________|          `, "\r\n")
	fmt.Print(RESET)

	fmt.Print("\r\nWelcome to BitBash! Here is a list of supported features:\r\n\r\n")

	fmt.Print("Redirection:\r\n")
	fmt.Print("    <            Redirect stdin from file\r\n")
	fmt.Print("    >  >>        Redirect stdout to file\r\n")
	fmt.Print("    2>  2>>      Redirect stderr to file\r\n")
	fmt.Print("    &>  &>>      Redirect botth stdin and stderr to file\r\n")
	fmt.Print("\r\n")
	fmt.Print("Piping:\r\n")
	fmt.Print("    |            Redirect the stdout of one command to the stdin of another\r\n")
	fmt.Print("\r\n")
	fmt.Print("Autocomplete:\r\n")
	fmt.Print("    TAB          Attempt to complete or partially complete the name of the commnd\r\n")
	fmt.Print("    TAB TAB      If multiple autocomplete matches print them all\r\n")
	fmt.Print("\r\n")
	fmt.Print("Quoting:\r\n")
	fmt.Print("    '...'            Characters quoted in single quotes preserve their literal value\r\n")
	fmt.Print(`    "..."            Same as single quotes but processes the escape sequences \\, \$, and \"`, "\r\n")
	fmt.Print("Command History:\r\n")
	fmt.Print("\r\n")
	fmt.Print("    Up Arrow     Replace current line with previous command in history\r\n")
	fmt.Print("    Down Arrow   Replace current line with next command in history\r\n")

	fmt.Print("\r\nType help for a list of builtin commands.\r\n")
}
