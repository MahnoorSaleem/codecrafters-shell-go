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

	// out is where normal output goes: the screen, unless there's a > redirect.
	var out io.Writer = os.Stdout

	// Redirect stdout: "echo hi > file.txt" or "echo hi 1> file.txt"
	i := slices.Index(parts, ">")
	if i == -1 {
		i = slices.Index(parts, "1>")
	}
	if i != -1 {
		if i+1 >= len(parts) {
			fmt.Fprintln(os.Stderr, "syntax error: missing file name after >")
			return
		}
		fileName := parts[i+1]
		parts = parts[:i] // keep only the words before >

		f, err := os.Create(fileName) // creates the file, or empties it if it exists
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		defer f.Close()
		out = f

		if len(parts) == 0 {
			return // just "> file.txt": the file is created, nothing to run
		}
	}

	name, args := parts[0], parts[1:]

	switch name {
	case "echo":
		handleEcho(args, out)
	case "exit":
		handleExit(args)
	case "type":
		handleType(args, out)
	case "pwd":
		handlePwd(out)
	case "cd":
		handleCd(args)
	default:
		runExternal(name, args, out)
	}
}

func parseArgs(line string) []string {
	var args []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	inWord := false
	for _, ch := range line {
		switch {
		case escaped:
			if inDouble && ch != '"' && ch != '\\' {
				current.WriteRune('\\')
			}
			current.WriteRune(ch)
			inWord = true
			escaped = false
		case ch == '\\' && !inSingle:

			escaped = true
			inWord = true
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

func handleEcho(args []string, out io.Writer) {
	fmt.Fprintln(out, strings.Join(args, " "))
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

func handleType(args []string, out io.Writer) {
	if len(args) == 0 {
		fmt.Println("type: missing operand")
		return
	}

	target := args[0]
	if slices.Contains(builtins, target) {
		fmt.Fprintln(out, target+" is a shell builtin")
	} else if path, err := exec.LookPath(target); err == nil {
		fmt.Fprintf(out, "%s is %s\n", target, path)
	} else {
		fmt.Println(target + ": not found")
	}
}

func handlePwd(out io.Writer) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "pwd:", err)
		return
	}
	fmt.Fprintln(out, cwd)
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

func runExternal(name string, args []string, out io.Writer) {
	if _, err := exec.LookPath(name); err != nil {
		fmt.Println(name + ": command not found")
		return
	}

	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = out       // the screen, or the file after >
	cmd.Stderr = os.Stderr // errors always go to the screen

	var exitErr *exec.ExitError
	if err := cmd.Run(); err != nil && !errors.As(err, &exitErr) {
		fmt.Fprintln(os.Stderr, "Error executing command:", err)
	}
}
