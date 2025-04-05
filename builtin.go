package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type BuiltInCommand struct {
	Name        string
	Usage       string
	Description string
	Handler     func(cmd *Command, cfg *config) error
}

func GetBuiltInCommands() map[string]BuiltInCommand {
	return map[string]BuiltInCommand{
		"exit": BuiltInCommand{
			Name:        "exit",
			Usage:       "exit <code>",
			Description: "cause the shell to exit with provided code",
			Handler:     HandlerExit,
		},
		"echo": BuiltInCommand{
			Name:        "echo",
			Usage:       "echo [arg...]",
			Description: "print all arguments separated by a space to stdout",
			Handler:     HandlerEcho,
		},
		"type": BuiltInCommand{
			Name:        "type",
			Usage:       "type <command>",
			Description: "print whether command is builtin, if not print location of executatable",
			Handler:     HandlerType,
		},
		"pwd": BuiltInCommand{
			Name:        "pwd",
			Usage:       "pwd",
			Description: "print the current working directory",
			Handler:     HandlerPwd,
		},
		"cd": BuiltInCommand{
			Name:        "cd",
			Usage:       "cd <directory>",
			Description: "change the current working directory to the provided directory",
			Handler:     HandlerCd,
		},
		"help": BuiltInCommand{
			Name:        "help",
			Usage:       "help",
			Description: "print this help message",
			Handler:     HandlerHelp,
		},
	}
}

func HandlerCd(cmd *Command, cfg *config) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("cd: expected 1 argument got %d", len(cmd.Args))
	}
	dir := cmd.Args[0]
	if dir == "~" {
		dir = cfg.homeDirectory
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("cd: %s: No such file or directory", dir)
	}
	cfg.currDirectory, _ = os.Getwd()
	return nil
}

func HandlerEcho(cmd *Command, cfg *config) error {
	for _, arg := range cmd.Args {
		fmt.Fprintf(cmd.Stdout, "%s ", arg)
	}
	fmt.Fprint(cmd.Stdout, "\r\n")
	return nil
}

func HandlerExit(cmd *Command, cfg *config) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("exit: expected 1 argument got %d", len(cmd.Args))
	}

	exitCode, err := strconv.Atoi(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("exit: invalid exit code '%s'\n", cmd.Args[0])
	}
	// Must clean up because os.Exit doesn't run defered functions
	cfg.CleanUp()
	os.Exit(exitCode)
	return nil
}

func HandlerPwd(cmd *Command, cfg *config) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pwd: %w", err)
	}
	fmt.Fprintf(cmd.Stdout, "%s\r\n", workingDir)
	return nil
}

func HandlerType(cmd *Command, cfg *config) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("exit: expected 1 argument got %d", len(cmd.Args))
	}
	commandArg := cmd.Args[0]
	if _, ok := GetBuiltInCommands()[commandArg]; ok {
		fmt.Fprintf(cmd.Stdout, "%s is a shell builtin\r\n", commandArg)
		return nil
	}

	pathEnv := os.Getenv("PATH")
	for dir := range strings.SplitSeq(pathEnv, ":") {
		dirEntries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, dirEntry := range dirEntries {
			if !dirEntry.IsDir() && dirEntry.Name() == commandArg {
				fmt.Fprintf(cmd.Stdout, "%s is %s\r\n", commandArg, filepath.Join(dir, commandArg))
				return nil
			}
		}
	}

	return fmt.Errorf("%s: not found", commandArg)
}

func HandlerHelp(cmd *Command, cfg *config) error {
	fmt.Fprint(cmd.Stdout, "These BitBash commands are defined internally\r\n\r\n")
	fmt.Fprint(cmd.Stdout, "Commands:\r\n")
	for _, builtin := range GetBuiltInCommands() {
		fmt.Fprintf(cmd.Stdout, "    %s\r\n", builtin.Usage)
		fmt.Fprintf(cmd.Stdout, "      -%s\r\n\r\n", builtin.Description)
	}
	return nil
}
