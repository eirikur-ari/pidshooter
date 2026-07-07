package outbound

import "github.com/eirikur-ari/pidshooter/internal/core/process"

// ProcessFinder is the outbound port for process discovery on the host.
type ProcessFinder interface {
	List() ([]process.Info, error)
	Find(patterns []string) ([]process.Info, error)
}

// ProcessKiller is the outbound port for sending a kill signal to a process.
// name is the expected process name as discovered at game start; Kill must
// verify the process still has that name before sending the signal.
type ProcessKiller interface {
	Kill(pid int, name string) error
}
