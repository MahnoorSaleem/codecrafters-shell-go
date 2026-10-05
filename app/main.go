package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {

	reader := bufio.NewReader(os.Stdin)
	builtins := []string{"echo", "exit", "type"}

	for {

		fmt.Print("$ ")

		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		command = strings.TrimSpace(command)  // echo hello world
		tokens := strings.Split(command, " ") // []string{"echo", "hello", "world"}
		baseCmd := tokens[0]                  // echo (2) type

		if baseCmd == "type" {
			if len(tokens) < 2 {
				fmt.Println("type: missing operand")
				continue
			}
			target := tokens[1]
			if slices.Contains(builtins, target) {
				fmt.Println(target + " is a shell builtin")
			} else if path, err := exec.LookPath(target); err == nil {
				fmt.Printf("%s is %s\n", target, path)
			} else if baseCmd == "type" {
				fmt.Println(target + ": not found")
			}
		} else if command == "pwd" {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "pwd error:", err)
			}
			fmt.Println(cwd)

		} else if command == "exit" {
			os.Exit(0) // terminates the program immediately.
		} else if strings.HasPrefix(command, "echo ") {
			fmt.Println(command[5:])
		} else {

			_, err := exec.LookPath(tokens[0])

			if err != nil {
				fmt.Println(command + ": command not found")
				continue
			}

			cmd := exec.Command(tokens[0], tokens[1:]...)

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err = cmd.Run()

			if err != nil {
				fmt.Println("Error executing command:", err)
			}
		}

	}
}
