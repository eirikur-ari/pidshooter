package game

// Stats tracks the kill results accumulated during a single game run.
type Stats struct {
	kills     int
	freedMem  int64
	highScore int
}

func (s *Stats) Kills() int         { return s.kills }
func (s *Stats) FreedMem() int64    { return s.freedMem }
func (s *Stats) SetHighScore(n int) { s.highScore = n }

func (s *Stats) RecordKill(rss int64) {
	s.kills++
	s.freedMem += rss
	if s.kills > s.highScore {
		s.highScore = s.kills
	}
}
