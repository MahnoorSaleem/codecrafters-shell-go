package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

// builtins is the list of commands our shell handles itself.
var builtins = []string{"echo", "exit", "type", "pwd", "cd"}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("$ ")

		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				os.Exit(0) // Ctrl+D
			}
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			continue
		}

		runCommand(line)
	}
}

func runCommand(line string) {
	parts := parseArgs(line) // "echo 'hello   world'" -> ["echo", "hello   world"]
	if len(parts) == 0 {
		return // empty line
	}

	name, args := parts[0], parts[1:]

	switch name {
	case "echo":
		handleEcho(args)
	case "exit":
		handleExit(args)
	case "type":
		handleType(args)
	case "pwd":
		handlePwd()
	case "cd":
		handleCd(args)
	default:
		runExternal(name, args)
	}
}

// echo "hello    world"
// hello    world

func parseArgs(line string) []string { // echo 'shell hello'
	var args []string
	var current strings.Builder
	inSingle := false
	inDouble := false

	inWord := false
	for _, ch := range line {
		switch {
		case ch == '"' && !inSingle:
			inDouble = !inDouble
			inWord = true
		case ch == '\'' && !inDouble:
			inSingle = !inSingle
			inWord = true
		case (!inSingle && !inDouble) && (ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'):
			if inWord {
				args = append(args, current.String())
				current.Reset()
				inWord = false
			}
		default:
			current.WriteRune(ch)
			inWord = true
		}
	}
	if inWord {
		args = append(args, current.String())
	}
	return args
}

func handleEcho(args []string) {
	fmt.Println(strings.Join(args, " "))
}

func handleExit(args []string) {
	code := 0
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil {
			code = n
		}
	}
	os.Exit(code)
}

func handleType(args []string) {
	if len(args) == 0 {
		fmt.Println("type: missing operand")
		return
	}

	target := args[0]
	if slices.Contains(builtins, target) {
		fmt.Println(target + " is a shell builtin")
	} else if path, err := exec.LookPath(target); err == nil {
		fmt.Printf("%s is %s\n", target, path)
	} else {
		fmt.Println(target + ": not found")
	}
}

func handlePwd() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "pwd:", err)
		return
	}
	fmt.Println(cwd)
}

func handleCd(args []string) {
	path := "~"
	if len(args) > 0 {
		path = args[0]
	}

	path = expandHome(path)
	if err := os.Chdir(path); err != nil {
		fmt.Println("cd: " + path + ": No such file or directory")
	}
}

func expandHome(path string) string {
	home := os.Getenv("HOME")
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return home + path[1:]
	}
	return path
}

func runExternal(name string, args []string) {
	if _, err := exec.LookPath(name); err != nil {
		fmt.Println(name + ": command not found")
		return
	}

	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var exitErr *exec.ExitError
	if err := cmd.Run(); err != nil && !errors.As(err, &exitErr) {
		fmt.Fprintln(os.Stderr, "Error executing command:", err)
	}
}
