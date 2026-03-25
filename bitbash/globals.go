package main

const (
	EOT = byte(4)   // Sent when pressed  Ctrl+D
	DEL = byte(127) // Sent when pressed Backspace
	ESC = byte(27)

	RESET             = "\x1b[0m"
	BOLD              = "\x1b[1m"
	CURSOR_UP         = "\x1b[A"
	CURSOR_DOWN       = "\x1b[B"
	CURSOR_FORWARD    = "\x1b[C"
	CURSOR_BACK       = "\x1b[D"
	CLEAR_FROM_CURSOR = "\x1b[K"

	BLACK   = "\x1b[30m"
	RED     = "\x1b[31m"
	GREEN   = "\x1b[32m"
	YELLOW  = "\x1b[33m"
	BLUE    = "\x1b[34m"
	MAGENTA = "\x1b[35m"
	CYAN    = "\x1b[36m"
	WHITE   = "\x1b[37m"

	NOT_IN_HISTORY = -1
)

var REDIRECTION_OPS = map[string]bool{
	"<":   true,
	">":   true,
	"1>":  true,
	"2>":  true,
	"&>":  true,
	">>":  true,
	"1>>": true,
	"2>>": true,
	"&>>": true,
}

var BUILTIN_CMDS map[string]BuiltInCommand
