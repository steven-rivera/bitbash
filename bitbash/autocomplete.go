package main

import (
	"os"
	"slices"
	"strings"
)

type prefix struct {
	dir  string
	base string
}

type match struct {
	name    string
	prefix  prefix
	is_dir  bool
	is_exec bool
}

func (m match) full_path() string {
	return m.prefix.dir + m.name
}

func auto_complete_prefix(prefix prefix, is_cmd bool) []match {
	if prefix.dir == "" && is_cmd {
		return cmds_with_prefix(prefix)
	}

	if prefix.dir == "" {
		return files_with_prefix(prefix, ".")
	}
	return files_with_prefix(prefix, prefix.dir)
}

func cmds_with_prefix(prefix prefix) []match {
	matches := []match{}

	for command := range GetBuiltInCommands() {
		if strings.HasPrefix(command, prefix.base) {
			return append(matches, match{
				name:   command,
				prefix: prefix,
			})
		}
	}

	for dir := range strings.SplitSeq(os.Getenv("PATH"), ":") {
		for _, match := range files_with_prefix(prefix, dir) {
			if match.is_exec {
				matches = append(matches, match)
			}
		}
	}

	slices.SortFunc(matches, func(a, b match) int {
		return strings.Compare(a.name, b.name)
	})

	return matches
}

func files_with_prefix(prefix prefix, search_dir string) []match {
	entries, err := os.ReadDir(search_dir)
	if err != nil {
		return []match{}
	}

	matches := []match{}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), prefix.base) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		is_exec := !entry.IsDir() && info.Mode()&0111 != 0

		matches = append(matches, match{
			name:    entry.Name(),
			prefix:  prefix,
			is_dir:  entry.IsDir(),
			is_exec: is_exec,
		})
	}

	slices.SortFunc(matches, func(a, b match) int {
		return strings.Compare(a.name, b.name)
	})

	return matches
}

func parse_prefix(prefix_str string) prefix {
	idx := strings.LastIndexByte(prefix_str, '/')
	if idx == -1 {
		return prefix{
			dir:  "",
			base: prefix_str,
		}
	}
	return prefix{
		dir:  prefix_str[:idx+1],
		base: prefix_str[idx+1:],
	}
}

func longest_common_match_prefix(matches []match) string {
	lcp := strings.Builder{}
	for i := 0; ; i++ {
		var currChar byte
		for j, match := range matches {
			if i >= len(match.name) {
				return lcp.String()
			}
			if j == 0 {
				currChar = match.name[i]
			}
			if match.name[i] != currChar {
				return lcp.String()
			}
		}
		lcp.WriteByte(currChar)
	}
}

func join_matches(matches []match) string {
	matches_str := ""
	for _, match := range matches {
		matches_str += match.name
		if match.is_dir {
			matches_str += "/  "
		} else {
			matches_str += "  "
		}
	}
	return matches_str
}
