package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStats_RecordKill(t *testing.T) {
	s := Stats{}

	s.RecordKill(1024)
	assert.Equal(t, 1, s.Kills())
	assert.Equal(t, int64(1024), s.FreedMem())

	s.RecordKill(2048)
	assert.Equal(t, 2, s.Kills())
	assert.Equal(t, int64(3072), s.FreedMem())
}

func TestStats_RecordKill_UpdatesHighScore(t *testing.T) {
	s := Stats{}
	s.SetHighScore(5)

	for i := 0; i < 5; i++ {
		s.RecordKill(0)
		assert.Equal(t, 5, s.highScore, "after %d kills: expected highScore=5 (not beaten yet)", i+1)
	}

	s.RecordKill(0)
	assert.Equal(t, 6, s.highScore, "expected highScore=6 after beating old record")

	s.RecordKill(0)
	assert.Equal(t, 7, s.highScore, "expected highScore=7 after second beat")
}

func TestStats_SetHighScore(t *testing.T) {
	s := Stats{}
	s.SetHighScore(100)
	assert.Equal(t, 100, s.highScore)
}
