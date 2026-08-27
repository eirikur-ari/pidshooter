package outbound

import (
	"errors"
	"time"
)

// ErrNotFound is returned by ScoreStore.Load when no board has been persisted yet.
var ErrNotFound = errors.New("score board not found")

// ScoreEntry is the persistence representation of a single high score record.
type ScoreEntry struct {
	Kills    int
	FreedMem int64
	Speed    float64
	Time     int
	Duration float64
	Date     time.Time
}

// ScoreBoard is the persistence representation of the high score table.
type ScoreBoard struct {
	Scores []ScoreEntry
}

// ScoreStore is the outbound port for persisting and retrieving the score board.
type ScoreStore interface {
	// Load returns the persisted score board. If no board has been
	// persisted yet, it returns an empty ScoreBoard and ErrNotFound.
	Load() (ScoreBoard, error)
	// Save persists board, overwriting any previously persisted board.
	Save(board ScoreBoard) error
}
