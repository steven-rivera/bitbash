package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/steven-rivera/shell/internal/builtin"
)

func readLine(stdin *bufio.Reader) (string, error) {
	fmt.Print("$ ")
	currentLine := strings.Builder{}
	lastCharTab := false
	for {
		char, err := stdin.ReadByte()
		if err != nil {
			return "", fmt.Errorf("Error reading input: %w", err)
		}

		if char == '\n' || char == '\r' {
			fmt.Print("\r\n")
			break
		}
		if char == '\t' {
			matches := autoComplete(currentLine.String())
			if len(matches) == 0 {
				fmt.Print("\a")
			} else if len(matches) == 1 {
				fmt.Printf("\r\x1b[0K$ %s ", matches[0])
				currentLine.Reset()
				currentLine.WriteString(fmt.Sprintf("%s ", matches[0]))
			} else {
				if lastCharTab {
					fmt.Printf("\r\n%s\r\n", strings.Join(matches, "  "))
					fmt.Printf("$ %s", currentLine.String())
				} else {
					fmt.Print("\a")
					lcp := longestCommonPrefix(matches)
					if lcp == currentLine.String() {
						lastCharTab = true
					} else {
						fmt.Printf("\r\x1b[0K$ %s", lcp)
						currentLine.Reset()
						currentLine.WriteString(lcp)
					}
				}

			}
			continue
		}
		if char == 3 {
			return "", fmt.Errorf("Siginterupt")
		}

		currentLine.WriteByte(char)
		fmt.Printf("%c", char)
		lastCharTab = false

	}
	return currentLine.String(), nil
}

func coalesceQuotes(argStr string) ([]string, error) {
	if argStr == "" {
		return []string{}, nil
	}

	if strings.Count(argStr, "'")%2 != 0 && strings.Count(argStr, "\"")%2 != 0 {
		return []string{}, fmt.Errorf("missing closing quote")
	}

	coalescedArgs := make([]string, 0)
	currentArg := strings.Builder{}
	inSingleQuotes, inDoubleQuotes := false, false
	for i := 0; i < len(argStr); i++ {
		char := argStr[i]

		if char == '\'' {
			if !inDoubleQuotes {
				inSingleQuotes = !inSingleQuotes
				continue
			}
		}
		if char == '"' {
			if !inSingleQuotes {
				inDoubleQuotes = !inDoubleQuotes
				continue
			}
		}

		if char == '\\' {
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
		}

		if char == ' ' && !inSingleQuotes && !inDoubleQuotes {
			if currentArg.Len() > 0 {
				coalescedArgs = append(coalescedArgs, currentArg.String())
				currentArg.Reset()
			}
			continue
		}

		currentArg.WriteByte(char)
	}

	if currentArg.Len() > 0 {
		coalescedArgs = append(coalescedArgs, currentArg.String())
	}

	return coalescedArgs, nil
}

func parseRedirection(args []string) (commandArgs []string, outFile *os.File, errFile *os.File, err error) {
	for index, arg := range args {
		if arg == ">" || arg == "1>" || arg == "2>" {
			commandArgs, filePathSlice := args[:index], args[index+1:]
			if len(filePathSlice) != 1 {
				return nil, nil, nil, fmt.Errorf("%s: expected 1 argument got %d", arg, len(filePathSlice))
			}
			file, err := os.Create(filePathSlice[0])
			if err != nil {
				return nil, nil, nil, fmt.Errorf("%s: %w", arg, err)
			}

			switch arg {
			case "2>":
				return commandArgs, os.Stdout, file, nil
			default:
				return commandArgs, file, os.Stderr, nil
			}
		}
		if arg == ">>" || arg == "1>>" || arg == "2>>" {
			commandArgs, filePathSlice := args[:index], args[index+1:]
			if len(filePathSlice) != 1 {
				return nil, nil, nil, fmt.Errorf("%s: expected 1 argument got %d", arg, len(filePathSlice))
			}
			file, err := os.OpenFile(filePathSlice[0], os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o666)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("%s: %w", arg, err)
			}

			switch arg {
			case "2>>":
				return commandArgs, os.Stdout, file, nil
			default:
				return commandArgs, file, os.Stderr, nil
			}
		}
	}
	return args, os.Stdout, os.Stderr, nil
}

func longestCommonPrefix(strs []string) string {
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

func autoComplete(partial string) []string {
	matches := make([]string, 0)
	for command, _ := range builtin.GetBuiltInCommands() {
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
