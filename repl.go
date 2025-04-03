package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/steven-rivera/shell/internal/builtin"
)

func startREPL() {
	stdin := bufio.NewReader(os.Stdin)
	builtInCommands := builtin.GetBuiltInCommands()

	for {
		input, err := readLine(stdin)
		if err != nil {
			return
		}
		input = strings.TrimSpace(input)
		splitInput, err := coalesceQuotes(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing input:", err)
			os.Exit(1)
		}

		if len(splitInput) == 0 {
			continue
		}
		command := splitInput[0]
		args := splitInput[1:]

		// fmt.Printf("cmd: `%s`, args: %#v\n", command, args)

		args, outputFile, errFile, err := parseRedirection(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		if callback, ok := builtInCommands[command]; ok {
			err := callback(args, outputFile, errFile)
			if err != nil {
				fmt.Fprintf(errFile, "%s\r\n", err)
			}
		} else {
			err := handlerExec(command, args, outputFile, errFile)
			if err != nil {
				fmt.Fprintf(errFile, "%s\r\n", err)
			}
		}

		if outputFile != os.Stdout {
			outputFile.Close()
		}
		if errFile != os.Stderr {
			errFile.Close()
		}
	}
}
