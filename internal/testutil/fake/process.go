package fake

import "github.com/eirikur-ari/pidshooter/internal/domain/process/ports/driven"

type processInfo struct {
	pid  int
	name string
	rss  int64
}

// NewProcess returns a driven.Info test double with the given fields.
func NewProcess(pid int, name string, rss int64) driven.Info {
	return &processInfo{pid: pid, name: name, rss: rss}
}

func (f *processInfo) Pid() int     { return f.pid }
func (f *processInfo) Name() string { return f.name }
func (f *processInfo) Rss() int64   { return f.rss }