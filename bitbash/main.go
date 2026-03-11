package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"sync"

	"golang.org/x/term"
)

type config struct {
	history             []string
	saved_up_to_index   int
	prev_terminal_state *term.State
	user_name           string
	curr_dir            string
	home_dir            string
	running             *sync.WaitGroup
}

func (cfg *config) restore_terminal() {
	// Restore terminal to previouse state before exiting
	term.Restore(int(os.Stdin.Fd()), cfg.prev_terminal_state)
}

func (cfg *config) load_history() error {
	cfg.history = make([]string, 0)

	path, ok := os.LookupEnv("HISTFILE")
	if !ok {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	s := bufio.NewScanner(file)
	for s.Scan() {
		cfg.history = append(cfg.history, s.Text())
	}

	cfg.saved_up_to_index = len(cfg.history)

	return nil
}

func (cfg *config) shell_prompt() string {
	//if cut, ok := strings.CutPrefix(cfg.currDirectory, cfg.homeDirectory); ok {
	//	cfg.currDirectory = fmt.Sprintf("~%s", cut)
	//}
	//userNameBlueBold := fmt.Sprintf("%s%s%s%s", BLUE, BOLD, cfg.userName, RESET)
	//currDirGreenBold := fmt.Sprintf("%s%s%s%s", GREEN, BOLD, cfg.currDirectory, RESET)
	//return fmt.Sprintf("%s:%s $ ", userNameBlueBold, currDirGreenBold)
	return "$ "
}

func main() {
	u, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	// Switch terminal from cooked to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	cfg := &config{
		prev_terminal_state: oldState,
		user_name:           u.Username,
		curr_dir:            dir,
		home_dir:            home,
		running:             &sync.WaitGroup{},
	}
	defer cfg.restore_terminal()

	if err := cfg.load_history(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load history file: %s\r\n", err)
		return
	}

	if err := run_repl(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s\r\n", err)
		return
	}
}
