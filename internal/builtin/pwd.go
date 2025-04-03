package builtin

import (
	"fmt"
	"os"
)

func HandlerPwd(args []string, outFile *os.File, errFile *os.File) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pwd: %w", err)
	}
	fmt.Fprintf(outFile, "%s\r\n", workingDir)
	return nil
}
