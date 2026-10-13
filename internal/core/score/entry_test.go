package score

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEntry_beats_RanksByKillsThenDudsThenSpeedThenDurationThenFreedMem(t *testing.T) {
	tests := newBeatsTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			winnerBeatsLoser := test.winner.beats(test.loser)
			loserBeatsWinner := test.loser.beats(test.winner)

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

	// When
	entryBeatsOther := entry.beats(other)
	otherBeatsEntry := other.beats(entry)

	// Then
	assert.False(t, entryBeatsOther, "date is not a ranked field")
	assert.False(t, otherBeatsEntry, "date is not a ranked field")
}

func TestEntry_isScore_IsTrueOnlyForGenuineScores(t *testing.T) {
	tests := newIsScoreTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			entry := newEntryFixture()
			test.mutate(&entry)

			// When
			genuine := entry.isScore()

			// Then
			assert.Equal(t, test.expected, genuine)
		})
	}
}

func newBeatsTestCase() []struct {
	name   string
	winner Entry
	loser  Entry
} {
	return []struct {
		name   string
		winner Entry
		loser  Entry
	}{
		{
			"more kills wins regardless of every other field",
			Entry{Kills: 10, FreedMem: 100},
			Entry{Kills: 5, Duds: 9, Speed: 5.0, FreedMem: 9000},
		},
		{
			"more duds wins a kills tie regardless of speed or freed memory",
			Entry{Kills: 5, Duds: 3, Speed: 1.0, FreedMem: 1000},
			Entry{Kills: 5, Duds: 1, Speed: 5.0, FreedMem: 9000},
		},
		{
			"higher speed wins a kills and duds tie regardless of freed memory",
			Entry{Kills: 5, Speed: 3.0, FreedMem: 1000},
			Entry{Kills: 5, Speed: 2.0, FreedMem: 9000},
		},
		{
			"shorter duration wins a kills, duds and speed tie regardless of freed memory",
			Entry{Kills: 5, Speed: 2.0, Duration: 10.0, FreedMem: 1000},
			Entry{Kills: 5, Speed: 2.0, Duration: 20.0, FreedMem: 9000},
		},
		{
			"more freed memory wins when every other ranked field ties",
			Entry{Kills: 5, FreedMem: 2000},
			Entry{Kills: 5, FreedMem: 1000},
		},
	}
}

func newIsScoreTestCase() []struct {
	name     string
	mutate   func(*Entry)
	expected bool
} {
	return []struct {
		name     string
		mutate   func(*Entry)
		expected bool
	}{
		{"positive kills", func(e *Entry) {}, true},
		{"zero duds and zero freed memory", func(e *Entry) { e.Duds, e.FreedMem = 0, 0 }, true},
		{"zero kills", func(e *Entry) { e.Kills = 0 }, false},
		{"negative kills", func(e *Entry) { e.Kills = -1 }, false},
		{"negative duds", func(e *Entry) { e.Duds = -1 }, false},
		{"negative freed memory", func(e *Entry) { e.FreedMem = -1 }, false},
	}
}
