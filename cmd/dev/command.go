package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
)

func pipeOutput(prefix string, color string, r io.Reader, isError bool) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		if isError {
			fmt.Fprintf(os.Stderr, "%s[%s] %s%s\n", color, prefix, text, colorReset)
		} else {
			fmt.Printf("%s[%s] %s%s\n", color, prefix, text, colorReset)
		}
	}
}

func runCommand(name string, color string, args []string, wg *sync.WaitGroup, cmds chan<- *exec.Cmd) {
	defer wg.Done()

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(),
		"FORCE_COLOR=true",
		"COLORTERM=truecolor",
		"TERM=xterm-256color",
		"DEV=1")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s[%s] Failed to create stdout pipe: %v%s\n", color, name, err, colorReset)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s[%s] Failed to create stderr pipe: %v%s\n", color, name, err, colorReset)
		return
	}

	go pipeOutput(name, color, stdout, false)
	go pipeOutput(name, color, stderr, true)

	cmds <- cmd

	fmt.Printf("%s[%s] Starting process%s\n", color, name, colorReset)
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%s[%s] Failed to start: %v%s\n", color, name, err, colorReset)
		return
	}

	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "%s[%s] Process ended with error: %v%s\n", color, name, err, colorReset)
	} else {
		fmt.Printf("%s[%s] Process completed successfully%s\n", color, name, colorReset)
	}
}
