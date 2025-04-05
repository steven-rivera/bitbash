package main

import (
	"bufio"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestReadLine(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "builtin tab completion",
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
			res, _ := ReadLine(&config{}, bufio.NewReader(strings.NewReader(tc.input)))
			if res != tc.expected {
				t.Fatalf("expected: %#v, got: %#v", tc.expected, res)
			}

		})
	}
}

func TestSplitInput(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "coalesce spaces",
			input:    "echo example     test",
			expected: []string{"echo", "example", "test"},
		},
		{
			name:     "remove single quotes",
			input:    "echo 'hello world'",
			expected: []string{"echo", "hello world"},
		},
		{
			name:     "coalesce single quotes",
			input:    "echo 'test''shell'",
			expected: []string{"echo", "testshell"},
		},
		{
			name:     "preserve double quote in single quotes",
			input:    `echo 'left"right'`,
			expected: []string{"echo", `left"right`},
		},
		{
			name:     "preserve space in single quotes",
			input:    "echo 'world    hello'",
			expected: []string{"echo", "world    hello"},
		},
		{
			name:     "remove double quotes",
			input:    `echo "hello world"`,
			expected: []string{"echo", "hello world"},
		},
		{
			name:     "preserve space in double quotes",
			input:    `echo "world    hello"`,
			expected: []string{"echo", "world    hello"},
		},
		{
			name:     "preserve single quote in double quotes",
			input:    `echo "world's"`,
			expected: []string{"echo", "world's"},
		},
		{
			name:     "coalesce double quotes",
			input:    `echo "test""shell"`,
			expected: []string{"echo", "testshell"},
		},
		{
			name:     "non-quoted backslash preserves space",
			input:    `echo world\ \ \ \ \ \ script`,
			expected: []string{"echo", "world      script"},
		},
		{
			name:     "non-quoted backslash preserves quotes",
			input:    `echo \'\"script example\"\'`,
			expected: []string{"echo", `'"script`, `example"'`},
		},
		{
			name:     "non-quoted backslash preserves char",
			input:    `echo example\nhello`,
			expected: []string{"echo", "examplenhello"},
		},
		{
			name:     "backslash within single-quotes is not interpreted",
			input:    `echo 'shell\nscript'`,
			expected: []string{"echo", `shell\nscript`},
		},
		{
			name:     `backslash within double-quotes escapes '\'`,
			input:    `echo "hello'script'\\n'world"`,
			expected: []string{"echo", `hello'script'\n'world`},
		},
		{
			name:     "backslash within double-quotes escapes '$'",
			input:    `echo "hello \$HOME"`,
			expected: []string{"echo", `hello $HOME`},
		},
		{
			name:     `backslash within double-quotes escapes '"'`,
			input:    `echo "example\"insidequotes"hello\"`,
			expected: []string{"echo", `example"insidequoteshello"`},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, _ := SplitInput(tc.input)
			reflect.DeepEqual(res, tc.expected)
			if !reflect.DeepEqual(res, tc.expected) {
				t.Fatalf("expected: %#v, got: %#v", tc.expected, res)
			}
		})
	}
}
