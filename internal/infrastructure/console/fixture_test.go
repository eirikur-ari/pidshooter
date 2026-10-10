package console

import (
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func dateFixture() time.Time {
	return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
}

func newScoreEntryFixture() outbound.ScoreEntry {
	return outbound.ScoreEntry{Kills: 5, Duds: 1, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: dateFixture()}
}

func newScoreSummaryFixture() outbound.ScoreSummary {
	return outbound.ScoreSummary{Kills: 5, Duds: 1, FreedMem: 2048, Duration: 12.5}
}
