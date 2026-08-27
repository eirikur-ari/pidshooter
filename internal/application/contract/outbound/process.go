package outbound

// ProcessInfo describes a single process discovered on the host.
type ProcessInfo struct {
	Pid  int
	Name string
	Rss  int64
}

// ProcessManager is the outbound port for process discovery and termination on the host.
type ProcessManager interface {
	// List returns the processes currently running on the host.
	List() ([]ProcessInfo, error)
	// OwnPid returns the PID of the calling process.
	OwnPid() int
	// LookupName returns the current name of the process with the given pid.
	LookupName(pid int) (string, error)
	// Kill terminates the process with the given pid, reporting whether it
	// was killed. It performs no safety or name verification itself —
	// callers must confirm via LookupName that pid still refers to the
	// intended process before calling Kill.
	Kill(pid int) (bool, error)
}
