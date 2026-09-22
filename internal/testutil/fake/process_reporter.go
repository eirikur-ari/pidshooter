package fake

// ProcessReporter is a test double for outbound.ProcessReporter.
type ProcessReporter struct {
	Reported *ProcessReport // captured by the most recent Report call, nil if Report was never called
}

// ProcessReport captures a single Report call's arguments.
type ProcessReport struct {
	Count    int
	Patterns []string
}

func (r *ProcessReporter) Report(count int, patterns []string) {
	r.Reported = &ProcessReport{Count: count, Patterns: patterns}
}
