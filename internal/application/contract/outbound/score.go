package outbound

import (
	"time"
)

// ScoreEntry is the outcome of a single session.
type ScoreEntry struct {
	Kills int
	// Duds is the number of targets whose backing process was already gone
	// before a kill could land on it.
	Duds int
	// FreedMem is the cumulative memory freed by kills, in bytes.
	FreedMem int64
	// Speed is the target movement speed the session ran at.
	Speed float64
	// Time is the configured session time limit in seconds, distinct from
	// Duration (how long the session actually ran).
	Time int
	// Duration is how long the session actually ran, in seconds.
	Duration float64
	Date     time.Time
}

// ScoreBoard is a collection of score entries.
type ScoreBoard struct {
	Scores []ScoreEntry
}

// ScoreStore persists and retrieves the score board.
type ScoreStore interface {
	// Load returns the persisted score board. If no board has been
	// persisted yet, it returns an empty ScoreBoard and a NotFoundError.
	// If the persisted data exists but cannot be parsed, it returns an
	// empty ScoreBoard and a CorruptedDataError.
	Load() (ScoreBoard, error)
	// Save persists board, overwriting any previously persisted board.
	Save(board ScoreBoard) error
}

// ScoreSummary is a session's outcome together with the board's entries.
type ScoreSummary struct {
	Kills    int
	Duds     int
	FreedMem int64
	Duration float64
	// IsTopScore reports whether the session's outcome beat all previous
	// entries.
	IsTopScore bool
	Entries    []ScoreEntry
}

// ScoreReporter reports session results and the score entries to the user.
type ScoreReporter interface {
	// Report displays summary to the user.
	Report(summary ScoreSummary)
}
