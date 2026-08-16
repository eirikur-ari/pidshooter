package score

// Tracker accumulates kills, freed memory, and the running high score during a live game run.
type Tracker struct {
	Kills     int
	FreedMem  int64
	HighScore int
}

// RecordKill tallies a kill and bumps HighScore if the running kill count beats it.
func (t *Tracker) RecordKill(rss int64) {
	t.Kills++
	t.FreedMem += rss
	if t.Kills > t.HighScore {
		t.HighScore = t.Kills
	}
}
