package main

import (
	"fmt"
	"path/filepath"
	"os"
	"os/exec"
	"time"
	"github.com/bhargav-yarlagadda/goMon/watcher"
)

// runApp restarts the application by killing the current running process and starting a new one.
func runApp(args []string) {
	// Check if there's an existing process and kill it before starting a new one.
	if currentCommand != nil && currentCommand.Process != nil {
		fmt.Println("Restarting the server")
		currentCommand.Process.Kill() // Kill the current server process
	}

	// Create a new command to run the server with the provided arguments.
	currentCommand = exec.Command("go", args...)
	// Set the standard output, input, and error for the command to be the same as the terminal.
	currentCommand.Stdout = os.Stdout
	currentCommand.Stdin = os.Stdin
	currentCommand.Stderr = os.Stderr
	// Inherit environment variables from the current process.
	currentCommand.Env = os.Environ()

	// Run the new command asynchronously.
	go func() {
		// Run the command and handle any errors that occur.
		if err := currentCommand.Run(); err != nil {
			fmt.Println("Error in restarting the server: ", err)
		}
	}()
}

var currentCommand *exec.Cmd // Global variable to hold the current command being executed

// main is the entry point of the application.
func main() {
	// Parse the arguments passed to the program
	args := os.Args[1:]

	// Validate the arguments to ensure that the command is in the correct format.
	if len(args) < 2 || args[0] != "run" {
		fmt.Println("Usage: gomon run <your_file.go> [args...]")
		os.Exit(1)
	}

	// Extract the file names to run. This skips the "run" argument and gets the remaining files.
	filesToRun := args[2:]
	absPaths := []string{}

	// Convert relative file paths to absolute paths to ensure the watcher is monitoring the correct directories.
	for _, f := range filesToRun {
		abs, err := filepath.Abs(f) // Get the absolute path of each file
		if err != nil {
			fmt.Println("Error resolving path:", err)
			os.Exit(1)
		}
		absPaths = append(absPaths, filepath.Dir(abs)) // Append the directory part of the absolute path
	}

	// Start the application by calling runApp with the arguments.
	runApp(args)

	// Initialize the watcher with the directories to watch and the polling interval.
	w := watcher.New(absPaths, 1*time.Second, func(path string) {
		// When a change is detected, print the path and restart the app.
		fmt.Printf(">>> Change detected: %s\n", path)
		runApp(args) // Restart the app with the same arguments
	})

	// Start the watcher to monitor file changes.
	if err := w.Start(); err != nil {
		// Handle any errors that occur during the watcher startup.
		fmt.Println("Watcher error:", err)
		os.Exit(1)
	}
}
