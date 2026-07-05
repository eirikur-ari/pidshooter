// Package process defines the process domain value objects.
package process

// MinPatternLength is the minimum allowed length for a process search pattern.
// Single- or double-character patterns match too broadly (e.g. "a" matches
// most system process names) and increase the risk of surfacing critical
// system processes as kill targets.
const MinPatternLength = 3

// MaxPatternLength is the maximum allowed length for a process search pattern.
const MaxPatternLength = 256

// Info is a snapshot of a single running process captured at discovery time.
type Info struct {
	Pid  int
	Name string
	Rss  int64
}
