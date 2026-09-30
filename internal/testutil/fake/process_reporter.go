package fake

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

// ProcessReporter is a test double for outbound.ProcessReporter.
type ProcessReporter struct {
	Reported *ProcessReport // captured by the most recent Report call, nil if Report was never called
}

// ProcessReport captures a single Report call's arguments.
type ProcessReport struct {
	Matches  []outbound.ProcessInfo
	Patterns []string
}

func (r *ProcessReporter) Report(matches []outbound.ProcessInfo, patterns []string) {
	r.Reported = &ProcessReport{Matches: matches, Patterns: patterns}
}
