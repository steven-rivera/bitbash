package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	stdin := bufio.NewReader(os.Stdin)
	commands := getBuitInCommands()

	for {
		fmt.Print("$ ")
		command, err := stdin.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		command = strings.TrimSuffix(command, "\n")

		if strings.HasPrefix(command, "exit ") {
			split := strings.SplitAfterN(command, " ", 2)
			exitCode, err := strconv.Atoi(split[1])
			if err != nil {
				fmt.Printf("exit: invalid exit code '%s'\n", split[1])
			}
			os.Exit(exitCode)
		}

		if strings.HasPrefix(command, "echo ") {
			split := strings.SplitAfterN(command, " ", 2)
			fmt.Println(split[1])
			continue
		}

		if strings.HasPrefix(command, "type ") {
			split := strings.SplitAfterN(command, " ", 2)
			if _, ok := commands[split[1]]; ok {
				fmt.Printf("%s is a shell builtin\n", split[1])
			} else {
				fmt.Printf("%s: not found\n", split[1])
			}
			continue
		}

		fmt.Println(command + ": command not found")
	}
}

func getBuitInCommands() map[string]bool {
	return map[string]bool{
		"exit": true,
		"echo": true,
		"type": true,
	}
}