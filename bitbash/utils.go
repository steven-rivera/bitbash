package main

import "fmt"

const (
	EOT = byte(4)   // Sent when pressed  Ctrl+D
	DEL = byte(127) // Sent when pressed Backspace
	ESC = byte(27)

	ARROW_UP    = "[A"
	ARROW_DOWN  = "[B"
	ARROW_RIGHT = "[C"
	ARROW_LEFT  = "[D"

	RESET             = "\x1b[0m"
	BOLD              = "\x1b[1m"
	CURSOR_FORWARD    = "\x1b[1C"
	CLEAR_FROM_CURSOR = "\x1b[K"

	BLACK   = "\x1b[30m"
	RED     = "\x1b[31m"
	GREEN   = "\x1b[32m"
	YELLOW  = "\x1b[33m"
	BLUE    = "\x1b[34m"
	MAGENTA = "\x1b[35m"
	CYAN    = "\x1b[36m"
	WHITE   = "\x1b[37m"
)

func move_cursor_back(n int) {
	for range n {
		fmt.Print("\b")
	}
}

func move_cursor_forward(n int) {
	for range n {
		fmt.Print(CURSOR_FORWARD)
	}
}
