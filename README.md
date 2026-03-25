# BitBash

**BitBash** is a lightweight Unix-like shell written in Go. It supports key features found in modern shells like Bash, including redirection, piping, autocompletion, quoting, and command history.

![logo](images/ascii-art.png)

## Features

### Redirection

- `<`:  Redirect `stdin` from a file
- `>`, `>>`: Redirect `stdout` to a file (overwrite or append)
- `2>`, `2>>`: Redirect `stderr` to a file (overwrite or append)
- `&>`, `&>>`: Redirect both `stdout` and `stderr` to a file (overwrite or append)

### Piping

- `|`: Pipe `stdout` of one command into `stdin` of another

### Autocomplete

- `<TAB>`: Attempt to complete or partially complete a command or file name
- `<TAB><TAB>`: If multiple matches exist, print all possible completions

###  Quoting

- `'...'`: Preserve literal value of characters inside single quotes
- `"..."`: Similar to single quotes, but supports escape sequences like `\\`, `\$`, and `\"`

###  Command History

- `↑`: Browse to the previous command
- `↓`: Browse to the next command

## Builtin Commands

Bitbash comes with the following builtin commands:

- `cd`: Changes the current working directory
- `echo`: Print all arguments to `stdout`
- `exit`: Exit the shell with the provided code. Default `0`
- `help`: Prints more detailed information about builtin commands
- `history`: Prints previously executed commands
- `pwd`: Prints the current working directory
- `type`: Provide information about a command

## History

Command history can optionally be saved and loaded from a file. Bitbash will load/save command history on startup/exit from the file specified in the `HISTFILE` environment variable. This allows history to persist between sessions of the Bitbash shell.

- `history.txt`: sample history file

## Installing

You can installing BitBash by running the following command:

```bash
go install github.com/steven-rivera/bitbash/bitbash@latest
```

Once installed you can simply run:

```bash
HISTFILE="history.txt" bitbash
```

You should now be inside the BitBash shell!

## Requirements

- Go `>= 1.24`