package testutil

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

// FakeProcessReporter is a test double for outbound.ProcessReporter.
type FakeProcessReporter struct {
	Reported *ProcessReport // captured by the most recent Report call, nil if Report was never called
}

// ProcessReport captures a single Report call's arguments.
type ProcessReport struct {
	Matches  []outbound.ProcessInfo
	Patterns []string
}

func (r *FakeProcessReporter) Report(matches []outbound.ProcessInfo, patterns []string) {
	r.Reported = &ProcessReport{Matches: matches, Patterns: patterns}
}
