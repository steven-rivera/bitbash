package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strings"
	"sync"

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
		UserName:         usr.Name,
		CurrentDirectory: dir,
		HomeDirectory:    home,
	}

	cfg.loadCommandHistory()

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
	//if cut, ok := strings.CutPrefix(cfg.currDirectory, cfg.homeDirectory); ok {
	//	cfg.currDirectory = fmt.Sprintf("~%s", cut)
	//}
	//userNameBlueBold := fmt.Sprintf("%s%s%s%s", BLUE, BOLD, cfg.userName, RESET)
	//currDirGreenBold := fmt.Sprintf("%s%s%s%s", GREEN, BOLD, cfg.currDirectory, RESET)
	//return fmt.Sprintf("%s:%s $ ", userNameBlueBold, currDirGreenBold)

	return "$ "
}

func (cfg *Config) loadCommandHistory() {
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

	fmt.Print("\r\n", "Welcome to BitBash! Type `help` for a list of builtin commands.", "\r\n\r\n")
}

func Execute(cfg *Config, input string) {
	tokens, err := Tokenize(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shell: %s\n", err)
		return
	}

	pipeLine, err := CreatePipeline(tokens)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shell: %s\n", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(pipeLine))

	for _, cmd := range pipeLine {
		go cmd.Run(&wg, cfg)
	}

	wg.Wait()
}

func RunREPL(cfg *Config) error {
	defer cfg.RestoreTerminal()

	//printWelcomeMessage()

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

		Execute(cfg, input)
	}
}

func main() {
	cfg := NewConfig()

	if err := RunREPL(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
