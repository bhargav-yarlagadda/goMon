package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bhargav-yarlagadda/goMon/watcher"
)

var (
	currentCommand *exec.Cmd
	cmdMutex       sync.Mutex

	debounceTimer *time.Timer
	debounceMutex sync.Mutex
)

// killProcess tries to gracefully stop the current process
func killProcess() {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	if currentCommand != nil && currentCommand.Process != nil {
		fmt.Println("Terminating the current server process...")

		// Graceful kill: SIGINT on Unix, Kill on Windows
		if runtime.GOOS == "windows" {
			if err := currentCommand.Process.Kill(); err != nil {
				fmt.Println("Error killing process:", err)
			}
		} else {
			if err := currentCommand.Process.Signal(syscall.SIGINT); err != nil {
				fmt.Println("Error sending SIGINT:", err)
			}
		}

		done := make(chan error, 1)
		go func() {
			done <- currentCommand.Wait()
		}()

		select {
		case err := <-done:
			if err != nil {
				fmt.Println("Process exited with error:", err)
			} else {
				fmt.Println("Process exited cleanly")
			}
		case <-time.After(5 * time.Second):
			fmt.Println("Timeout waiting for process to exit, killing forcibly")
			_ = currentCommand.Process.Kill()
		}

		currentCommand = nil
	}
}

// buildBinary compiles the Go binary to a temporary file and returns its path
func buildBinary(entryFile string) (string, error) {
	tmpBinary := filepath.Join(os.TempDir(), "gomon_temp_binary")

	// On Windows, executable must have .exe extension
	if runtime.GOOS == "windows" {
		tmpBinary += ".exe"
	}

	args := []string{"build", "-o", tmpBinary, entryFile}
	fmt.Printf("Building binary: go %s\n", strings.Join(args, " "))
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return tmpBinary, nil
}

func startProcess(binaryPath string, args []string) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	allArgs := args[3:] // skip program name, "run", and file path
	fmt.Printf("Starting process: %s %s\n", binaryPath, strings.Join(allArgs, " "))

	currentCommand = exec.Command(binaryPath, allArgs...)
	currentCommand.Stdout = os.Stdout
	currentCommand.Stderr = os.Stderr
	currentCommand.Stdin = os.Stdin
	currentCommand.Env = os.Environ()

	// Retry starting up to 3 times if port in use error occurs
	for attempt := 1; attempt <= 3; attempt++ {
		err := currentCommand.Start()
		if err != nil {
			fmt.Println("Error starting process:", err)
			return
		}

		errChan := make(chan error, 1)
		go func() {
			errChan <- currentCommand.Wait()
		}()

		select {
		case err := <-errChan:
			if err != nil && strings.Contains(err.Error(), "bind: Only one usage of each socket address") {
				fmt.Printf("Port in use, retrying (%d/3)...\n", attempt)
				_ = currentCommand.Process.Kill()
				time.Sleep(1 * time.Second)
				continue
			}
			if err != nil {
				fmt.Println("Process exited with error:", err)
			} else {
				fmt.Println("Process exited normally")
			}
			return
		case <-time.After(2 * time.Second):
			// Process started successfully
			return
		}
	}

	fmt.Println("Failed to start process after 3 attempts")
	currentCommand = nil
}

func runApp(args []string) {
	cmdMutex.Lock()
	entryFile := ""
	if len(args) > 2 {
		entryFile = args[2]
	}
	cmdMutex.Unlock()

	// Kill existing process
	killProcess()

	// Build new binary
	binaryPath, err := buildBinary(entryFile)
	if err != nil {
		fmt.Println("Build failed:", err)
		return
	}

	// Small delay to ensure port is freed
	time.Sleep(1 * time.Second)

	// Start new process
	startProcess(binaryPath, args)
}

func debounceRestart(args []string, path string) {
	debounceMutex.Lock()
	defer debounceMutex.Unlock()

	if debounceTimer != nil {
		debounceTimer.Stop()
	}

	debounceTimer = time.AfterFunc(1*time.Second, func() {
		fmt.Printf(">>> Change detected: %s\n", path)
		runApp(args)
	})
}

func main() {
	args := os.Args

	if len(args) < 2 || args[1] == "--help" {
		fmt.Print(`
Gomon - A Go hot-reload tool for seamless development

Usage:
  gomon run <your_file.go> [args...]

Description:
  Watches for changes in .go files in the current directory and restarts the app automatically.

Example:
  gomon run main.go
  gomon run server/server.go

Note:
  This tool uses polling and may be inefficient for large directories.
`)
		os.Exit(0)
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
		// Ignore temporary or non-Go files
		if !strings.HasSuffix(path, ".go") || strings.Contains(path, ".swp") || strings.Contains(path, "~") {
			return
		}
		normalized := filepath.ToSlash(path)
		debounceRestart(args, normalized)
	})

	if err := w.Start(); err != nil {
		fmt.Println("Watcher error:", err)
		os.Exit(1)
	}
}
