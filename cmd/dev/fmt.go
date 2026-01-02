package main

import (
	"os/exec"
	"sync"
)

func formatter(wg *sync.WaitGroup, cmds chan<- *exec.Cmd) {
	runCommand("FMT", colorCyan, []string{"/home/exedev/go/bin/air", "-c", ".air.fmt.toml"}, wg, cmds)
}
