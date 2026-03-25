package main

import (
	"bufio"
	"fmt"
	"io"
)

type Line struct {
	CurrentLine    []byte
	PreviousLine   []byte
	CursorIndex    int
	HistoryIndex   int
	ShowAllMatches bool
}

func NewLine() *Line {
	return &Line{
		CurrentLine:    make([]byte, 0),
		PreviousLine:   make([]byte, 0),
		CursorIndex:    0,
		HistoryIndex:   NOT_IN_HISTORY,
		ShowAllMatches: false,
	}
}

func (l *Line) ToString() string {
	return string(l.CurrentLine)
}

func (l *Line) InsertByte(b byte) {
	if l.CursorIndex == len(l.CurrentLine) {
		l.CurrentLine = append(l.CurrentLine, b)
	} else {
		l.CurrentLine = append(l.CurrentLine, 0)
		copy(l.CurrentLine[l.CursorIndex+1:], l.CurrentLine[l.CursorIndex:])
		l.CurrentLine[l.CursorIndex] = b
	}

	l.CursorIndex++
	txtAfterCursor := l.CurrentLine[l.CursorIndex:]

	fmt.Printf("%c%s", b, txtAfterCursor)
	CursorBack(len(txtAfterCursor))
}

func (l *Line) DeleteByte() {
	if l.CursorIndex == 0 {
		fmt.Print("\a")
		return
	}

	if l.CursorIndex == len(l.CurrentLine) {
		l.CurrentLine = l.CurrentLine[:len(l.CurrentLine)-1]
	} else {
		copy(l.CurrentLine[l.CursorIndex-1:], l.CurrentLine[l.CursorIndex:])
		l.CurrentLine = l.CurrentLine[:len(l.CurrentLine)-1]
	}

	CursorBack(1)
	l.CursorIndex--
	txtAfterCursor := l.CurrentLine[l.CursorIndex:]

	fmt.Printf("%s%s", CLEAR_FROM_CURSOR, txtAfterCursor)
	CursorBack(len(txtAfterCursor))
}

func (l *Line) SetLine(line string) {
	CursorBack(l.CursorIndex)
	fmt.Print(CLEAR_FROM_CURSOR)

	l.CurrentLine = []byte(line)
	l.CursorIndex = len(l.CurrentLine)

	fmt.Printf("%s", l.CurrentLine)
}

func (l *Line) HandleArrowKey(cfg *Config) {
	sequence := ReadCSISequence(cfg.StdinReader)

	switch sequence {
	case CURSOR_UP:
		l.ArrowKeyUp(cfg)
	case CURSOR_DOWN:
		l.ArrowKeyDown(cfg)
	case CURSOR_FORWARD:
		l.ArrowKeyRight()
	case CURSOR_BACK:
		l.ArrowKeyLeft()
	}
}

func (l *Line) ArrowKeyUp(cfg *Config) {
	if l.HistoryIndex == NOT_IN_HISTORY && len(cfg.History) != 0 {
		l.HistoryIndex = len(cfg.History)
		l.PreviousLine = l.CurrentLine
	}

	if l.HistoryIndex > 0 {
		l.HistoryIndex--
		l.SetLine(cfg.History[l.HistoryIndex])
	} else {
		fmt.Print("\a")
	}
}

func (l *Line) ArrowKeyDown(cfg *Config) {
	if l.HistoryIndex == NOT_IN_HISTORY {
		fmt.Print("\a")
		return
	}

	if l.HistoryIndex < len(cfg.History)-1 {
		l.HistoryIndex++
		l.SetLine(cfg.History[l.HistoryIndex])

	} else {
		l.HistoryIndex = NOT_IN_HISTORY
		l.SetLine(string(l.PreviousLine))
	}
}

func (l *Line) ArrowKeyRight() {
	if l.CursorIndex == len(l.CurrentLine) {
		fmt.Print("\a")
		return
	}

	l.CursorIndex++
	CursorForward(1)
}

func (l *Line) ArrowKeyLeft() {
	if l.CursorIndex == 0 {
		fmt.Print("\a")
		return
	}

	l.CursorIndex--
	CursorBack(1)
}

func (l *Line) PerformTabCompletion(cfg *Config) {
	start, isCmd := l.GetPrefixStart()
	prefixStr := string(l.CurrentLine[start:l.CursorIndex])
	prefix := NewPrefix(prefixStr, isCmd)

	matches := AutoCompletePrefix(prefix)

	if len(matches) == 0 {
		fmt.Print("\a")
		return
	}
	if len(matches) == 1 {
		l.PrintMatch(start, matches[0])
		return
	}

	fmt.Print("\a")

	if lcp := LongestCommonPrefix(matches); lcp != prefix.PrefixBase {
		l.PrintPartialMatch(start, prefix.Directory, lcp)
	} else if l.ShowAllMatches {
		// if TAB pressed twice in sequence, print all matches on a new line
		fmt.Printf("\r\n%s\r\n%s%s", JoinMatches(matches), cfg.ShellPrompt(), l.CurrentLine)
	} else {
		l.ShowAllMatches = true
	}

}

func (l *Line) PrintMatch(prefixStart int, match Match) {
	before := string(l.CurrentLine[:prefixStart])
	after := string(l.CurrentLine[l.CursorIndex:])
	new_line := ""

	if match.IsDir {
		new_line = fmt.Sprintf("%s%s/%s", before, match.ToString(), after)
	} else {
		new_line = fmt.Sprintf("%s%s %s", before, match.ToString(), after)
	}

	l.SetLine(new_line)
	CursorBack(len(after))
	l.CursorIndex -= len(after)
}

func (l *Line) PrintPartialMatch(prefixStart int, dir, partial string) {
	before := string(l.CurrentLine[:prefixStart])
	after := string(l.CurrentLine[l.CursorIndex:])

	new_line := fmt.Sprintf("%s%s%s%s", before, dir, partial, after)

	l.SetLine(new_line)
	CursorBack(len(after))
	l.CursorIndex -= len(after)
}

// Returns the starting index of the substring that will be used
// to attempt to perform autocompletion when TAB is pressed. isCmd
// specifies if we are trying to autocomplete a command or a file.

func (l *Line) GetPrefixStart() (start int, isCmd bool) {
	prefixStart := l.CursorIndex
	for prefixStart >= 1 {
		if l.CurrentLine[prefixStart-1] == ' ' {
			break
		}
		prefixStart--
	}

	cmdStart := 0
	for cmdStart < len(l.CurrentLine) {
		if l.CurrentLine[cmdStart] != ' ' {
			break
		}
		cmdStart++
	}

	return prefixStart, cmdStart == prefixStart
}

func CursorBack(n int) {
	for range n {
		fmt.Print("\b")
	}
}

func CursorForward(n int) {
	for range n {
		fmt.Print(CURSOR_FORWARD)
	}
}

func ReadCSISequence(reader *bufio.Reader) string {
	char, err := reader.ReadByte()
	if char != '[' || err != nil {
		return ""
	}

	csi := []byte{ESC, '['}

	for {
		char, err := reader.ReadByte()
		if err != nil {
			return ""
		}

		csi = append(csi, char)

		// CSI sequences end with a char in the following range
		if '\x40' <= char && char <= '\x7E' {
			return string(csi)
		}
	}
}

func ReadLine(cfg *Config) (string, error) {
	line := NewLine()

	for {
		char, err := cfg.StdinReader.ReadByte()
		if err != nil {
			return "", nil
		}

		switch char {
		case '\r', '\n':
			return line.ToString(), nil
		case EOT:
			return "", io.EOF
		case DEL:
			line.DeleteByte()
		case '\t':
			line.PerformTabCompletion(cfg)
		case ESC:
			line.HandleArrowKey(cfg)
		default:
			if 32 <= char && char <= 126 {
				line.InsertByte(char)
			}
		}

		if char != '\t' {
			line.ShowAllMatches = false
		}
	}
}
