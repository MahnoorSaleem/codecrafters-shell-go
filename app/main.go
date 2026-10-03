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

		command = strings.TrimSpace(command)
		tokens := strings.Split(command, " ")
		baseCmd := tokens[0]

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
		} else if command == "exit" {
			os.Exit(0)
		} else if strings.HasPrefix(command, "echo ") {
			fmt.Println(command[5:])
		} else {
			fmt.Println(command + ": command not found")
		}

	}
}
