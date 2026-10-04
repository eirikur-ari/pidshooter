package score

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEntry_beats_RanksByKillsThenDudsThenSpeedThenDurationThenFreedMem(t *testing.T) {
	tests := []struct {
		name   string
		winner Entry
		loser  Entry
	}{
		{
			name:   "more kills wins regardless of every other field",
			winner: Entry{Kills: 10, FreedMem: 100},
			loser:  Entry{Kills: 5, Duds: 9, Speed: 5.0, FreedMem: 9000},
		},
		{
			name:   "more duds wins a kills tie regardless of speed or freed memory",
			winner: Entry{Kills: 5, Duds: 3, Speed: 1.0, FreedMem: 1000},
			loser:  Entry{Kills: 5, Duds: 1, Speed: 5.0, FreedMem: 9000},
		},
		{
			name:   "higher speed wins a kills and duds tie regardless of freed memory",
			winner: Entry{Kills: 5, Speed: 3.0, FreedMem: 1000},
			loser:  Entry{Kills: 5, Speed: 2.0, FreedMem: 9000},
		},
		{
			name:   "shorter duration wins a kills, duds and speed tie regardless of freed memory",
			winner: Entry{Kills: 5, Speed: 2.0, Duration: 10.0, FreedMem: 1000},
			loser:  Entry{Kills: 5, Speed: 2.0, Duration: 20.0, FreedMem: 9000},
		},
		{
			name:   "more freed memory wins when every other ranked field ties",
			winner: Entry{Kills: 5, FreedMem: 2000},
			loser:  Entry{Kills: 5, FreedMem: 1000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			winnerBeatsLoser := tt.winner.beats(tt.loser)
			loserBeatsWinner := tt.loser.beats(tt.winner)

			// Then
			assert.True(t, winnerBeatsLoser)
			assert.False(t, loserBeatsWinner)
		})
	}
}

func TestEntry_beats_IsFalseBothWaysWhenAllRankedFieldsMatch(t *testing.T) {
	// Given
	entry := newEntryFixture()
	other := newEntryFixture()
	other.Date = entry.Date.Add(time.Hour)

	// Then
	assert.False(t, entry.beats(other), "date is not a ranked field")
	assert.False(t, other.beats(entry), "date is not a ranked field")
}

func TestEntry_isScore_IsTrueOnlyForGenuineScores(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(*Entry)
		expected bool
	}{
		{name: "positive kills", mutate: func(e *Entry) {}, expected: true},
		{name: "zero duds and zero freed memory", mutate: func(e *Entry) { e.Duds, e.FreedMem = 0, 0 }, expected: true},
		{name: "zero kills", mutate: func(e *Entry) { e.Kills = 0 }, expected: false},
		{name: "negative kills", mutate: func(e *Entry) { e.Kills = -1 }, expected: false},
		{name: "negative duds", mutate: func(e *Entry) { e.Duds = -1 }, expected: false},
		{name: "negative freed memory", mutate: func(e *Entry) { e.FreedMem = -1 }, expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			entry := newEntryFixture()
			tt.mutate(&entry)

			// When
			genuine := entry.isScore()

			// Then
			assert.Equal(t, tt.expected, genuine)
		})
	}
}
