package outbound

type ProcessInfo struct {
	Pid  int
	Name string
	Rss  int64
}

// Process is the outbound port for process discovery and termination on the host.
// Kill must verify the process still has the given name before sending the signal.
type Process interface {
	List() ([]ProcessInfo, error)
	Find(patterns []string) ([]ProcessInfo, error)
	Kill(pid int, name string) (bool, error)
}
