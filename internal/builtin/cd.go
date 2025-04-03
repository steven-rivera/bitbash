package builtin

import (
	"fmt"
	"os"
)

func HandlerCd(args []string, outFile *os.File, errFile *os.File) error {
	if len(args) != 1 {
		return fmt.Errorf("cd: expected 1 argument got %d", len(args))
	}
	dir := args[0]
	if dir == "~" {
		dir, _ = os.UserHomeDir()
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("cd: %s: No such file or directory", dir)
	}
	return nil
}
