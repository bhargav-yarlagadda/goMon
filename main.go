package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

var currentCommand *exec.Cmd // Global variable to hold the current command being executed
var cmdMutex sync.Mutex      // Mutex to protect currentCommand

// killProcess function to kill the process in a goroutine
func killProcess(wg *sync.WaitGroup) {
	defer wg.Done() // Decrement the counter when the goroutine completes

	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	if currentCommand != nil && currentCommand.Process != nil {
		fmt.Println("Attempting to kill the current server process")
		err := currentCommand.Process.Kill()
		if err != nil {
			fmt.Println("Failed to kill the existing process:", err)
		} else {
			fmt.Println("Successfully killed the existing process")
		}
	}
}

// runApp restarts the application by killing the current running process and starting a new one.
func runApp(args []string, wg *sync.WaitGroup) {
	// Deploy a goroutine to handle killing the current process
	wg.Add(1) // Increment the counter before the goroutine starts
	go killProcess(wg)

	// Create a new command to run the server
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	currentCommand = exec.Command("go", args[1:]...)
	currentCommand.Stdout = os.Stdout
	currentCommand.Stdin = os.Stdin
	currentCommand.Stderr = os.Stderr
	currentCommand.Env = os.Environ()

	// Run the new command asynchronously
	wg.Add(1) // Increment the counter for the new goroutine
	go func(cmd *exec.Cmd, wg *sync.WaitGroup) {
		defer wg.Done() // Decrement the counter when the goroutine completes
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error in restarting the server:", err)
		}
	}(currentCommand, wg)
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

	// Create a WaitGroup to wait for the goroutines to finish
	var wg sync.WaitGroup

	// Start the app
	runApp(args, &wg)

	// Wait for all goroutines to finish before exiting the program
	wg.Wait()
}
