package util

import "fmt"

// FormatBytes formats a byte count as a human-readable size in binary units
// (1 KB is 1024 bytes), with one decimal place above bytes. A size that would
// round up to 1024 of one unit is shown in the next unit instead.
func FormatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	// Each threshold is the first size that would otherwise print as "1024.0"
	// of the smaller unit.
	switch {
	case bytes >= GB-MB/20:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB-KB/20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// ClonePointer returns a pointer to a shallow copy of the value pointer refers
// to, or nil if pointer is nil.
func ClonePointer[T any](pointer *T) *T {
	if pointer == nil {
		return nil
	}
	return new(*pointer)
}
