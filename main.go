package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/bhargav-yarlagadda/goMon/watcher"
)

var currentCommand *exec.Cmd // Global variable to hold the current command being executed

// runApp restarts the application by killing the current running process and starting a new one.
func runApp(args []string) {
	// Kill current process if running
	if currentCommand != nil && currentCommand.Process != nil {
		fmt.Println("Restarting the server")
		currentCommand.Process.Kill()
	}
	// Prepare "go run ..." command using args[2:]
	goArgs := append([]string{"run"}, args[2:]...)
	currentCommand = exec.Command("go", goArgs...)
	currentCommand.Stdout = os.Stdout
	currentCommand.Stdin = os.Stdin
	currentCommand.Stderr = os.Stderr
	currentCommand.Env = os.Environ()

	// Run the new command asynchronously
	 
	go func() {
		if err := currentCommand.Run(); err != nil {
			fmt.Println("Error in restarting the server:", err)
		}
	}()
}

// main is the entry point of the application.
func main() {
	args := os.Args
	fmt.Println("Starting the server with arguments:", args)

	// Validate the format: gomon run <file.go> [args...]
	if len(args) < 3 || args[1] != "run" {
		fmt.Println("Usage: gomon run <your_file.go> [args...]")
		os.Exit(1)
	}

	// Extract the Go files to run (skip "gomon" and "run")
	filesToRun := args[2:]
	absPaths := []string{}

	// Convert relative file paths to absolute directories
	for _, f := range filesToRun {
		abs, err := filepath.Abs(f)
		if err != nil {
			fmt.Println("Error resolving path:", err)
			os.Exit(1)
		}
		absPaths = append(absPaths, filepath.Dir(abs))
	}

	// Start the app
	runApp(args)

	// Initialize the watcher
	w := watcher.New(absPaths, 1*time.Second, func(path string) {
		fmt.Printf(">>> Change detected: %s\n", path)
		runApp(args)
	})

	// Start watching
	if err := w.Start(); err != nil {
		fmt.Println("Watcher error:", err)
		os.Exit(1)
	}
}
