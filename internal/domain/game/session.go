package game

import "time"

// Session tracks the results accumulated during a single game run.
type Session struct {
	kills     int
	freedMem  int64
	highScore int
	startTime time.Time
}

func (s *Session) Kills() int           { return s.kills }
func (s *Session) FreedMem() int64      { return s.freedMem }
func (s *Session) StartTime() time.Time { return s.startTime }
func (s *Session) SetHighScore(n int)   { s.highScore = n }

func (s *Session) RecordKill(rss int64) {
	s.kills++
	s.freedMem += rss
}
