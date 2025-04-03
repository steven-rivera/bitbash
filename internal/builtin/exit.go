package builtin

import (
	"fmt"
	"os"
	"strconv"
)

func HandlerExit(args []string, outFile *os.File, errFile *os.File) error {
	if len(args) != 1 {
		return fmt.Errorf("exit: expected 1 argument got %d", len(args))
	}

	exitCode, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("exit: invalid exit code '%s'\n", args[0])
	}
	os.Exit(exitCode)
	return nil
}
