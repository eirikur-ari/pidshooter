package fake

import "github.com/eirikur-ari/pidshooter/internal/domain/process"

// Finder is a test double for driven.Finder.
type Finder struct {
	Processes []process.Info
	Err       error
}

func (f *Finder) List() ([]process.Info, error)           { return f.Processes, f.Err }
func (f *Finder) Find(_ []string) ([]process.Info, error) { return f.Processes, f.Err }
