package outbound

// ProcessInfo describes a single process.
type ProcessInfo struct {
	// UID is the ID of the user that owns the process.
	UID int
	PID int
	// Rss is the process's resident memory, in bytes.
	Rss int64
	// State is the process's current state.
	State string
	Name  string
}

// ProcessHandle references one specific process. Kill acts on that process
// even if its PID has since been reused by another.
type ProcessHandle interface {
	// Kill terminates the referenced process. If the process no longer
	// exists, Kill returns a NotFoundError.
	Kill() error
	// Release frees the resources held by the handle.
	Release() error
}

// ProcessManager provides access to running processes.
type ProcessManager interface {
	// Discover returns the processes currently running.
	Discover() ([]ProcessInfo, error)
	// OwnPID returns the PID of the running program.
	OwnPID() int
	// OwnUID returns the effective user ID of the running program.
	OwnUID() int
	// LookupName returns the current name of the process with the given pid.
	// If the process no longer exists, LookupName returns a NotFoundError.
	LookupName(pid int) (string, error)
	// Pin returns a ProcessHandle to the process with the given pid, bound
	// to that process's identity at the time of the call.
	Pin(pid int) (ProcessHandle, error)
}

// ProcessReporter reports which processes matched a set of search patterns.
type ProcessReporter interface {
	// Report displays matches, the processes that matched patterns.
	Report(matches []ProcessInfo, patterns []string)
}
