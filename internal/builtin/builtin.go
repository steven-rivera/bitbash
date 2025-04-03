package builtin

import "os"

type BuiltIn = func(args []string, outFile *os.File, errFile *os.File) error

func GetBuiltInCommands() map[string]BuiltIn {
	return map[string]BuiltIn{
		"exit": HandlerExit,
		"echo": HandlerEcho,
		"type": HandlerType,
		"pwd":  HandlerPwd,
		"cd":   HandlerCd,
	}
}
