// Package testutil provides shared test helpers for use across packages.
package testutil

import "github.com/eirikur-ari/pidshooter/internal/process"

// TODO: is this the correct way to provide fake objects?
type fakeProcessInfo struct {
	pid  int
	name string
	rss  int64
}

// NewFakeProcess returns a process.Info test double with the given fields.
func NewFakeProcess(pid int, name string, rss int64) process.Info {
	return &fakeProcessInfo{pid: pid, name: name, rss: rss}
}

func (f *fakeProcessInfo) Pid() int     { return f.pid }
func (f *fakeProcessInfo) Name() string { return f.name }
func (f *fakeProcessInfo) Rss() int64   { return f.rss }
