package outbound

type ProcessInfo struct {
	Pid  int
	Name string
	Rss  int64
}

// Lister discovers processes currently running on the host.
type Lister interface {
	List() ([]ProcessInfo, error)
	OwnPid() int
}

// Killer looks up a process's current name and terminates it by PID.
// Kill performs no safety or name verification itself — callers must confirm
// via LookupName that pid still refers to the intended process before calling Kill.
type Killer interface {
	LookupName(pid int) (string, error)
	Kill(pid int) (bool, error)
}

// Process is the outbound port for process discovery and termination on the host.
type Process interface {
	Lister
	Killer
}
