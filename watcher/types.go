package watcher

import "time"

// FileMeta represents metadata for a file, including its path and last modification time.
type FileMeta struct {
	// Path is the file's full path in the system.
	Path string
	// ModTime is the last modification time of the file.
	ModTime time.Time
}

// Watcher monitors a set of file paths for changes and triggers a callback when changes are detected.
type Watcher struct {
	// Paths holds a list of file paths to be monitored.
	Paths []string

	// Files is a map where the key is the file path and the value is the FileMeta data for each file.
	Files map[string]FileMeta

	// PollInterval specifies how often the watcher checks for file changes.
	PollInterval time.Duration

	// onChange is a callback function that gets triggered when a change is detected in any of the monitored files.
	onChange func(path string)
}
