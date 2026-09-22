package console

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessReporterReportPrintsCountAndPatterns(t *testing.T) {
	var buf bytes.Buffer
	r := &ProcessReporter{writer: &buf}

	r.Report(3, []string{"chrome", "sleep"})

	assert.Contains(t, buf.String(), "Found 3 process(es)")
	assert.Contains(t, buf.String(), "[chrome sleep]")
}
