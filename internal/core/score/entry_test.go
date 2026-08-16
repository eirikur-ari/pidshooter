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

func TestEntryBeatsTiebreakByFreedMem(t *testing.T) {
	more := Entry{Kills: 5, FreedMem: 2000}
	less := Entry{Kills: 5, FreedMem: 1000}
	assert.True(t, more.beats(less), "expected higher freed mem to win on kills tie")
	assert.False(t, less.beats(more), "expected lower freed mem to lose on kills tie")
}
