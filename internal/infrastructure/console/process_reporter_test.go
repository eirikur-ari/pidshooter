package console

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestProcessReporter_Report_PrintsMatchCountAndPatterns(t *testing.T) {
	for name, test := range newProcessReporterTestCases() {
		t.Run(name, func(t *testing.T) {
			// Given
			var output bytes.Buffer
			reporter := &ProcessReporter{writer: &output}

			// When
			reporter.Report(test.matches, test.patterns)

			// Then
			assert.Equal(t, test.expected, output.String())
		})
	}
}

type processReporterTestCase struct {
	matches  []outbound.ProcessInfo
	patterns []string
	expected string
}

func newProcessReporterTestCases() map[string]processReporterTestCase {
	return map[string]processReporterTestCase{
		"several matches and patterns": {
			matches:  []outbound.ProcessInfo{{PID: 1}, {PID: 2}, {PID: 3}},
			patterns: []string{"chrome", "sleep"},
			expected: "Found 3 process(es) matching [chrome sleep].\n",
		},
		"no matches": {
			matches:  nil,
			patterns: []string{"chrome"},
			expected: "Found 0 process(es) matching [chrome].\n",
		},
		"no patterns": {
			matches:  nil,
			patterns: nil,
			expected: "Found 0 process(es) matching [].\n",
		},
	}
}
