package console

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestProcessReporter_Report_PrintsMatchCountAndPatterns(t *testing.T) {
	tests := newProcessReporterTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			var output bytes.Buffer
			reporter := &ProcessReporter{writer: &output}

			// When
			reporter.Report(test.matches, test.patterns)
			actual := output.String()

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func newProcessReporterTestCases() []struct {
	name     string
	matches  []outbound.ProcessInfo
	patterns []string
	expected string
} {
	return []struct {
		name     string
		matches  []outbound.ProcessInfo
		patterns []string
		expected string
	}{
		{
			"several matches and patterns",
			[]outbound.ProcessInfo{{PID: 1}, {PID: 2}, {PID: 3}},
			[]string{"chrome", "sleep"},
			"Found 3 process(es) matching [chrome sleep].\n",
		},
		{"no matches", nil, []string{"chrome"}, "Found 0 process(es) matching [chrome].\n"},
		{"no patterns", nil, nil, "Found 0 process(es) matching [].\n"},
	}
}
