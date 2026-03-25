package main

import (
	"fmt"
	"strings"
)

func Tokenize(input string) ([]string, error) {
	var tokens []string
	var curr strings.Builder

	inSingle, inDouble := false, false

	for i := 0; i < len(input); i++ {
		c := input[i]

		if c == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if c == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if c == '\\' && i+1 < len(input) {
			next := input[i+1]

			// Outside quotes: escape anything
			if !inSingle && !inDouble {
				curr.WriteByte(next)
				i++
				continue
			}

			// Inside double quotes: only escape specific chars
			if inDouble && (next == '\\' || next == '$' || next == '"') {
				curr.WriteByte(next)
				i++
				continue
			}
		}

		// Handle token splitting
		if c == ' ' && !inSingle && !inDouble {
			if curr.Len() > 0 {
				tokens = append(tokens, curr.String())
				curr.Reset()
			}
			continue
		}

		curr.WriteByte(c)
	}

	if curr.Len() > 0 {
		tokens = append(tokens, curr.String())
	}

	if inSingle || inDouble {
		return nil, fmt.Errorf("missing closing quote")
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
	var pipelineTokens [][]string
	var curr []string

	for _, token := range tokens {
		if token == "|" {
			pipelineTokens = append(pipelineTokens, curr)
			curr = nil
			continue
		}

		curr = append(curr, token)
	}

	if len(curr) != 0 {
		pipelineTokens = append(pipelineTokens, curr)
	}

	return pipelineTokens
}