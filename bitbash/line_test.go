package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// Needed so that text isn't printed to stdout when function is tested

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()

	os.Stdout = w
	f()
	os.Stdout = old

	w.Close()
	out, _ := io.ReadAll(r)
	r.Close()
	
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
			name:     "partial command tab completion",
			input:    "git-u\t\n",
			expected: "git-upload-",
		},
		{
			name:     "delete characters",
			input:    fmt.Sprintf("echoss%c%c\n", DEL, DEL),
			expected: "echo",
		},
		{
			name:     "move cursor back and insert space",
			input:    fmt.Sprintf("echo%s \n", CURSOR_BACK),
			expected: "ech o",
		},
		{
			name:     "file tab completion",
			input:    "cat main\t\n",
			expected: "cat main.go ",
		},
		{
			name:     "file tab completion #2",
			input:    fmt.Sprintf("cat repl.go%s%s%s%s%s%s%smain\t\n", CURSOR_BACK, CURSOR_BACK, CURSOR_BACK, CURSOR_BACK, CURSOR_BACK, CURSOR_BACK, CURSOR_BACK),
			expected: "cat main.go repl.go",
		},
		{
			name:     "file tab completion inside directory",
			input:    "cat ../test/foo.\t\n",
			expected: "cat ../test/foo.txt ",
		},
		{
			name:     "partial file tab completion inside directory",
			input:    "cat ../test/b\t\n",
			expected: "cat ../test/ba",
		},
		{
			name:     "directory tab completion",
			input:    "cd ../tes\t\n",
			expected: "cd ../test/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			var result string

			_ = captureStdout(func() {
				result, _ = ReadLine(&Config{
					StdinReader: bufio.NewReader(strings.NewReader(tc.input)),
				})
			})

			if result != tc.expected {
				t.Fatalf("expected: %#v, got: %#v", tc.expected, result)
			}
		})
	}
}
