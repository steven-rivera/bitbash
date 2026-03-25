# BitBash

**BitBash** is a lightweight Unix-like shell written in Go. It supports key features found in modern shells like Bash, including redirection, piping, autocompletion, quoting, and command history.

![logo](images/ascii-art.png)

## Features

### Redirection

- `<`:  Redirect `stdin` from a file
- `>`, `>>`: Redirect `stdout` to a file (overwrite or append)
- `2>`, `2>>`: Redirect `stderr` to a file (overwrite or append)
- `&>`, `&>>`: Redirect both `stdout` and `stderr` to a file (overwrite or append)

Ex:

```bash
$ cat < file.txt
$ echo "Hello World" > file.txt
$ cat doesnotexist.txt 2> err.txt
```

### Piping

- `|`: Pipe `stdout` of one command into `stdin` of another

Ex:

```bash
$ echo 'one two three' | tr ' ' '\n' | sort
```

### Autocomplete

- `<TAB>`: Attempt to complete or partially complete a command or file name
- `<TAB><TAB>`: If multiple matches exist, print all possible completions

Ex:

```bash
$ ec<TAB>
$ echo 
```

###  Quoting

- `'...'`: Preserve literal value of characters inside single quotes and disables word splitting
- `"..."`: Similar to single quotes, but supports escape sequences like `\\`, `\$`, and `\"`

Ex:

```bash
$ printf '%s\n' "hello world"
hello world
$ printf '%s\n' hello world
hello
world
```
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

## Installing

You can install BitBash by running the following command:

```bash
go install github.com/steven-rivera/bitbash/bitbash@latest
```

Once installed you can simply run the following:

```bash
bitbash
```

You should now be inside the BitBash shell!

## History

Command history can optionally be loaded from a file on startup and saved to the same file on exit. This allows history to persist between sessions of the Bitbash shell. BitBash will use the file specified in the `HISTFILE` environment variable to load and save command history. 

This repo provides an example history file `history.txt` containing a few commands. You can tell BitBash to use this file by running:

```bash
HISTFILE="history.txt" bitbash
```

You should then be able to browse the commands that were loaded from the file by pressing the up and down arrow keys.

## Requirements

- Go `>= 1.24`