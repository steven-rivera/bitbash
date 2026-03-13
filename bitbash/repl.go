package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type line_state struct {
	line             []byte
	prev_line        []byte
	cursor_idx       int
	history_idx      int
	show_all_matches bool
}

func new_line_state() *line_state {
	return &line_state{
		line:             make([]byte, 0),
		prev_line:        make([]byte, 0),
		cursor_idx:       0,
		history_idx:      -1,
		show_all_matches: false,
	}
}

func (ls *line_state) insert_byte(b byte) {
	if ls.cursor_idx == len(ls.line) {
		ls.line = append(ls.line, b)
	} else {
		ls.line = append(ls.line, 0)
		copy(ls.line[ls.cursor_idx+1:], ls.line[ls.cursor_idx:])
		ls.line[ls.cursor_idx] = b
	}

	ls.cursor_idx++
	txt_after_cursor := ls.line[ls.cursor_idx:]

	fmt.Printf("%c%s", b, txt_after_cursor)
	move_cursor_back(len(txt_after_cursor))
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
	txt_after_cursor := ls.line[ls.cursor_idx:]

	move_cursor_back(1)
	fmt.Printf("%s%s", CLEAR_FROM_CURSOR, txt_after_cursor)
	move_cursor_back(len(txt_after_cursor))
}

func (ls *line_state) clear_line() {
	move_cursor_back(ls.cursor_idx)
	fmt.Print(CLEAR_FROM_CURSOR)
	ls.cursor_idx = 0
}

func (ls *line_state) set_line(line string) {
	ls.clear_line()

	ls.line = []byte(line)
	ls.cursor_idx = len(ls.line)
	fmt.Printf("%s", ls.line)
}

func (ls *line_state) handle_arrow_key(cfg *config, seq string) {
	switch seq {
	case CURSOR_UP:
		ls.handle_arrow_key_up(cfg)
	case CURSOR_DOWN:
		ls.handle_arrow_key_down(cfg)
	case CURSOR_FORWARD:
		ls.handle_arrow_key_right()
	case CURSOR_BACK:
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
	move_cursor_forward(1)
}

func (ls *line_state) handle_arrow_key_left() {
	if ls.cursor_idx == 0 {
		fmt.Print("\a")
		return
	}

	ls.cursor_idx--
	move_cursor_back(1)
}

func (ls *line_state) print_match(prefix_start int, match match) {
	before := string(ls.line[:prefix_start])
	after := string(ls.line[ls.cursor_idx:])
	new_line := ""

	if match.is_dir {
		new_line = fmt.Sprintf("%s%s/%s", before, match.full_path(), after)
	} else {
		new_line = fmt.Sprintf("%s%s %s", before, match.full_path(), after)
	}

	ls.set_line(new_line)
	move_cursor_back(len(after))
	ls.cursor_idx -= len(after)
}

func (ls *line_state) print_partial_match(prefix_start int, dir, partial string) {
	before := string(ls.line[:prefix_start])
	after := string(ls.line[ls.cursor_idx:])
	new_line := ""

	new_line = fmt.Sprintf("%s%s%s%s", before, dir, partial, after)

	ls.set_line(new_line)
	move_cursor_back(len(after))
	ls.cursor_idx -= len(after)
}

func (ls *line_state) tab_completion(cfg *config) {
	start, is_cmd := ls.get_prefix_start()
	prefix := parse_prefix(string(ls.line[start:ls.cursor_idx]))
	matches := auto_complete_prefix(prefix, is_cmd)

	if len(matches) == 0 {
		fmt.Print("\a")
		return
	}
	if len(matches) == 1 {
		ls.print_match(start, matches[0])
		return
	}

	fmt.Print("\a")

	if lcp := longest_common_match_prefix(matches); lcp != prefix.base {
		ls.print_partial_match(start, prefix.dir, lcp)
		ls.show_all_matches = false
		return
	}

	// if TAB pressed twice in sequence, print all matches on new line
	if ls.show_all_matches {
		fmt.Printf("\r\n%s\r\n%s%s", join_matches(matches), cfg.shell_prompt(), ls.line)
		return
	}

	ls.show_all_matches = true
}

// Returns the starting index of the substring that will be used
// to attempt to perform autocompletion when TAB is pressed.
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

func read_control_sequence_introducer(stdin *bufio.Reader) []byte {
	char, err := stdin.ReadByte()
	if err != nil || char != '[' {
		return []byte{}
	}

	csi := []byte{ESC, '['}

	for {
		char, err := stdin.ReadByte()
		if err != nil {
			return []byte{}
		}

		csi = append(csi, char)

		// CSI sequences end in a char in the following range
		if '\x40' <= char && char <= '\x7E' {
			return csi
		}
	}
}

func read_line(cfg *config, stdin *bufio.Reader) (string, error) {

	fmt.Print(cfg.shell_prompt())
	line := new_line_state()

	for {
		char, err := stdin.ReadByte()
		if err != nil {
			return "", err
		}

		switch char {
		case '\r', '\n':
			return string(line.line), nil
		case EOT:
			return "", fmt.Errorf("End of Text")
		case '\t':
			line.tab_completion(cfg)
		case DEL:
			line.delete_byte()
		case ESC:
			line.handle_arrow_key(cfg, string(read_control_sequence_introducer(stdin)))
		default:
			if char < 32 || char > 126 {
				continue
			}

			line.insert_byte(char)
		}

		if char != '\t' {
			line.show_all_matches = false
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
	fmt.Print("    TAB          Attempt to complete or partially complete the name of a commnd or file\r\n")
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
