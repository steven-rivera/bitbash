package builtin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func HandlerType(args []string, outFile *os.File, errFile *os.File) error {
	if len(args) != 1 {
		return fmt.Errorf("exit: expected 1 argument got %d", len(args))
	}
	commandArg := args[0]
	if _, ok := GetBuiltInCommands()[commandArg]; ok {
		fmt.Fprintf(outFile, "%s is a shell builtin\r\n", commandArg)
		return nil
	}

	pathEnv := os.Getenv("PATH")
	for dir := range strings.SplitSeq(pathEnv, ":") {
		dirEntries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, dirEntry := range dirEntries {
			if !dirEntry.IsDir() && dirEntry.Name() == commandArg {
				fmt.Fprintf(outFile, "%s is %s\r\n", commandArg, filepath.Join(dir, commandArg))
				return nil
			}
		}
	}

	return fmt.Errorf("%s: not found", commandArg)
}
