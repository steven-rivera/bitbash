package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	stdin := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("$ ")
		command, err := stdin.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		command = strings.TrimSuffix(command, "\n")

		if strings.HasPrefix(command, "exit") {
			split := strings.SplitAfterN(command, " ", 2)
			exitCode, err := strconv.Atoi(split[1])
			if err == nil {
				os.Exit(exitCode)
			}
		}

		fmt.Println(command + ": command not found")
	}
}
