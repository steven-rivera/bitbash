package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type CommandHandler = func(cmd *Command) error

func GetBuiltInCommands() map[string]CommandHandler {
	return map[string]CommandHandler{
		"exit": HandlerExit,
		"echo": HandlerEcho,
		"type": HandlerType,
		"pwd":  HandlerPwd,
		"cd":   HandlerCd,
	}
}

func HandlerCd(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("cd: expected 1 argument got %d", len(cmd.Args))
	}
	dir := cmd.Args[0]
	if dir == "~" {
		dir, _ = os.UserHomeDir()
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("cd: %s: No such file or directory", dir)
	}
	return nil
}

func HandlerEcho(cmd *Command) error {
	for _, arg := range cmd.Args {
		fmt.Fprintf(cmd.Stdout, "%s ", arg)
	}
	fmt.Fprint(cmd.Stdout, "\r\n")
	return nil
}

func HandlerExit(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("exit: expected 1 argument got %d", len(cmd.Args))
	}

	exitCode, err := strconv.Atoi(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("exit: invalid exit code '%s'\n", cmd.Args[0])
	}
	// Must clean up because os.Exit doesn't run defered function
	cmd.Cfg.CleanUp()
	os.Exit(exitCode)
	return nil
}

func HandlerPwd(cmd *Command) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pwd: %w", err)
	}
	fmt.Fprintf(cmd.Stdout, "%s\r\n", workingDir)
	return nil
}

func HandlerType(cmd *Command) error {
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
