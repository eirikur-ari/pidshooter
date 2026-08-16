// Package fixture provides reusable test data builders for domain types.
package fixture

import "github.com/eirikur-ari/pidshooter/internal/core/process"

// Process returns a process.Info with the given pid and name, and zero RSS.
func Process(pid int, name string) process.Info {
	return process.NewInfo(pid, name, 0)
}

// Processes returns n process.Info fixtures with sequential PIDs starting at 1
// and letter names starting at "a".
func Processes(n int) []process.Info {
	infos := make([]process.Info, n)
	for i := range infos {
		infos[i] = Process(i+1, string(rune('a'+i)))
	}
	return infos
}
