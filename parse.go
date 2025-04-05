package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

const ETX = byte(3)   // Ctrl+C (SIGNINT)
const DEL = byte(127) // Backspace

func ReadLine(cfg *config, stdin *bufio.Reader) (string, error) {
	currentLine := []byte{}
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
				prevCharWasTab = true
			}
		case ETX:
			return "", fmt.Errorf("SIGINT")
		case DEL:
			if len(currentLine) != 0 {
				// Move cursor back, print space, then move back again
				fmt.Print("\b \b")
				currentLine = currentLine[:len(currentLine)-1]
			}

		default:
			currentLine = append(currentLine, char)
			fmt.Printf("%c", char)
			prevCharWasTab = false
		}
	}
}

func ShellPrompt(cfg *config) string {
	//if cut, ok := strings.CutPrefix(cfg.currDirectory, cfg.homeDirectory); ok {
	//	cfg.currDirectory = fmt.Sprintf("~%s", cut)
	//}
	//userNameBlueBold := fmt.Sprintf("%s%s%s%s", BLUE, BOLD, cfg.userName, RESET)
	//currDirGreenBold := fmt.Sprintf("%s%s%s%s", GREEN, BOLD, cfg.currDirectory, RESET)
	//return fmt.Sprintf("%s:%s\r\n$ ", userNameBlueBold, currDirGreenBold)
	return "$ "
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

func CoalesceQuotes(argStr string) ([]string, error) {
	singleQuoteCount := strings.Count(argStr, "'")
	doubleQuoteCount := strings.Count(argStr, "\"")

	if (singleQuoteCount%2 != 0 && doubleQuoteCount == 0) ||
		(doubleQuoteCount%2 != 0 && singleQuoteCount == 0) {
		return []string{}, fmt.Errorf("missing closing quote")
	}

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

	return coalescedArgs, nil
}

func ParseRedirection(args []string) ([]string, IOStream, error) {
	redirectOperators := map[string]bool{
		">":   false,
		"1>":  false,
		"2>":  false,
		"&>":  false,
		">>":  true,
		"1>>": true,
		"2>>": true,
		"&>>": true,
	}

	commandArgs := []string{}
	commandIO := IOStream{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		appendMode, isRedirOp := redirectOperators[arg]

		if !isRedirOp {
			commandArgs = append(commandArgs, arg)
			continue
		}

		flags := os.O_WRONLY | os.O_CREATE
		if appendMode {
			flags |= os.O_APPEND
		}

		// Make sure oporater is not last argument
		if i == len(args)-1 {
			return nil, IOStream{}, fmt.Errorf("%s: expected file name", arg)
		}

		filePath := args[i+1]
		file, err := os.OpenFile(filePath, flags, 0o666)
		if err != nil {
			return nil, IOStream{}, fmt.Errorf("%s: %w", arg, err)
		}

		switch arg {
		case ">", "1>", "1>>", ">>":
			commandIO.Stdout = file
		case "2>", "2>>":
			commandIO.Stderr = file
		case "&>", "&>>":
			commandIO.Stdout = file
			commandIO.Stderr = file
		}

		i++
	}
	return commandArgs, commandIO, nil
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
