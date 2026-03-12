package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

const (
	EOT = byte(4)   // Ctrl+D
	DEL = byte(127) // Backspace
	ESC = byte(27)  // ESC

	ARROW_UP    = "[A"
	ARROW_DOWN  = "[B"
	ARROW_RIGHT = "[C"
	ARROW_LEFT  = "[D"
)

type line_state struct {
	line                   []byte
	prev_line              []byte
	cursor_idx             int
	history_idx            int
	next_tab_auto_complete bool
}

func (ls *line_state) insert_byte(b byte) {
	ls.line = append(ls.line, 0)
	copy(ls.line[ls.cursor_idx+1:], ls.line[ls.cursor_idx:])
	ls.line[ls.cursor_idx] = b

	ls.cursor_idx++

	fmt.Printf("%c%s", b, ls.line[ls.cursor_idx:])
	move_cursor_left(len(ls.line[ls.cursor_idx:]))
}

func (ls *line_state) delete_byte() {
	if ls.cursor_idx == 0 {
		fmt.Print("\a")
		return
	}

	if ls.cursor_idx == len(ls.line) {
		ls.line = ls.line[:len(ls.line)-1]
	} else {
		copy(ls.line[ls.cursor_idx-1:], ls.line[ls.cursor_idx:])
		ls.line = ls.line[:len(ls.line)-1]
	}

	ls.cursor_idx--

	move_cursor_left(1)
	fmt.Printf("\x1b[K%s", ls.line[ls.cursor_idx:])
	move_cursor_left(len(ls.line[ls.cursor_idx:]))
}

func (ls *line_state) clear_line() {
	move_cursor_left(ls.cursor_idx)
	fmt.Printf("\x1b[K")
}

func (ls *line_state) set_line(line string) {
	ls.clear_line()

	ls.line = []byte(line)
	ls.cursor_idx = len(ls.line)
	fmt.Printf("%s", ls.line)
}

func (ls *line_state) handle_arrow_key(cfg *config, seq string) {
	switch seq {
	case ARROW_UP:
		ls.handle_arrow_key_up(cfg)
	case ARROW_DOWN:
		ls.handle_arrow_key_down(cfg)
	case ARROW_RIGHT:
		ls.handle_arrow_key_right()
	case ARROW_LEFT:
		ls.handle_arrow_key_left()
	}
}

func (ls *line_state) handle_arrow_key_up(cfg *config) {
	if ls.history_idx == -1 && len(cfg.history) != 0 {
		ls.history_idx = len(cfg.history)
		ls.prev_line = ls.line
	}

	if ls.history_idx > 0 {
		ls.history_idx--
		ls.set_line(cfg.history[ls.history_idx])
	} else {
		fmt.Print("\a")
	}
}

func (ls *line_state) handle_arrow_key_down(cfg *config) {
	if ls.history_idx == -1 {
		fmt.Print("\a")
		return
	}

	if ls.history_idx < len(cfg.history)-1 {
		ls.history_idx++
		ls.set_line(cfg.history[ls.history_idx])

	} else {
		ls.history_idx = -1
		ls.set_line(string(ls.prev_line))
	}
}

func (ls *line_state) handle_arrow_key_right() {
	if ls.cursor_idx == len(ls.line) {
		fmt.Print("\a")
		return
	}

	ls.cursor_idx++
	move_cursor_right(1)
}

func (ls *line_state) handle_arrow_key_left() {
	if ls.cursor_idx == 0 {
		fmt.Print("\a")
		return
	}

	ls.cursor_idx--
	move_cursor_left(1)
}

func (ls *line_state) tab_completion(cfg *config) {
	var matches []string

	start, is_cmd := ls.get_prefix_start()
	prefix := string(ls.line[start:ls.cursor_idx])
	path, name := split_prefix(prefix)

	if is_cmd {
		matches = auto_complete_command(path, name)
	} else {
		matches = auto_complete_file_name(path, name)
	}

	if len(matches) == 0 {
		fmt.Print("\a")
		return
	}

	if len(matches) == 1 {
		before, after := ls.line[:start], ls.line[ls.cursor_idx:]
		matched := fmt.Sprintf("%s%s%s %s", before, path, matches[0], after)
		ls.set_line(matched)
		ls.cursor_idx -= len(after)
		move_cursor_left(len(after))
		return
	}

	fmt.Print("\a")

	

	// check for partial completions
	if lcp := longest_common_prefix(matches); lcp != name {
		before, after := ls.line[:start], ls.line[ls.cursor_idx:]
		partial_match := fmt.Sprintf("%s%s%s%s", before, path, lcp, after)
		ls.set_line(partial_match)
		ls.cursor_idx -= len(after)
		move_cursor_left(len(after))
		ls.next_tab_auto_complete = false
		return
	}

	// if TAB pressed twice in sequence, print all matches on new line
	if ls.next_tab_auto_complete {
		fmt.Printf("\r\n%s", strings.Join(matches, "  "))
		fmt.Printf("\r\n%s%s", cfg.shell_prompt(), ls.line)
		return
	}

	ls.next_tab_auto_complete = true
}

func (ls *line_state) get_prefix_start() (start int, is_cmd bool) {
	prefix_start := ls.cursor_idx
	for prefix_start >= 1 {
		if ls.line[prefix_start-1] == ' ' {
			break
		}
		prefix_start--
	}

	cmd_start := 0
	for cmd_start < len(ls.line) {
		if ls.line[cmd_start] != ' ' {
			break
		}
		cmd_start++
	}

	return prefix_start, cmd_start == prefix_start
}

func split_prefix(prefix string) (path string, name string) {
	idx := strings.LastIndexByte(prefix, '/')
	if idx == -1 {
		return "", prefix
	}
	return prefix[:idx+1], prefix[idx+1:]
}

func auto_complete_command(path string, prefix string) []string {
	matches_set := make(map[string]struct{})

	for command := range GetBuiltInCommands() {
		if strings.HasPrefix(command, prefix) {
			matches_set[command] = struct{}{}
		}
	}

	for dir := range strings.SplitSeq(os.Getenv("PATH"), ":") {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
				continue
			}

			matches_set[entry.Name()] = struct{}{}
		}
	}

	matches := make([]string, 0, len(matches_set))
	for cmd := range matches_set {
		matches = append(matches, cmd)
	}

	slices.Sort(matches)

	return matches
}

func auto_complete_file_name(path string, prefix string) []string {
	matches_set := make(map[string]struct{})

	if path == "" {
		path = "."
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return []string{}
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}

		matches_set[entry.Name()] = struct{}{}
	}

	matches := make([]string, 0, len(matches_set))
	for cmd := range matches_set {
		matches = append(matches, cmd)
	}

	slices.Sort(matches)

	return matches
}

func longest_common_prefix(strs []string) string {
	lcp := strings.Builder{}
	for i := 0; ; i++ {
		var currChar byte
		for j, str := range strs {
			if i >= len(str) {
				return lcp.String()
			}
			if j == 0 {
				currChar = str[i]
			}
			if str[i] != currChar {
				return lcp.String()
			}
		}
		lcp.WriteByte(currChar)
	}
}

func move_cursor_left(n int) {
	for range n {
		fmt.Print("\b")
	}
}

func move_cursor_right(n int) {
	for range n {
		fmt.Print("\x1b[1C")
	}
}

func read_line(cfg *config, stdin *bufio.Reader) (string, error) {
	line := &line_state{
		line:                   make([]byte, 0),
		prev_line:              make([]byte, 0),
		history_idx:            -1,
		next_tab_auto_complete: false,
	}

	fmt.Print(cfg.shell_prompt())

	for {
		char, err := stdin.ReadByte()
		if err != nil {
			return "", err
		}

		switch char {
		case '\n', '\r':
			return string(line.line), nil
		case EOT:
			return "", fmt.Errorf("End of Text")
		case '\t':
			line.tab_completion(cfg)
		case DEL:
			line.delete_byte()
		case ESC:
			seq := make([]byte, 2)
			stdin.Read(seq)
			line.handle_arrow_key(cfg, string(seq))
		default:
			if char < 32 || char > 126 {
				fmt.Printf("\r\nline='%+v' idx=%+v", string(line.line), line.cursor_idx)
			} else {
				line.insert_byte(char)
			}

		}

		if char != '\t' {
			line.next_tab_auto_complete = false
		}
	}
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

	fmt.Print("\r\nWelcome to BitBash! Here is a list of supported features:\r\n\r\n")

	fmt.Print("Redirection:\r\n")
	fmt.Print("    <            Redirect stdin from file\r\n")
	fmt.Print("    >  >>        Redirect stdout to file\r\n")
	fmt.Print("    2>  2>>      Redirect stderr to file\r\n")
	fmt.Print("    &>  &>>      Redirect botth stdin and stderr to file\r\n")
	fmt.Print("\r\n")
	fmt.Print("Piping:\r\n")
	fmt.Print("    |            Redirect the stdout of one command to the stdin of another\r\n")
	fmt.Print("\r\n")
	fmt.Print("Autocomplete:\r\n")
	fmt.Print("    TAB          Attempt to complete or partially complete the name of the commnd\r\n")
	fmt.Print("    TAB TAB      If multiple autocomplete matches print them all\r\n")
	fmt.Print("\r\n")
	fmt.Print("Quoting:\r\n")
	fmt.Print("    '...'            Characters quoted in single quotes preserve their literal value\r\n")
	fmt.Print(`    "..."            Same as single quotes but processes the escape sequences \\, \$, and \"`, "\r\n")
	fmt.Print("Command History:\r\n")
	fmt.Print("\r\n")
	fmt.Print("    Up Arrow     Replace current line with previous command in history\r\n")
	fmt.Print("    Down Arrow   Replace current line with next command in history\r\n")

	fmt.Print("\r\nType help for a list of builtin commands.\r\n")
}

func run_repl(cfg *config) error {
	// printWelcomeMessage()
	stdin := bufio.NewReader(os.Stdin)

	for {
		input, err := read_line(cfg, stdin)
		if err != nil {
			return err
		}

		fmt.Print("\r\n")

		input = strings.TrimSpace(input)
		if len(input) == 0 {
			continue
		}

		cfg.history = append(cfg.history, input)

		command, err := NewCommand(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "shell: %s\r\n", err)
			continue
		}

		command.Exec(cfg)
	}
}
