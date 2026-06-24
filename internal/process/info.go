package process

// process holds the raw data for a single running process.
type process struct {
	pid  int
	name string
	rss  int64 // resident set size in bytes
}

// Info represents a running process.
type Info interface {
	Pid() int     // operating system process ID
	Name() string // short process name (no path)
	Rss() int64   // resident set size in bytes
}

func (p *process) Pid() int     { return p.pid }
func (p *process) Name() string { return p.name }
func (p *process) Rss() int64   { return p.rss }
