package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type BuiltIn = func(args []string) error

func main() {
	stdin := bufio.NewReader(os.Stdin)
	builtInCommands := getBuiltInCommands()

	for {
		fmt.Print("$ ")
		input, err := stdin.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		input = strings.TrimSuffix(input, "\n")
		splitInput := strings.Split(input, " ")
		command, args := splitInput[0], splitInput[1:]

		if callback, ok := builtInCommands[command]; ok {
			err := callback(args)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Printf("%s: not found\n", command)
		}
	}
}

func getBuiltInCommands() map[string]BuiltIn {
	return map[string]BuiltIn{
		"exit": exitCommand,
		"echo": echoCommand,
		"type": typeCommand,
	}
}

func exitCommand(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("exit: expected 1 argument got %d", len(args))
	}

	exitCode, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("exit: invalid exit code '%s'\n", args[0])
	}
	os.Exit(exitCode)
	return nil
}

func echoCommand(args []string) error {
	for _, arg := range args {
		fmt.Printf("%s ", arg)
	}
	fmt.Println()
	return nil
}

func typeCommand(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("exit: expected 1 argument got %d", len(args))
	}
	commandArg := args[0]
	if _, ok := getBuiltInCommands()[commandArg]; ok {
		fmt.Printf("%s is a shell builtin\n", commandArg)
		return nil
	}

	pathEnv := os.Getenv("PATH")
	for path := range strings.SplitSeq(pathEnv, ":") {
		commandPath := filepath.Join(path, commandArg)
		file, err := os.Stat(commandPath)
		if err == nil && !file.IsDir() {
			fmt.Printf("%s is %s\n", commandArg, commandPath)
			return nil
		}
	}
	fmt.Printf("%s: not found\n", commandArg)
	return nil
}
