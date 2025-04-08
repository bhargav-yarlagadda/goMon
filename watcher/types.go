package watcher

import "time"

type FileMeta struct {
	Path    string;
	ModTime time.Time;
}
type Watcher struct{
	Paths [] string;
	Files map[string]FileMeta;
	PollInterval time.Duration;
	onChange func(path string);
}