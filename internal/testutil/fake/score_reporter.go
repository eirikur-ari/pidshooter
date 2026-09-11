package fake

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

// ScoreReporter is a test double for outbound.ScoreReporter.
type ScoreReporter struct {
	Reported *outbound.ScoreSummary // captured by the most recent Report call, nil if Report was never called
}

func (r *ScoreReporter) Report(summary outbound.ScoreSummary) {
	r.Reported = &summary
}
