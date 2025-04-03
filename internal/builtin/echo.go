package builtin

import (
	"fmt"
	"os"
)

func HandlerEcho(args []string, outFile *os.File, errFile *os.File) error {
	for _, arg := range args {
		fmt.Fprintf(outFile, "%s ", arg)
	}
	fmt.Fprint(outFile, "\r\n")
	return nil
}
