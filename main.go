package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bhargav-yarlagadda/goMon/watcher"
)

var currentCommand *exec.Cmd
var cmdMutex sync.Mutex

var debounceTimer *time.Timer
var debounceMutex sync.Mutex

func killProcess() {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	if currentCommand != nil && currentCommand.Process != nil {
		fmt.Println("Killing the current server process...")
		_ = currentCommand.Process.Kill()
	}
}

func startProcess(args []string) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	currentCommand = exec.Command("go", args[1:]...)
	currentCommand.Stdout = os.Stdout
	currentCommand.Stdin = os.Stdin
	currentCommand.Stderr = os.Stderr
	currentCommand.Env = os.Environ()

	go func() {
		if err := currentCommand.Run(); err != nil {
			fmt.Println("Error in restarting the server:", err.Error())
		}
	}()
}

func runApp(args []string) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		killProcess()
	}()

	go func() {
		defer wg.Done()
		startProcess(args)
	}()

	wg.Wait()
}

func debounceRestart(args []string, path string) {
	debounceMutex.Lock()
	defer debounceMutex.Unlock()

	if debounceTimer != nil {
		debounceTimer.Stop()
	}

	// Debounce interval: 500ms
	debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
		fmt.Printf(">>> Change detected: %s\n", path)
		runApp(args)
	})
}

func main() {
	args := os.Args
	fmt.Println("Starting the server with arguments:", args)

	if len(args) < 3 || args[1] != "run" {
		fmt.Println("Usage: gomon run <your_file.go> [args...]")
		os.Exit(1)
	}

	runApp(args)

	// Watcher with debounce logic
	w := watcher.New([]string{"."}, 1*time.Second, func(path string) {
		if strings.HasSuffix(path, ".go") {
			normalized := filepath.ToSlash(path)
			debounceRestart(args, normalized)
		}
	})

	if err := w.Start(); err != nil {
		fmt.Println("Watcher error:", err)
		os.Exit(1)
	}
}


