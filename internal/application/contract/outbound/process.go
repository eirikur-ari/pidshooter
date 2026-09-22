package outbound

// ProcessInfo describes a single process discovered on the host.
type ProcessInfo struct {
	UID   int
	PID   int
	Rss   int64
	State string
	Name  string
}

// ProcessHandle references a specific process obtained via Process.Pin,
// pinning its identity so a later Kill call cannot be redirected to a
// different process that has since reused the same PID.
type ProcessHandle interface {
	// Kill terminates the process this ProcessHandle refers to. If the
	// process no longer exists, Kill returns a NotFoundError instead of
	// treating it as a failure.
	Kill() error
	// Release releases any resources held by this ProcessHandle. Callers
	// must call Release exactly once when finished with the handle, whether
	// or not Kill was called.
	Release() error
}

// Process is the outbound port for process discovery and termination on the host.
type Process interface {
	// Discover returns the processes currently running on the host.
	Discover() ([]ProcessInfo, error)
	// OwnPID returns the PID of the calling process.
	OwnPID() int
	// OwnUID returns the effective UID of the calling process, used to
	// determine which discovered processes the caller is permitted to kill.
	OwnUID() int
	// LookupName returns the current name of the process with the given pid.
	// If the process no longer exists, LookupName returns a NotFoundError
	// instead of treating it as any other failure.
	LookupName(pid int) (string, error)
	// Pin returns a ProcessHandle to the process with the given pid. Callers
	// should Pin a pid before verifying it via LookupName and hold the
	// resulting ProcessHandle through to Kill, rather than re-resolving pid
	// at kill time, so identity is pinned across the verify-then-kill
	// sequence.
	Pin(pid int) (ProcessHandle, error)
}

// ProcessReporter is the outbound port for reporting how many processes
// matched the requested search patterns.
type ProcessReporter interface {
	// Report displays how many processes matched patterns.
	Report(count int, patterns []string)
}
