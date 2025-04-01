package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type BuiltIn = func(args []string, out *os.File) error

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

		args, outputFile, err := parseRedirection(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		if callback, ok := builtInCommands[command]; ok {
			err := callback(args, outputFile)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		} else {
			err := runProgram(command, args, outputFile)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}

		if outputFile != os.Stdout {
			outputFile.Close()
		}
	}
}

func getBuiltInCommands() map[string]BuiltIn {
	return map[string]BuiltIn{
		"exit": exitCommand,
		"echo": echoCommand,
		"type": typeCommand,
		"pwd":  pwdCommand,
		"cd":   cdCommand,
	}
}

func exitCommand(args []string, out *os.File) error {
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

func echoCommand(args []string, out *os.File) error {
	for _, arg := range args {
		fmt.Fprintf(out, "%s ", strings.Trim(arg, "'\""))
	}
	fmt.Fprintln(out)
	return nil
}

func typeCommand(args []string, out *os.File) error {
	if len(args) != 1 {
		return fmt.Errorf("exit: expected 1 argument got %d", len(args))
	}
	commandArg := args[0]
	if _, ok := getBuiltInCommands()[commandArg]; ok {
		fmt.Fprintf(out, "%s is a shell builtin\n", commandArg)
		return nil
	}

	pathEnv := os.Getenv("PATH")
	for path := range strings.SplitSeq(pathEnv, ":") {
		commandPath := filepath.Join(path, commandArg)
		file, err := os.Stat(commandPath)
		if err == nil && !file.IsDir() {
			fmt.Fprintf(out, "%s is %s\n", commandArg, commandPath)
			return nil
		}
	}

	return fmt.Errorf("%s: not found", commandArg)
}

func runProgram(program string, args []string, out *os.File) error {
	cmd := exec.Command(program, args...)
	if cmd.Err != nil {
		return fmt.Errorf("%s: command not found", program)
	}
	cmd.Stdout = out
	cmd.Stderr = os.Stdin
	cmd.Run()
	return nil
}

func pwdCommand(args []string, out *os.File) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pwd: %w", err)
	}
	fmt.Fprintln(out, workingDir)
	return nil
}

func cdCommand(args []string, out *os.File) error {
	if len(args) != 1 {
		return fmt.Errorf("cd: expected 1 argument got %d", len(args))
	}

	dir := args[0]
	if dir == "~" {
		dir, _ = os.UserHomeDir()
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("cd: %s: No such file or directory", dir)
	}
	return nil
}

func parseRedirection(args []string) (commandArgs []string, outputFile *os.File, err error) {
	for index, arg := range args {
		if arg == ">" || arg == "1>" {
			commandArgs, filePathSlice := args[:index], args[index+1:]
			if len(filePathSlice) != 1 {
				return nil, nil, fmt.Errorf("%s: expected 1 argument got %d", arg, len(filePathSlice))
			}
			file, err := os.Create(filePathSlice[0])
			if err != nil {
				return nil, nil, fmt.Errorf("%s: %w", arg, err)
			}
			return commandArgs, file, nil
		}
	}
	return args, os.Stdout, nil
}
