package game

// Session tracks the results accumulated during a single game run.
type Session struct {
	kills     int
	freedMem  int64
	highScore int
}

func (s *Session) Kills() int         { return s.kills }
func (s *Session) FreedMem() int64    { return s.freedMem }
func (s *Session) SetHighScore(n int) { s.highScore = n }

func (s *Session) RecordKill(rss int64) {
	s.kills++
	s.freedMem += rss
	if s.kills > s.highScore {
		s.highScore = s.kills
	}
}
