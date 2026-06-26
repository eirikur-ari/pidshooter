package testutil

import "github.com/eirikur-ari/pidshooter/internal/domain/process/ports/driven"

// FakeFinder is a test double for driven.Finder.
type FakeFinder struct {
	Processes []driven.Info
	Err       error
}

func (f *FakeFinder) List() ([]driven.Info, error)           { return f.Processes, f.Err }
func (f *FakeFinder) Find(_ []string) ([]driven.Info, error) { return f.Processes, f.Err }
