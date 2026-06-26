package driven

// MaxPatternLength is the maximum allowed length for a process search pattern.
const MaxPatternLength = 256

// Finder is the driven port for process discovery on the host.
type Finder interface {
	List() ([]Info, error)
	Find(patterns []string) ([]Info, error)
}
