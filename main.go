package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bhargav-yarlagadda/goMon/watcher"
)

func main() {
	// Watch current directory recursively
	fmt.Print("goMON started successfully.")
	w := watcher.New(
		[]string{"."},
		2*time.Second, // poll every 2 seconds
		func(path string) {
			fmt.Println("🟡 Detected change in:", path)
		},
	)

	if err := w.Start(); err != nil {
		log.Fatal("Watcher error:", err)
	}
}
