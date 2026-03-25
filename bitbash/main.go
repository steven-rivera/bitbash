package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strings"

	"golang.org/x/term"
)

type Config struct {
	StdinReader           *bufio.Reader
	PreviousTerminalState *term.State
	History               []string
	SavedUpToIndex        int
	UserName              string
	CurrentDirectory      string
	HomeDirectory         string
}

func NewConfig() *Config {
	usr, _ := user.Current()
	dir, _ := os.Getwd()
	home, _ := os.UserHomeDir()
	stdin := bufio.NewReader(os.Stdin)

	cfg := &Config{
		StdinReader:      stdin,
		UserName:         usr.Username,
		CurrentDirectory: dir,
		HomeDirectory:    home,
	}

	cfg.LoadCommandHistory()

	return cfg
}

func (cfg *Config) MakeTerminalRaw() {
	prevState, _ := term.MakeRaw(int(os.Stdin.Fd()))
	cfg.PreviousTerminalState = prevState
}

func (cfg *Config) RestoreTerminal() {
	term.Restore(int(os.Stdin.Fd()), cfg.PreviousTerminalState)
}

func (cfg *Config) ShellPrompt() string {
	if cut, ok := strings.CutPrefix(cfg.CurrentDirectory, cfg.HomeDirectory); ok {
		cfg.CurrentDirectory = fmt.Sprintf("~%s", cut)
	}
	userNameBlueBold := fmt.Sprintf("%s%s%s%s", BLUE, BOLD, cfg.UserName, RESET)
	currDirGreenBold := fmt.Sprintf("%s%s%s%s", GREEN, BOLD, cfg.CurrentDirectory, RESET)
	return fmt.Sprintf("%s:%s $ ", userNameBlueBold, currDirGreenBold)

	//return "$ "
}

func (cfg *Config) LoadCommandHistory() {
	cfg.History = make([]string, 0)

	path, ok := os.LookupEnv("HISTFILE")
	if !ok {
		return
	}

	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	s := bufio.NewScanner(file)
	for s.Scan() {
		cfg.History = append(cfg.History, s.Text())
	}

	cfg.SavedUpToIndex = len(cfg.History)
}

func PrintWelcomeMessage() {
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

	fmt.Print("\r\n", "Welcome to BitBash! Type `help` for a list of builtin commands.", "\r\n\r\n")
}

func RunREPL(cfg *Config) error {
	defer cfg.RestoreTerminal()

	PrintWelcomeMessage()

	for {
		cfg.MakeTerminalRaw()
		fmt.Print(cfg.ShellPrompt())

		input, err := ReadLine(cfg)
		if err != nil {
			return err
		}

		fmt.Print("\r\n")
		cfg.RestoreTerminal()

		input = strings.TrimSpace(input)
		if len(input) == 0 {
			continue
		}

		cfg.History = append(cfg.History, input)

		tokens, err := Tokenize(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "shell: %s\n", err)
			continue
		}

		NewPipeline(tokens).Execute(cfg)
	}
}

func main() {
	cfg := NewConfig()

	if err := RunREPL(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
