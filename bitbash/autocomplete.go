package main

import (
	"os"
	"slices"
	"strings"
)

type Prefix struct {
	Directory  string
	PrefixBase string
	IsCmd      bool
}

func NewPrefix(prefixStr string, isCmd bool) Prefix {
	idx := strings.LastIndexByte(prefixStr, '/')
	if idx == -1 {
		return Prefix{
			Directory:  "",
			PrefixBase: prefixStr,
			IsCmd:      isCmd,
		}
	}
	return Prefix{
		Directory:  prefixStr[:idx+1],
		PrefixBase: prefixStr[idx+1:],
		IsCmd:      isCmd,
	}
}

type Match struct {
	PrefixDirectory string
	Match           string
	IsDir           bool
	IsExec          bool
}

func (m Match) ToString() string {
	return m.PrefixDirectory + m.Match
}

func AutoCompletePrefix(prefix Prefix) []Match {
	if prefix.IsCmd {
		return CommandsWithPrefix(prefix)
	}

	if prefix.Directory == "" {
		return FilesWithPrefix(prefix, ".")
	}

	return FilesWithPrefix(prefix, prefix.Directory)
}

func CommandsWithPrefix(prefix Prefix) []Match {
	matches := []Match{}

	// If Prefix struct has no directory then search builtin commands and commands
	// in PATH that have the prefix prefix.PrefixBase, else search for commands with
	// the prefix prefix.PrefixBase only in the directory prefix.Directory

	if prefix.Directory == "" {
		for command := range BUILTIN_CMDS {
			if strings.HasPrefix(command, prefix.PrefixBase) {
				return append(matches, Match{
					PrefixDirectory: prefix.Directory,
					Match:           command,
					IsDir:           false,
					IsExec:          true,
				})
			}
		}

		for dir := range strings.SplitSeq(os.Getenv("PATH"), ":") {
			for _, match := range FilesWithPrefix(prefix, dir) {
				if match.IsExec {
					matches = append(matches, match)
				}
			}
		}

	} else {
		for _, match := range FilesWithPrefix(prefix, prefix.Directory) {
			if match.IsExec {
				matches = append(matches, match)
			}
		}
	}

	slices.SortFunc(matches, func(a, b Match) int {
		return strings.Compare(a.Match, b.Match)
	})

	return matches
}

func FilesWithPrefix(prefix Prefix, searchDirectory string) []Match {
	entries, err := os.ReadDir(searchDirectory)
	if err != nil {
		return []Match{}
	}

	matches := []Match{}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), prefix.PrefixBase) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		is_exec := !entry.IsDir() && info.Mode()&0111 != 0

		matches = append(matches, Match{
			PrefixDirectory: prefix.Directory,
			Match:           entry.Name(),
			IsDir:           entry.IsDir(),
			IsExec:          is_exec,
		})
	}

	slices.SortFunc(matches, func(a, b Match) int {
		return strings.Compare(a.Match, b.Match)
	})

	return matches
}

func LongestCommonPrefix(matches []Match) string {
	lcp := strings.Builder{}
	for i := 0; ; i++ {
		var currChar byte
		for j, match := range matches {
			if i >= len(match.Match) {
				return lcp.String()
			}
			if j == 0 {
				currChar = match.Match[i]
			}
			if match.Match[i] != currChar {
				return lcp.String()
			}
		}
		lcp.WriteByte(currChar)
	}
}

func JoinMatches(matches []Match) string {
	matches_str := ""
	for _, match := range matches {
		matches_str += match.Match
		if match.IsDir {
			matches_str += "/  "
		} else {
			matches_str += "  "
		}
	}
	return matches_str
}
