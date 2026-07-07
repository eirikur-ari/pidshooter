package outbound

import "github.com/eirikur-ari/pidshooter/internal/core/process"

// Process is the outbound port for process discovery and termination on the host.
// Kill must verify the process still has the given name before sending the signal.
type Process interface {
	List() ([]process.Info, error)
	Find(patterns []string) ([]process.Info, error)
	Kill(pid int, name string) (bool, error)
}
