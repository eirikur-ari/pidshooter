package driven

import "github.com/eirikur-ari/pidshooter/internal/domain/process"

// Finder is the driven port for process discovery on the host.
type Finder interface {
	List() ([]process.Info, error)
	Find(patterns []string) ([]process.Info, error)
}
