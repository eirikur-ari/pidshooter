package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatsRecordKill(t *testing.T) {
	s := Stats{}

	s.RecordKill(1024)
	assert.Equal(t, 1, s.Kills())
	assert.Equal(t, int64(1024), s.FreedMem())

	s.RecordKill(2048)
	assert.Equal(t, 2, s.Kills())
	assert.Equal(t, int64(3072), s.FreedMem())
}

func TestStatsRecordKillUpdatesHighScore(t *testing.T) {
	s := Stats{}
	s.SetHighScore(5)

	for i := 0; i < 5; i++ {
		s.RecordKill(0)
		assert.Equal(t, 5, s.HighScore(), "after %d kills: expected highScore=5 (not beaten yet)", i+1)
	}

	s.RecordKill(0)
	assert.Equal(t, 6, s.HighScore(), "expected highScore=6 after beating old record")

	s.RecordKill(0)
	assert.Equal(t, 7, s.HighScore(), "expected highScore=7 after second beat")
}

func TestStatsSetHighScore(t *testing.T) {
	s := Stats{}
	s.SetHighScore(100)
	assert.Equal(t, 100, s.HighScore())
}

func TestStatsHighScore(t *testing.T) {
	s := Stats{}
	assert.Equal(t, 0, s.HighScore())
	s.SetHighScore(42)
	assert.Equal(t, 42, s.HighScore())
}
