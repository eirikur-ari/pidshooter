// Package testutil provides shared test doubles for use across packages.
package testutil

import "github.com/eirikur-ari/pidshooter/internal/domain/process/ports/driven"

type fakeProcessInfo struct {
	pid  int
	name string
	rss  int64
}

// NewFakeProcess returns a driven.Info test double with the given fields.
func NewFakeProcess(pid int, name string, rss int64) driven.Info {
	return &fakeProcessInfo{pid: pid, name: name, rss: rss}
}

func (f *fakeProcessInfo) Pid() int     { return f.pid }
func (f *fakeProcessInfo) Name() string { return f.name }
func (f *fakeProcessInfo) Rss() int64   { return f.rss }
