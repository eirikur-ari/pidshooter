package score

import (
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/score"
)

func dateFixture() time.Time {
	return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
}

func newBoardEntryFixture() BoardEntry {
	return BoardEntry{Kills: 5, Duds: 1, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: dateFixture()}
}

func newEntryFixture() score.Entry {
	return score.Entry{Kills: 5, Duds: 1, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: dateFixture()}
}

func newScoreEntryFixture() outbound.ScoreEntry {
	return outbound.ScoreEntry{Kills: 5, Duds: 1, FreedMem: 2048, Speed: 2.5, Time: 30, Duration: 12.5, Date: dateFixture()}
}
