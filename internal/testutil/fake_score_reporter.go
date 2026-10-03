package testutil

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

// FakeScoreReporter is a test double for outbound.ScoreReporter.
type FakeScoreReporter struct {
	Reported *outbound.ScoreSummary // captured by the most recent Report call, nil if Report was never called
}

func (r *FakeScoreReporter) Report(summary outbound.ScoreSummary) {
	r.Reported = &summary
}
