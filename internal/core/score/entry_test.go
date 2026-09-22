package score

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Entry.beats ---

func TestEntryBeatsByKills(t *testing.T) {
	high := Entry{Kills: 10, FreedMem: 100}
	low := Entry{Kills: 5, FreedMem: 9000}
	assert.True(t, high.beats(low), "expected higher kills to win regardless of freed mem")
	assert.False(t, low.beats(high), "expected lower kills to lose")
}

func TestEntryBeatsTiebreakByDuds(t *testing.T) {
	fewerDuds := Entry{Kills: 5, Duds: 1, Speed: 5.0, FreedMem: 9000}
	moreDuds := Entry{Kills: 5, Duds: 3, Speed: 1.0, FreedMem: 1000}
	assert.True(t, moreDuds.beats(fewerDuds), "expected more duds to win on kills tie, regardless of speed or freed mem")
	assert.False(t, fewerDuds.beats(moreDuds), "expected fewer duds to lose on kills tie")
}

func TestEntryBeatsTiebreakBySpeed(t *testing.T) {
	faster := Entry{Kills: 5, Speed: 3.0, FreedMem: 1000}
	slower := Entry{Kills: 5, Speed: 2.0, FreedMem: 9000}
	assert.True(t, faster.beats(slower), "expected higher speed to win on kills tie, regardless of freed mem")
	assert.False(t, slower.beats(faster), "expected lower speed to lose on kills tie")
}

func TestEntryBeatsTiebreakByDuration(t *testing.T) {
	quicker := Entry{Kills: 5, Speed: 2.0, Duration: 10.0, FreedMem: 1000}
	slower := Entry{Kills: 5, Speed: 2.0, Duration: 20.0, FreedMem: 9000}
	assert.True(t, quicker.beats(slower), "expected shorter duration to win on kills and speed tie")
	assert.False(t, slower.beats(quicker), "expected longer duration to lose on kills and speed tie")
}

func TestEntryBeatsTiebreakByFreedMem(t *testing.T) {
	more := Entry{Kills: 5, FreedMem: 2000}
	less := Entry{Kills: 5, FreedMem: 1000}
	assert.True(t, more.beats(less), "expected higher freed mem to win on kills, speed, and duration tie")
	assert.False(t, less.beats(more), "expected lower freed mem to lose on kills, speed, and duration tie")
}

// --- Entry.isScore ---

func TestEntryIsScoreTrueForPositiveKillsAndFreedMem(t *testing.T) {
	assert.True(t, Entry{Kills: 1, FreedMem: 0}.isScore())
}

func TestEntryIsScoreFalseForZeroKills(t *testing.T) {
	assert.False(t, Entry{Kills: 0, FreedMem: 100}.isScore())
}

func TestEntryIsScoreFalseForNegativeKills(t *testing.T) {
	assert.False(t, Entry{Kills: -1, FreedMem: 100}.isScore())
}

func TestEntryIsScoreFalseForNegativeFreedMem(t *testing.T) {
	assert.False(t, Entry{Kills: 5, FreedMem: -1}.isScore(), "a real session can never free negative memory")
}
