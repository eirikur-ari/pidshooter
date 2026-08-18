package score

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrackerRecordKill(t *testing.T) {
	tr := &Tracker{}

	tr.RecordKill(1024)
	assert.Equal(t, 1, tr.Kills)
	assert.Equal(t, int64(1024), tr.FreedMem)

	tr.RecordKill(2048)
	assert.Equal(t, 2, tr.Kills)
	assert.Equal(t, int64(3072), tr.FreedMem)
}

func TestTrackerRecordKillUpdatesHighScore(t *testing.T) {
	tr := &Tracker{HighScore: 5}

	for i := range 5 {
		tr.RecordKill(0)
		assert.Equal(t, 5, tr.HighScore, "after %d kills: expected highScore=5 (not beaten yet)", i+1)
	}

	tr.RecordKill(0)
	assert.Equal(t, 6, tr.HighScore, "expected highScore=6 after beating old record")

	tr.RecordKill(0)
	assert.Equal(t, 7, tr.HighScore, "expected highScore=7 after second beat")
}
