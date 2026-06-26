package driven

// Info represents a single running process as seen by the domain.
type Info interface {
	Pid() int
	Name() string
	Rss() int64
}
