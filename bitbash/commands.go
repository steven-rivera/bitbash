package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type IO struct {
	in  *os.File
	out *os.File
	err *os.File
}

type Command struct {
	Name      string
	IsBuiltin bool
	Args      []string
	IO
}

func (cmd *Command) init(cmd_tokens []string) {
	for i := 0; i < len(cmd_tokens); {
		token := cmd_tokens[i]

		if _, ok := REDIRECTION_OPS[token]; ok {
			cmd.redirect(token, cmd_tokens[i+1])

			i += 2
			continue
		}

		if cmd.Name == "" {
			cmd.Name = token
			_, ok := BUILTIN_CMDS[token]
			cmd.IsBuiltin = ok

		} else {
			cmd.Args = append(cmd.Args, token)
		}

		i++
	}
}

func (cmd *Command) redirect(operator, file_name string) error {
	var flags int

	if operator == "<" {
		flags = os.O_RDONLY
	} else {
		flags = os.O_WRONLY | os.O_CREATE

		switch operator {
		case ">", "1>", "2>", "&>":
			flags |= os.O_TRUNC
		case ">>", "1>>", "2>>", "&>>":
			flags |= os.O_APPEND
		}
	}

	file, err := os.OpenFile(file_name, flags, 0o666)
	if err != nil {
		return fmt.Errorf("%s: %w", operator, err)
	}

	switch operator {
	case "<":
		cmd.in = file
	case ">", "1>", "1>>", ">>":
		cmd.out = file
	case "2>", "2>>":
		cmd.err = file
	case "&>", "&>>":
		cmd.out = file
		cmd.err = file
	}

	return nil
}

func (cmd *Command) ClosePipes() {
	if cmd.in != os.Stdin {
		cmd.in.Close()
	}
	if cmd.out != os.Stdout {
		cmd.out.Close()
	}
	if cmd.err != os.Stderr {
		cmd.err.Close()
	}
}

func (cmd *Command) Run(wg *sync.WaitGroup, cfg *Config) {
	if cmd.IsBuiltin {
		cmd.runBuiltin(wg, cfg)
	} else {
		cmd.runExec(wg)
	}
}

func (cmd *Command) runExec(wg *sync.WaitGroup) {
	defer wg.Done()
	defer cmd.ClosePipes()

	exec := exec.Command(cmd.Name, cmd.Args...)
	if exec.Err != nil {
		fmt.Fprintf(os.Stderr, "%s: command not found\r\n", cmd.Name)
		return
	}

	exec.Stdin = cmd.in
	exec.Stdout = cmd.out
	exec.Stderr = cmd.err

	if err := exec.Start(); err != nil {
		return
	}

	exec.Wait()
}

func (cmd *Command) runBuiltin(wg *sync.WaitGroup, cfg *Config) {
	defer wg.Done()
	defer cmd.ClosePipes()

	BUILTIN_CMDS[cmd.Name].Handler(cmd, cfg)
}

func Tokenize(input string) ([]string, error) {
	tokens, curr_token := make([]string, 0), strings.Builder{}
	inSingleQuotes, inDoubleQuotes := false, false

	for i := 0; i < len(input); i++ {
		char := input[i]
		switch char {
		case '\'':
			if !inDoubleQuotes {
				inSingleQuotes = !inSingleQuotes
				continue
			}
		case '"':
			if !inSingleQuotes {
				inDoubleQuotes = !inDoubleQuotes
				continue
			}
		case '\\':
			if !inSingleQuotes && !inDoubleQuotes {
				if i+1 < len(input) {
					curr_token.WriteByte(input[i+1])
					i++
				}
				continue
			}
			if inDoubleQuotes {
				if i+1 < len(input) {
					switch input[i+1] {
					case '\\', '$', '"':
						curr_token.WriteByte(input[i+1])
						i++
						continue
					}
				}
			}
		case ' ':
			if !inSingleQuotes && !inDoubleQuotes {
				if curr_token.Len() > 0 {
					tokens = append(tokens, curr_token.String())
					curr_token.Reset()
				}
				continue
			}
		}
		curr_token.WriteByte(char)
	}

	if curr_token.Len() > 0 {
		tokens = append(tokens, curr_token.String())
	}

	if inDoubleQuotes || inSingleQuotes {
		return []string{}, fmt.Errorf("missing closing quote")
	}

	if err := validateTokens(tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

func validateTokens(tokens []string) error {
	for i, token := range tokens {
		isLastToken := i == (len(tokens) - 1)

		if token == "|" && (i == 0 || isLastToken) {
			return fmt.Errorf("missing command before/after '|'")
		}

		_, ok := REDIRECTION_OPS[token]
		if ok && (isLastToken || tokens[i+1] == "|") {
			return fmt.Errorf("expected file name after '%s'", token)
		}
	}
	return nil
}

func splitOnPipes(tokens []string) [][]string {
	var pipeline [][]string
	var cmdTokens []string

	for _, token := range tokens {
		if token == "|" {
			pipeline = append(pipeline, cmdTokens)
			cmdTokens = nil
			continue
		}

		cmdTokens = append(cmdTokens, token)
	}

	if len(cmdTokens) != 0 {
		pipeline = append(pipeline, cmdTokens)
	}

	return pipeline
}

func CreatePipeline(tokens []string) ([]*Command, error) {
	cmds_tokens := splitOnPipes(tokens)
	cmds := make([]*Command, 0, len(cmds_tokens))

	for range len(cmds_tokens) {
		cmds = append(cmds, &Command{})
	}

	connect_pipes(cmds)

	for i, cmd := range cmds {
		cmd.init(cmds_tokens[i])
	}

	return cmds, nil
}

func connect_pipes(cmds []*Command) {
	cmds[0].in = os.Stdin

	for i := 0; i < len(cmds)-1; i++ {
		r, w, _ := os.Pipe()

		cmds[i].out = w
		cmds[i].err = os.Stderr
		cmds[i+1].in = r
	}

	cmds[len(cmds)-1].out = os.Stdout
	cmds[len(cmds)-1].err = os.Stderr
}
