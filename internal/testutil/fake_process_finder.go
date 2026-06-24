package testutil

import "github.com/eirikur-ari/pidshooter/internal/process"

// FakeFinder is a test double for process.Finder.
type FakeFinder struct {
	Processes []process.Info
	Err       error
}

func (f *FakeFinder) List() ([]process.Info, error)           { return f.Processes, f.Err }
func (f *FakeFinder) Find(_ []string) ([]process.Info, error) { return f.Processes, f.Err }
