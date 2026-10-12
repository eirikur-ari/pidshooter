package testutil

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

type FakeScoreReporter struct {
	Reported *outbound.ScoreSummary
}

func (r *FakeScoreReporter) Report(summary outbound.ScoreSummary) {
	r.Reported = &summary
}
