// Package process defines the process domain value objects.
package process

// Info is a snapshot of a single running process captured at discovery time.
type Info struct {
	Pid  int
	Name string
	Rss  int64
}
