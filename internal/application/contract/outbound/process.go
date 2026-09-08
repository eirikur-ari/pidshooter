package outbound

// ProcessInfo describes a single process discovered on the host.
type ProcessInfo struct {
	PID  int
	Name string
	Rss  int64
	UID  int
}

// ProcessHandle references a specific process obtained via
// ProcessManager.Pin, pinning its identity so a later Kill call cannot be
// redirected to a different process that has since reused the same PID.
type ProcessHandle interface {
	// Kill terminates the process this ProcessHandle refers to. If the
	// process no longer exists, Kill returns a NotFoundError instead of
	// treating it as a failure.
	Kill() error
}

// ProcessManager is the outbound port for process discovery and termination on the host.
type ProcessManager interface {
	// List returns the processes currently running on the host.
	List() ([]ProcessInfo, error)
	// OwnPID returns the PID of the calling process.
	OwnPID() int
	// OwnUID returns the effective UID of the calling process, used to
	// determine which discovered processes the caller is permitted to kill.
	OwnUID() int
	// LookupName returns the current name of the process with the given pid.
	LookupName(pid int) (string, error)
	// Pin returns a ProcessHandle to the process with the given pid. Callers
	// should Pin a pid before verifying it via LookupName and hold the
	// resulting ProcessHandle through to Kill, rather than re-resolving pid
	// at kill time, so identity is pinned across the verify-then-kill
	// sequence.
	Pin(pid int) (ProcessHandle, error)
}
