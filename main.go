package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
	"github.com/bhargav-yarlagadda/goMon/watcher"
)

var currentCommand *exec.Cmd   // Global variable to hold the current command being executed
var cmdMutex sync.Mutex        // Mutex to protect currentCommand

// killProcess kills the currently running process, if any
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

// startProcess starts a new process to run the Go application
func startProcess(args []string, wg *sync.WaitGroup) {
	defer wg.Done() // Decrement the counter when the goroutine completes

	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	// Create a new command to run the Go application
	currentCommand = exec.Command("go", args[1:]...)
	currentCommand.Stdout = os.Stdout
	currentCommand.Stdin = os.Stdin
	currentCommand.Stderr = os.Stderr
	currentCommand.Env = os.Environ()

	// Run the new command asynchronously
	go func() {
		err := currentCommand.Run()
		if err != nil {
			fmt.Println("Error in restarting the server:", err.Error())
		}
	}()
}


// runApp restarts the application by killing the current running process and starting a new one
func runApp(args []string, wg *sync.WaitGroup) {
	// Kill the current process and start a new one
	wg.Add(2) // Add two goroutines to the WaitGroup (one for killing and one for starting)

	// Kill the current process
	go killProcess(wg)

	// Start the new process
	go startProcess(args, wg)
}

// main is the entry point of the application
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

	// Initialize the watcher
	w := watcher.New([]string{"."}, 1*time.Second, func(path string) {
		fmt.Printf(">>> Change detected: %s\n", path)
		runApp(args, &wg)
	})

	// Start watching
	if err := w.Start(); err != nil {
		fmt.Println("Watcher error:", err)
		os.Exit(1)
	}

	// Wait for all goroutines to finish before exiting the program
	wg.Wait()
}
