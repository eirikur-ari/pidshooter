package outbound

import "time"

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
	Load() (ScoreBoard, error)
	Save(board ScoreBoard) error
}
