package main

import (
	"fmt"
	"os"
	"strings"
)

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
