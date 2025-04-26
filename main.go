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

	// Debounce interval: 500ms,debounce is actually cancelling previous call  need to merge calls.
	debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
		fmt.Printf(">>> Change detected: %s\n", path)
		runApp(args)
	})
}

func main() {
	args := os.Args

	if args[1]=="--help" || args[1] == "" {
		fmt.Print(`
Gomon - A Go hot-reload tool for seamless development

Usage:
  gomon run <your_file.go> [args...]


Description:
  Watches for changes in go  files in the current directory and restarts the app automatically.

Example:
  gomon run main.go
  gomon run server/server.go

Note:
  This tool is currently using polling to detect file changes and may be inefficient for large directories.
`)
	}
	 if len(args) < 3 || args[1] != "run" {
		fmt.Print(`				
		Invalid command.
		Usage:
		gomon run <your_file.go> 

		Description:
		Starts your Go application and watches for changes in .go files.
		On any change, the application will be rebuilt and restarted automatically.

		Example:
		gomon run server/main.go
		gomon run cmd/app.go --port=8080

		Tip:
		Use this tool during development to avoid restarting your server manually.
		`)

		os.Exit(1)
	}

	fmt.Println("Starting the server with arguments:", args)
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


