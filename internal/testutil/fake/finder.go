package fake

import "github.com/eirikur-ari/pidshooter/internal/domain/process/ports/driven"

// Finder is a test double for driven.Finder.
type Finder struct {
	Processes []driven.Info
	Err       error
}

func (f *Finder) List() ([]driven.Info, error)           { return f.Processes, f.Err }
func (f *Finder) Find(_ []string) ([]driven.Info, error) { return f.Processes, f.Err }