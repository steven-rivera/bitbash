package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

type Command struct {
	Name      string
	Args      []string
	IsBuiltin bool
	in        *os.File
	out       *os.File
	err       *os.File
	initErr   error
}

func (cmd *Command) Init(tokens []string) {
	for i := 0; i < len(tokens); {
		token := tokens[i]

		if _, ok := REDIRECTION_OPS[token]; ok {
			cmd.SetRedirect(token, tokens[i+1])

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

func (cmd *Command) SetRedirect(operator, fileName string) {
	flags := REDIRECTION_OPS[operator]

	file, err := os.OpenFile(fileName, flags, 0o666)
	if err != nil {
		cmd.initErr = err
		return
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
	defer wg.Done()
	defer cmd.ClosePipes()

	if cmd.initErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", cmd.initErr)
		return
	}

	if cmd.IsBuiltin {
		BUILTIN_CMDS[cmd.Name].Handler(cmd, cfg)
	} else {
		cmd.runExec()
	}
}

func (cmd *Command) runExec() {
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

type PipeLine struct {
	Commands []*Command
	Len      int
}

func NewPipeline(tokens []string) *PipeLine {
	splitTokens := splitOnPipes(tokens)

	pipeline := PipeLine{
		Commands: make([]*Command, 0, len(splitTokens)),
		Len:      len(splitTokens),
	}

	for range len(splitTokens) {
		pipeline.Commands = append(pipeline.Commands, &Command{})
	}

	pipeline.ConnectPipes()

	for i := range pipeline.Len {
		pipeline.Commands[i].Init(splitTokens[i])
	}

	return &pipeline
}

func (pl *PipeLine) ConnectPipes() {
	pl.Commands[0].in = os.Stdin

	for i := 0; i < len(pl.Commands)-1; i++ {
		r, w, _ := os.Pipe()

		pl.Commands[i+1].in = r
		pl.Commands[i].out = w
		pl.Commands[i].err = os.Stderr

	}

	pl.Commands[len(pl.Commands)-1].out = os.Stdout
	pl.Commands[len(pl.Commands)-1].err = os.Stderr
}

func (pl *PipeLine) Execute(cfg *Config) {
	var wg sync.WaitGroup
	wg.Add(pl.Len)

	for _, cmd := range pl.Commands {
		go cmd.Run(&wg, cfg)
	}

	wg.Wait()
}
