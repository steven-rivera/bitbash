package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// Needed so that text printed to stdout when function is tested
// isn't printed to `go test` output if function fails test
func capture_stdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	return string(out)
}

func TestReadLine(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "command",
			input:    "echo 'Hello World'\n",
			expected: "echo 'Hello World'",
		},
		{
			name:     "command completion",
			input:    "ech\t\n",
			expected: "echo ",
		},
		{
			name:     "partial tab completion",
			input:    "git-u\t\n",
			expected: "git-upload-",
		},
		{
			name:     "delete characters",
			input:    fmt.Sprintf("echoss%c%c\n", DEL, DEL),
			expected: "echo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			var result string

			_ = capture_stdout(func() {
				result, _ = read_line(&config{}, bufio.NewReader(strings.NewReader(tc.input)))
			})

			if result != tc.expected {
				t.Fatalf("expected: %#v, got: %#v", tc.expected, result)
			}
		})
	}
}
