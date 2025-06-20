package main

import (
	"bufio"
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
		"history": BuiltInCommand{
			Name:        "history",
			Usage:       "history [<n> | (-r|-w|-a) <path_to_history_file>]",
			Description: "list previously executed commands or load/write/append commands from/to a file",
			Handler:     HandlerHistory,
		},
	}
}

func HandlerCd(cmd *Command, cfg *config) {
	defer cfg.running.Done()
	defer cmd.Close()
	if len(cmd.Args) != 1 {
		fmt.Fprintf(cmd.Stderr, "cd: expected 1 argument got %d\r\n", len(cmd.Args))
		return
	}
	dir := cmd.Args[0]
	if dir == "~" {
		dir = cfg.homeDirectory
	}

	if err := os.Chdir(dir); err != nil {
		fmt.Fprintf(cmd.Stderr, "cd: %s: No such file or directory\r\n", dir)
		return
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
		return
	}

	exitCode, err := strconv.Atoi(cmd.Args[0])
	if err != nil {
		fmt.Fprintf(cmd.Stderr, "exit: invalid exit code '%s'\r\n", cmd.Args[0])
		return
	}

	path, ok := os.LookupEnv("HISTFILE")
	if ok {
		// Save history before exiting if HISTFILE environment variables is specified
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
		if err == nil {
			for i := cfg.savedUpToIndex; i < len(cfg.history); i++ {
				file.Write(fmt.Appendf(nil, "%s\n", cfg.history[i]))
			}
			file.Close()
		}
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
		return
	}
	fmt.Fprintf(cmd.Stdout, "%s\r\n", workingDir)
}

func HandlerType(cmd *Command, cfg *config) {
	defer cfg.running.Done()
	defer cmd.Close()
	if len(cmd.Args) != 1 {
		fmt.Fprintf(cmd.Stderr, "exit: expected 1 argument got %d\r\n", len(cmd.Args))
		return
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

func HandlerHistory(cmd *Command, cfg *config) {
	defer cfg.running.Done()
	defer cmd.Close()

	switch len(cmd.Args) {
	case 0:
		// No arguments, print entire history
		for i, entry := range cfg.history {
			fmt.Fprintf(cmd.Stdout, "%d %s\r\n", i+1, entry)
		}
	case 1:
		// One argument, print last n entries
		size, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			fmt.Fprintf(cmd.Stderr, "history: %s: numeric argument required\r\n", cmd.Args[0])
		}

		historySize := len(cfg.history)
		size = min(size, historySize)

		for i := historySize - size; i < historySize; i++ {
			fmt.Fprintf(cmd.Stdout, "%d %s\r\n", i+1, cfg.history[i])
		}
	case 2:
		// Two arguments, load history from file
		if cmd.Args[0] == "-r" {
			historyFile, err := os.Open(cmd.Args[1])
			if err != nil {
				fmt.Fprintf(cmd.Stderr, "history: could not open history file: %s\r\n", err)
			}
			defer historyFile.Close()

			history := bufio.NewScanner(historyFile)
			for history.Scan() {
				cfg.history = append(cfg.history, history.Text())
			}
			return
		}

		// Two arguments, write history to file
		if cmd.Args[0] == "-w" {
			historyFile, err := os.Create(cmd.Args[1])
			if err != nil {
				fmt.Fprintf(cmd.Stderr, "history: could not create history file: %s\r\n", err)
			}
			defer historyFile.Close()

			for _, entry := range cfg.history {
				historyFile.Write(fmt.Appendf(nil, "%s\n", entry))
			}

			return
		}

		// Two arguments, append history to file
		if cmd.Args[0] == "-a" {
			historyFile, err := os.OpenFile(cmd.Args[1], os.O_WRONLY|os.O_APPEND, 0o666)
			if err != nil {
				fmt.Fprintf(cmd.Stderr, "history: could not open history file: %s\r\n", err)
			}
			defer historyFile.Close()

			for i := cfg.savedUpToIndex; i < len(cfg.history); i++ {
				historyFile.Write(fmt.Appendf(nil, "%s\n", cfg.history[i]))
			}
			cfg.savedUpToIndex = len(cfg.history)

			return
		}

		fmt.Fprintf(cmd.Stderr, "history: %s: invalid argument\r\n", cmd.Args[0])
	default:
		fmt.Fprint(cmd.Stderr, "history: too many arguments\r\n")
	}
}
