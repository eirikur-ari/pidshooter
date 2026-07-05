package scorefilestore

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/domain/score"
)

func newTempStore(t *testing.T) *Store {
	t.Helper()
	return &Store{path: filepath.Join(t.TempDir(), "scores.json")}
}

func TestLoad_FileNotExist_ReturnsEmptyBoard(t *testing.T) {
	s := newTempStore(t)
	board, err := s.Load()
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if board.HighScore() != 0 {
		t.Errorf("expected empty board, got high score %d", board.HighScore())
	}
}

func TestLoad_InvalidJSON_ReturnsError(t *testing.T) {
	s := newTempStore(t)
	if err := os.WriteFile(s.path, []byte("not valid json{{{"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := s.Load()
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestSave_CreatesFile(t *testing.T) {
	s := newTempStore(t)
	if err := s.Save(&score.Board{}); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		t.Error("expected file to be created after Save")
	}
}

func TestSave_FilePermissions(t *testing.T) {
	s := newTempStore(t)
	if err := s.Save(&score.Board{}); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	info, err := os.Stat(s.path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected file permissions 0600, got %04o", perm)
	}
}

func TestSave_Load_RoundTrip(t *testing.T) {
	s := newTempStore(t)
	board := &score.Board{}
	board.Add(score.Entry{Kills: 7, FreedMem: 4096, Speed: 2.5, Time: 60, Duration: 45.0, Date: time.Now()})

	if err := s.Save(board); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.HighScore() != 7 {
		t.Errorf("expected high score 7 after round-trip, got %d", loaded.HighScore())
	}
	if len(loaded.Scores) != 1 {
		t.Errorf("expected 1 entry after round-trip, got %d", len(loaded.Scores))
	}
}

func TestSave_Load_MultipleEntries(t *testing.T) {
	s := newTempStore(t)
	board := &score.Board{}
	board.Add(score.Entry{Kills: 3, FreedMem: 1024, Speed: 2.0, Date: time.Now()})
	board.Add(score.Entry{Kills: 9, FreedMem: 8192, Speed: 3.0, Date: time.Now()})
	board.Add(score.Entry{Kills: 1, FreedMem: 512, Speed: 1.0, Date: time.Now()})

	if err := s.Save(board); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.HighScore() != 9 {
		t.Errorf("expected high score 9, got %d", loaded.HighScore())
	}
	if len(loaded.Scores) != 3 {
		t.Errorf("expected 3 entries, got %d", len(loaded.Scores))
	}
}

func TestSave_OverwritesPreviousFile(t *testing.T) {
	s := newTempStore(t)

	first := &score.Board{}
	first.Add(score.Entry{Kills: 2, Date: time.Now()})
	if err := s.Save(first); err != nil {
		t.Fatalf("first save failed: %v", err)
	}

	second := &score.Board{}
	second.Add(score.Entry{Kills: 10, Date: time.Now()})
	if err := s.Save(second); err != nil {
		t.Fatalf("second save failed: %v", err)
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.HighScore() != 10 {
		t.Errorf("expected high score 10 from second save, got %d", loaded.HighScore())
	}
}
