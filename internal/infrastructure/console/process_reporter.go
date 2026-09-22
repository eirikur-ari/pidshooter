package console

import (
	"fmt"
	"io"
	"os"
)

// ProcessReporter implements outbound.ProcessReporter by writing to an io.Writer.
type ProcessReporter struct {
	writer io.Writer
}

// NewProcessReporter returns a ProcessReporter that writes to os.Stdout.
func NewProcessReporter() *ProcessReporter {
	return &ProcessReporter{writer: os.Stdout}
}

// Report writes how many processes matched patterns, and that the game is starting.
func (r *ProcessReporter) Report(count int, patterns []string) {
	_, _ = fmt.Fprintf(r.writer, "Found %d process(es) matching %v. Starting game...\n", count, patterns)
}
