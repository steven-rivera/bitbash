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
	Handler     func(cmd *Command, cfg *config)
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

func HandlerCd(cmd *Command, cfg *config) {
	defer cfg.running.Done()
	defer cmd.Close()
	if len(cmd.Args) != 1 {
		fmt.Fprintf(cmd.Stderr, "cd: expected 1 argument got %d\r\n", len(cmd.Args))
	}
	dir := cmd.Args[0]
	if dir == "~" {
		dir = cfg.homeDirectory
	}

	if err := os.Chdir(dir); err != nil {
		fmt.Fprintf(cmd.Stderr, "cd: %s: No such file or directory\r\n", dir)
	}
	cfg.currDirectory, _ = os.Getwd()
}

func HandlerEcho(cmd *Command, cfg *config) {
	defer cfg.running.Done()
	defer cmd.Close()

	fmt.Fprint(cmd.Stdout, strings.Join(cmd.Args, " "))
	if cmd.Stdout == os.Stdout {
		fmt.Fprint(cmd.Stdout, "\r\n")
	} else {
		fmt.Fprint(cmd.Stdout, "\n")
	}
}

func HandlerExit(cmd *Command, cfg *config) {
	defer cfg.running.Done()
	defer cmd.Close()
	if len(cmd.Args) != 1 {
		fmt.Fprintf(cmd.Stderr, "exit: expected 1 argument got %d\r\n", len(cmd.Args))
	}

	exitCode, err := strconv.Atoi(cmd.Args[0])
	if err != nil {
		fmt.Fprintf(cmd.Stderr, "exit: invalid exit code '%s'\r\n", cmd.Args[0])
	}

	// Call because os.Exit won't run defered call
	cfg.CleanUp()
	os.Exit(exitCode)
}

func HandlerPwd(cmd *Command, cfg *config) {
	defer cfg.running.Done()
	defer cmd.Close()
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(cmd.Stderr, "pwd: %s\r\n", err)
	}
	fmt.Fprintf(cmd.Stdout, "%s\r\n", workingDir)
}

func HandlerType(cmd *Command, cfg *config) {
	defer cfg.running.Done()
	defer cmd.Close()
	if len(cmd.Args) != 1 {
		fmt.Fprintf(cmd.Stderr, "exit: expected 1 argument got %d\r\n", len(cmd.Args))
	}
	commandArg := cmd.Args[0]
	if _, ok := GetBuiltInCommands()[commandArg]; ok {
		fmt.Fprintf(cmd.Stdout, "%s is a shell builtin\r\n", commandArg)
		return
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
				return
			}
		}
	}

	fmt.Fprintf(cmd.Stderr, "%s: not found\r\n", commandArg)
}

func HandlerHelp(cmd *Command, cfg *config) {
	defer cfg.running.Done()
	defer cmd.Close()
	fmt.Fprint(cmd.Stdout, "These BitBash commands are defined internally\r\n\r\n")
	fmt.Fprint(cmd.Stdout, "Commands:\r\n")
	for _, builtin := range GetBuiltInCommands() {
		fmt.Fprintf(cmd.Stdout, "    %s\r\n", builtin.Usage)
		fmt.Fprintf(cmd.Stdout, "      -%s\r\n\r\n", builtin.Description)
	}
}
