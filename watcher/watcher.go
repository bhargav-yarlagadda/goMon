package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// New creates a new Watcher instance.
// Takes a list of directories to monitor, polling interval, and a callback function.
func New(paths []string, interval time.Duration, onChange func(path string)) *Watcher {
	return &Watcher{
		Paths:        paths,
		PollInterval: interval,
		Files:        make(map[string]FileMeta),
		onChange:     onChange,
	}
}

// Start begins the file watcher.
// Continuously checks for file changes and triggers the onChange callback.
func (w *Watcher) Start() error {
	// Initial scan
	if err := w.scan(); err != nil {
		return err
	}

	// Polling loop
	for {
		time.Sleep(w.PollInterval)

		changes, err := w.detectChanges()
		if err != nil {
			return err
		}

		for _, path := range changes {
			if w.onChange != nil {
				w.onChange(path)
			}
		}
	}
}

// scan populates the initial state of .go files in the specified paths.
func (w *Watcher) scan() error {
	for _, root := range w.Paths {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".go") {
				w.Files[path] = FileMeta{
					Path:    path,
					ModTime: info.ModTime(),
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// detectChanges finds newly created, modified, or deleted .go files.
func (w *Watcher) detectChanges() ([]string, error) {
	changed := []string{}
	seen := make(map[string]bool)

	// Copy of the current known files to detect deletions
	originalFiles := make(map[string]FileMeta)
	for path, meta := range w.Files {
		originalFiles[path] = meta
	}

	// Walk directories again and check for changes
	for _, root := range w.Paths {
		// filepath.Walk traverses every directory and file and calls the callback on each file
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			// os.FileInfo is an interface from Go’s os package that provides metadata about a file or directory.

			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".go") {
				seen[path] = true
				prev, exists := w.Files[path]

				// New file or modified file
				if !exists || !prev.ModTime.Equal(info.ModTime()) {
					w.Files[path] = FileMeta{
						Path:    path,
						ModTime: info.ModTime(),
					}
					changed = append(changed, path)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// Check for deleted files
	for path := range originalFiles {
		if !seen[path] {
			delete(w.Files, path) // Remove from tracked files
			changed = append(changed, path) // Treat as a change
		}
	}

	return changed, nil
}
