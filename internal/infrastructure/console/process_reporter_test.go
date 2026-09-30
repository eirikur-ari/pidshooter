package console

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestProcessReporterReportPrintsCountAndPatterns(t *testing.T) {
	var buf bytes.Buffer
	r := &ProcessReporter{writer: &buf}

	r.Report([]outbound.ProcessInfo{{PID: 1}, {PID: 2}, {PID: 3}}, []string{"chrome", "sleep"})

	assert.Contains(t, buf.String(), "Found 3 process(es)")
	assert.Contains(t, buf.String(), "[chrome sleep]")
}
