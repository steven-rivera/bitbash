package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func handlerExec(program string, args []string, outFile *os.File, errFile *os.File) error {
	cmd := exec.Command(program, args...)
	if cmd.Err != nil {
		return fmt.Errorf("%s: command not found", program)
	}
	//fmt.Printf("Path: '%s', Dir: '%s'\r\n", cmd.Path, cmd.Dir)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Run()
	// if err != nil {
	// 	return fmt.Errorf("error: %s", err)
	// }

	if stdout.Len() != 0 {
		fmt.Fprint(outFile, strings.ReplaceAll(stdout.String(), "\n", "\r\n"))
	}
	if stderr.Len() != 0 {
		fmt.Fprint(errFile, strings.ReplaceAll(stderr.String(), "\n", "\r\n"))
	}

	return nil
}
