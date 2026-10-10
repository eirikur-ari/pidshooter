package console

import (
	"fmt"
	"io"
	"os"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// ProcessReporter reports the processes that matched the requested patterns.
type ProcessReporter struct {
	writer io.Writer
}

// NewProcessReporter returns a ProcessReporter that writes to standard output.
func NewProcessReporter() *ProcessReporter {
	return &ProcessReporter{writer: os.Stdout}
}

// Report writes the number of matches and the patterns that produced them.
func (r *ProcessReporter) Report(matches []outbound.ProcessInfo, patterns []string) {
	_, _ = fmt.Fprintf(r.writer, "Found %d process(es) matching %v.\n", len(matches), patterns)
}
