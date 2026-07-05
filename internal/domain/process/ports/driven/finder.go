package driven

import "github.com/eirikur-ari/pidshooter/internal/domain/process"

// MinPatternLength is the minimum allowed length for a process search pattern.
// Single- or double-character patterns match too broadly (e.g. "a" matches
// most system process names) and increase the risk of surfacing critical
// system processes as kill targets.
const MinPatternLength = 3

// MaxPatternLength is the maximum allowed length for a process search pattern.
const MaxPatternLength = 256

// Finder is the driven port for process discovery on the host.
type Finder interface {
	List() ([]process.Info, error)
	Find(patterns []string) ([]process.Info, error)
}
