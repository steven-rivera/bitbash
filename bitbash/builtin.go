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
	Handler     func(cmd *Command, cfg *Config)
}

func HandlerCd(cmd *Command, cfg *Config) {
	if len(cmd.Args) != 1 {
		fmt.Fprintf(cmd.err, "cd: expected 1 argument got %d\n", len(cmd.Args))
		return
	}

	dir := cmd.Args[0]

	if after, ok := strings.CutPrefix(dir, "~"); ok {
		dir = cfg.HomeDirectory + after
	}

	if err := os.Chdir(dir); err != nil {
		fmt.Fprintf(cmd.err, "cd: %s: No such file or directory\n", dir)
		return
	}

	cfg.CurrentDirectory, _ = os.Getwd()
}

func HandlerEcho(cmd *Command, cfg *Config) {
	fmt.Fprint(cmd.out, strings.Join(cmd.Args, " "), "\n")
}

func HandlerExit(cmd *Command, cfg *Config) {
	exitCode := 0

	if len(cmd.Args) > 1 {
		fmt.Fprintf(cmd.err, "exit: expected 1 argument got %d\n", len(cmd.Args))
		return
	}

	if len(cmd.Args) == 1 {
		num, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			fmt.Fprintf(cmd.err, "exit: invalid exit code '%s'\n", cmd.Args[0])
			return
		}
		exitCode = num
	}

	// Save history before exiting if HISTFILE env variable is defined

	if path, ok := os.LookupEnv("HISTFILE"); ok {
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
		if err == nil {
			for i := cfg.SavedUpToIndex; i < len(cfg.History); i++ {
				file.Write(fmt.Appendf(nil, "%s\n", cfg.History[i]))
			}
			file.Close()
		}
	}

	os.Exit(exitCode)
}

func HandlerPwd(cmd *Command, cfg *Config) {
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(cmd.err, "pwd: %s\n", err)
		return
	}
	fmt.Fprintf(cmd.out, "%s\n", workingDir)
}

func HandlerType(cmd *Command, cfg *Config) {
	if len(cmd.Args) != 1 {
		fmt.Fprintf(cmd.err, "exit: expected 1 argument got %d\r", len(cmd.Args))
		return
	}

	command := cmd.Args[0]
	if _, ok := BUILTIN_CMDS[command]; ok {
		fmt.Fprintf(cmd.out, "%s is a shell builtin\n", command)
		return
	}

	pathEnv := os.Getenv("PATH")
	for dir := range strings.SplitSeq(pathEnv, ":") {
		dirEntries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, dirEntry := range dirEntries {
			if dirEntry.IsDir() || dirEntry.Name() != command {
				continue
			}

			info, err := dirEntry.Info()
			if err != nil {
				continue
			}

			if isExecutable := info.Mode()&0111 != 0; isExecutable {
				fmt.Fprintf(cmd.out, "%s is %s\n", command, filepath.Join(dir, command))
				return
			}
		}
	}

	fmt.Fprintf(cmd.err, "%s: not found\n", command)
}

func HandlerHelp(cmd *Command, cfg *Config) {
	fmt.Fprint(cmd.out, "These BitBash commands are defined internally\n\n")
	fmt.Fprint(cmd.out, "Commands:\n")
	for _, builtin := range BUILTIN_CMDS {
		fmt.Fprintf(cmd.out, "    %s\n", builtin.Usage)
		fmt.Fprintf(cmd.out, "      -%s\n\n", builtin.Description)
	}
}

func HandlerHistory(cmd *Command, cfg *Config) {
	switch len(cmd.Args) {
	// No arguments, print entire history
	case 0:
		for i, entry := range cfg.History {
			fmt.Fprintf(cmd.out, "%d %s\r\n", i+1, entry)
		}
	// One argument, print last n entries
	case 1:
		size, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			fmt.Fprintf(cmd.err, "history: %s: numeric argument required\n", cmd.Args[0])
		}

		historySize := len(cfg.History)
		size = min(size, historySize)

		for i := historySize - size; i < historySize; i++ {
			fmt.Fprintf(cmd.out, "%d %s\n", i+1, cfg.History[i])
		}
	case 2:
		// Load history from file
		if cmd.Args[0] == "-r" {
			historyFile, err := os.Open(cmd.Args[1])
			if err != nil {
				fmt.Fprintf(cmd.err, "history: could not open history file: %s\r\n", err)
			}
			defer historyFile.Close()

			history := bufio.NewScanner(historyFile)
			for history.Scan() {
				cfg.History = append(cfg.History, history.Text())
			}
			return
		}

		// Write history to file
		if cmd.Args[0] == "-w" {
			historyFile, err := os.Create(cmd.Args[1])
			if err != nil {
				fmt.Fprintf(cmd.err, "history: could not create history file: %s\r\n", err)
			}
			defer historyFile.Close()

			for _, entry := range cfg.History {
				historyFile.Write(fmt.Appendf(nil, "%s\n", entry))
			}

			return
		}

		// Append history to file
		if cmd.Args[0] == "-a" {
			historyFile, err := os.OpenFile(cmd.Args[1], os.O_WRONLY|os.O_APPEND, 0o666)
			if err != nil {
				fmt.Fprintf(cmd.err, "history: could not open history file: %s\r\n", err)
			}
			defer historyFile.Close()

			for i := cfg.SavedUpToIndex; i < len(cfg.History); i++ {
				historyFile.Write(fmt.Appendf(nil, "%s\n", cfg.History[i]))
			}
			cfg.SavedUpToIndex = len(cfg.History)

			return
		}

		fmt.Fprintf(cmd.err, "history: %s: invalid argument\n", cmd.Args[0])
	default:
		fmt.Fprint(cmd.err, "history: too many arguments\n")
	}
}

func init() {
	BUILTIN_CMDS = make(map[string]BuiltInCommand)

	BUILTIN_CMDS["exit"] = BuiltInCommand{
		Name:        "exit",
		Usage:       "exit [<code>]",
		Description: "cause the shell to exit with provided code",
		Handler:     HandlerExit,
	}

	BUILTIN_CMDS["echo"] = BuiltInCommand{
		Name:        "echo",
		Usage:       "echo [arg...]",
		Description: "print all arguments separated by a space to stdout",
		Handler:     HandlerEcho,
	}

	BUILTIN_CMDS["type"] = BuiltInCommand{
		Name:        "type",
		Usage:       "type <command>",
		Description: "print whether command is builtin, if not print location of executatable",
		Handler:     HandlerType,
	}

	BUILTIN_CMDS["pwd"] = BuiltInCommand{
		Name:        "pwd",
		Usage:       "pwd",
		Description: "print the current working directory",
		Handler:     HandlerPwd,
	}

	BUILTIN_CMDS["cd"] = BuiltInCommand{
		Name:        "cd",
		Usage:       "cd <directory>",
		Description: "change the current working directory to the provided directory",
		Handler:     HandlerCd,
	}

	BUILTIN_CMDS["help"] = BuiltInCommand{
		Name:        "help",
		Usage:       "help",
		Description: "print this help message",
		Handler:     HandlerHelp,
	}

	BUILTIN_CMDS["history"] = BuiltInCommand{
		Name:        "history",
		Usage:       "history [<n> | (-r|-w|-a) <path_to_history_file>]",
		Description: "list previously executed commands or load/write/append commands from/to a file",
		Handler:     HandlerHistory,
	}
}
