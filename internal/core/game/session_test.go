package game

import "testing"

func TestSession_RecordKill(t *testing.T) {
	s := Session{}

	s.RecordKill(1024)
	if s.Kills() != 1 {
		t.Errorf("expected kills=1, got %d", s.Kills())
	}
	if s.FreedMem() != 1024 {
		t.Errorf("expected freedMem=1024, got %d", s.FreedMem())
	}

	s.RecordKill(2048)
	if s.Kills() != 2 {
		t.Errorf("expected kills=2, got %d", s.Kills())
	}
	if s.FreedMem() != 3072 {
		t.Errorf("expected freedMem=3072, got %d", s.FreedMem())
	}
}

func TestSession_RecordKill_UpdatesHighScore(t *testing.T) {
	s := Session{}
	s.SetHighScore(5)

	for i := 0; i < 5; i++ {
		s.RecordKill(0)
		if s.highScore != 5 {
			t.Errorf("after %d kills: expected highScore=5 (not beaten yet), got %d", i+1, s.highScore)
		}
	}

	s.RecordKill(0)
	if s.highScore != 6 {
		t.Errorf("expected highScore=6 after beating old record, got %d", s.highScore)
	}

	s.RecordKill(0)
	if s.highScore != 7 {
		t.Errorf("expected highScore=7 after second beat, got %d", s.highScore)
	}
}

func TestSession_SetHighScore(t *testing.T) {
	s := Session{}
	s.SetHighScore(100)
	if s.highScore != 100 {
		t.Errorf("expected highScore=100, got %d", s.highScore)
	}
}

func TestSession_StartTime_ZeroOnNew(t *testing.T) {
	s := Session{}
	if !s.StartTime().IsZero() {
		t.Error("expected zero StartTime on a new session")
	}
}
