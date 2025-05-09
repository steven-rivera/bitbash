package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

const (
	ETX = byte(3)   // Ctrl+C (SIGNINT)
	DEL = byte(127) // Backspace
	ESC = byte(27)  // ESC
)

const (
	CURSOR_UP      = byte('A')
	CURSOR_DOWN    = byte('B')
	CURSOR_FORWARD = byte('C')
	CURSOR_BACK    = byte('D')
)

func ReadLine(cfg *config, stdin *bufio.Reader) (string, error) {
	currentLine := []byte{}
	prevCurrentLine := []byte{}
	currHistoryIndex := -1
	prevCharWasTab := false

	fmt.Print(ShellPrompt(cfg))
	for {
		char, err := stdin.ReadByte()
		if err != nil {
			return "", err
		}

		switch char {
		case '\n', '\r':
			fmt.Print("\r\n")
			return string(currentLine), nil
		case '\t':
			matches := AutoComplete(string(currentLine))
			switch len(matches) {
			case 0:
				// no matches, print BELL char
				fmt.Print("\a")
			case 1:
				// move cursor to beginning, erase current line,
				// and replace line with match adding a space after
				fmt.Print("\r\x1b[K")
				fmt.Printf("%s%s ", ShellPrompt(cfg), matches[0])
				currentLine = []byte(fmt.Sprintf("%s ", matches[0]))
			default:
				// if TAB pressed twice in sequence, print all matches on new line
				if prevCharWasTab {
					fmt.Printf("\r\n%s\r\n", strings.Join(matches, "  "))
					fmt.Printf("%s%s", ShellPrompt(cfg), string(currentLine))
					break
				}

				// multiple matches, print BELL char
				fmt.Print("\a")
				// check for partial completions
				if lcp := LongestCommonPrefix(matches); lcp != string(currentLine) {
					fmt.Print("\r\x1b[K")
					fmt.Printf("%s%s", ShellPrompt(cfg), lcp)
					currentLine = []byte(lcp)
				}
			}
			prevCharWasTab = true
			continue
		case ETX:
			return "", fmt.Errorf("SIGINT")
		case DEL:
			if len(currentLine) != 0 {
				// Move cursor back, print space, then move back again
				fmt.Print("\b \b")
				currentLine = currentLine[:len(currentLine)-1]
			}
		case ESC:
			char, _ := stdin.ReadByte()
			// Check if a Control Sequence Introducer
			if char != '[' {
				currentLine = append(currentLine, ESC, char)
				fmt.Printf("%c%c", ESC, char)
				prevCharWasTab = false
				break
			}
			char, _ = stdin.ReadByte()
			switch char {
			case CURSOR_UP:
				if len(cfg.history) == 0 {
					fmt.Print("\a")
					break
				}

				if currHistoryIndex == -1 {
					currHistoryIndex = len(cfg.history)
				}

				if currHistoryIndex > 0 {
					currHistoryIndex--
					currentLine = []byte(cfg.history[currHistoryIndex])
					fmt.Printf("\r\x1b[K%s%s", ShellPrompt(cfg), currentLine)
				} else {
					fmt.Print("\a")
				}

			case CURSOR_DOWN:
				if currHistoryIndex == -1 {
					fmt.Print("\a")
					break
				}

				if currHistoryIndex < len(cfg.history)-1 {
					currHistoryIndex++
					currentLine = []byte(cfg.history[currHistoryIndex])
					fmt.Printf("\r\x1b[K%s%s", ShellPrompt(cfg), currentLine)
				} else {
					currentLine = prevCurrentLine
					fmt.Printf("\r\x1b[K%s%s", ShellPrompt(cfg), currentLine)
					currHistoryIndex = -1
				}
			}

		default:
			currentLine = append(currentLine, char)
			fmt.Printf("%c", char)

		}
		prevCharWasTab = false
	}
}

func ShellPrompt(cfg *config) string {
	if cut, ok := strings.CutPrefix(cfg.currDirectory, cfg.homeDirectory); ok {
		cfg.currDirectory = fmt.Sprintf("~%s", cut)
	}
	userNameBlueBold := fmt.Sprintf("%s%s%s%s", BLUE, BOLD, cfg.userName, RESET)
	currDirGreenBold := fmt.Sprintf("%s%s%s%s", GREEN, BOLD, cfg.currDirectory, RESET)
	return fmt.Sprintf("%s:%s $ ", userNameBlueBold, currDirGreenBold)
	//return "$ "
}

func AutoComplete(partial string) []string {
	matches := make([]string, 0)
	for command, _ := range GetBuiltInCommands() {
		if strings.HasPrefix(command, partial) {
			return append(matches, command)
		}
	}

	pathEnv := os.Getenv("PATH")
	for dir := range strings.SplitSeq(pathEnv, ":") {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), partial) {
				matches = append(matches, entry.Name())
			}
		}
	}
	slices.Sort(matches)
	return matches
}

func SplitInput(argStr string) ([]string, error) {
	coalescedArgs, currentArg := make([]string, 0), strings.Builder{}
	inSingleQuotes, inDoubleQuotes := false, false

	for i := 0; i < len(argStr); i++ {
		char := argStr[i]
		switch char {
		case '\'':
			if !inDoubleQuotes {
				inSingleQuotes = !inSingleQuotes
				continue
			}
		case '"':
			if !inSingleQuotes {
				inDoubleQuotes = !inDoubleQuotes
				continue
			}
		case '\\':
			if !inSingleQuotes && !inDoubleQuotes {
				if i+1 < len(argStr) {
					currentArg.WriteByte(argStr[i+1])
					i++
				}
				continue
			}
			if inDoubleQuotes {
				if i+1 < len(argStr) {
					switch argStr[i+1] {
					case '\\', '$', '"':
						currentArg.WriteByte(argStr[i+1])
						i++
						continue
					}
				}
			}
		case ' ':
			if !inSingleQuotes && !inDoubleQuotes {
				if currentArg.Len() > 0 {
					coalescedArgs = append(coalescedArgs, currentArg.String())
					currentArg.Reset()
				}
				continue
			}
		}
		currentArg.WriteByte(char)
	}

	if currentArg.Len() > 0 {
		coalescedArgs = append(coalescedArgs, currentArg.String())
	}

	if inDoubleQuotes || inSingleQuotes {
		return []string{}, fmt.Errorf("missing closing quote")
	}

	return coalescedArgs, nil
}

func ParseRedirection(parentCmd *Command) error {
	redirectOperators := map[string]bool{
		"<":   false,
		">":   false,
		"1>":  false,
		"2>":  false,
		"&>":  false,
		">>":  true,
		"1>>": true,
		"2>>": true,
		"&>>": true,
	}

	currCmd := parentCmd
	for currCmd != nil {

		cmdArgs := []string{}

		for i := 0; i < len(currCmd.Args); i++ {
			token := currCmd.Args[i]
			appendMode, isRedirOp := redirectOperators[token]

			if !isRedirOp {
				cmdArgs = append(cmdArgs, token)
				continue
			}

			flags := os.O_RDWR | os.O_CREATE
			if appendMode {
				flags |= os.O_APPEND
			} else {
				flags |= os.O_TRUNC
			}

			// Make sure oporater is not last argument
			if i == len(currCmd.Args)-1 {
				return fmt.Errorf("%s: expected file name", token)
			}

			switch token {
			case "<":
				if currCmd != parentCmd {
					return fmt.Errorf("invalid redirect of stdin when reading from pipe")
				}
			case ">", "1>", "1>>", ">>", "&>", "&>>":
				if currCmd.PipedInto != nil {
					return fmt.Errorf("invalid redirect of stdout when piping")
				}
			}

			i++
			filePath := currCmd.Args[i]
			file, err := os.OpenFile(filePath, flags, 0o666)
			if err != nil {
				return fmt.Errorf("%s: %w", token, err)
			}

			switch token {
			case "<":
				currCmd.Stdin = file
			case ">", "1>", "1>>", ">>":
				currCmd.Stdout = file
			case "2>", "2>>":
				currCmd.Stderr = file
			case "&>", "&>>":
				currCmd.Stdout = file
				currCmd.Stderr = file
			}
		}

		currCmd.Args = cmdArgs
		currCmd = currCmd.PipedInto
	}
	return nil
}

func LongestCommonPrefix(strs []string) string {
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

func ParsePipes(splitInput []string) (*Command, error) {
	defaultIO := IOStream{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	parentCmd := &Command{
		IOStream: defaultIO,
	}
	currCmd := parentCmd
	tokens := []string{}

	for i := 0; i < len(splitInput); i++ {
		token := splitInput[i]
		if token == "|" {
			if len(tokens) == 0 {
				return nil, fmt.Errorf("No command provided to write side of pipe")
			}
			currCmd.Name, currCmd.Args = tokens[0], tokens[1:]
			currCmd.PipedInto = &Command{
				IOStream: defaultIO,
			}
			currCmd = currCmd.PipedInto

			tokens = []string{}
			continue
		}

		tokens = append(tokens, token)
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("No command provided to read side of pipe")
	}
	currCmd.Name, currCmd.Args = tokens[0], tokens[1:]
	currCmd.PipedInto = nil

	// Create pipes and connect them to each subcommand
	currCmd = parentCmd
	for currCmd.PipedInto != nil {
		r, w, err := os.Pipe()
		if err != nil {
			return nil, fmt.Errorf("Error creating pipe: %w", err)
		}
		currCmd.Stdout, currCmd.PipedInto.Stdin = w, r
		currCmd = currCmd.PipedInto
	}

	return parentCmd, nil
}
