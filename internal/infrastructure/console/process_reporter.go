package console

import (
	"fmt"
	"io"
	"os"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// ProcessReporter writes process-match reports to an io.Writer.
type ProcessReporter struct {
	writer io.Writer
}

// NewProcessReporter returns a ProcessReporter that writes to os.Stdout.
func NewProcessReporter() *ProcessReporter {
	return &ProcessReporter{writer: os.Stdout}
}

// Report writes the given matches and the patterns that produced them, and
// notes that the game is starting.
func (r *ProcessReporter) Report(matches []outbound.ProcessInfo, patterns []string) {
	_, _ = fmt.Fprintf(r.writer, "Found %d process(es) matching %v. Starting game...\n", len(matches), patterns)
}
