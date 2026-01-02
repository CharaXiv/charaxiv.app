package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// Signal channel
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Command channel to track processes
	cmds := make(chan *exec.Cmd, 4)
	var wg sync.WaitGroup
	wg.Add(4)

	// Start processes
	go reloader(&wg, sigCh)
	go formatter(&wg, cmds)
	go templ(&wg, cmds)
	go server(&wg, cmds)

	// Track child processes
	processes := []*exec.Cmd{}
	go func() {
		for cmd := range cmds {
			processes = append(processes, cmd)
		}
	}()

	fmt.Println("Development server started. Press Ctrl+C to stop.")

	// Wait for signal then stop child processes
	<-sigCh
	fmt.Println("\nReceived shutdown signal. Stopping all processes...")

	for _, proc := range processes {
		if proc != nil && proc.Process != nil {
			proc.Process.Kill()
		}
	}

	close(cmds)
	wg.Wait()
	fmt.Println("All processes stopped.")
}
