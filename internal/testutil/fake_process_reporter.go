package testutil

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

type FakeProcessReporter struct {
	Reported *ProcessReport
}

type ProcessReport struct {
	Matches  []outbound.ProcessInfo
	Patterns []string
}

func (r *FakeProcessReporter) Report(matches []outbound.ProcessInfo, patterns []string) {
	r.Reported = &ProcessReport{Matches: matches, Patterns: patterns}
}
