// Package score manages the high score persistence and display.
package score

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/game"
)

// Entry represents a single high score record.
type Entry struct {
	Kills    int       `json:"kills"`
	FreedMem int64     `json:"freed_mem"`
	Speed    float64   `json:"speed"`
	Time     int       `json:"time_limit"`
	Duration float64   `json:"duration_secs"`
	Date     time.Time `json:"date"`
}

// Board holds all high scores.
type Board struct {
	Scores []Entry `json:"scores"`
}

const maxScores = 10

func filePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".config", "pidshooter")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "highscores.json")
}

// Load reads the score board from disk.
func Load() *Board {
	b := &Board{}
	data, err := os.ReadFile(filePath())
	if err != nil {
		return b
	}
	_ = json.Unmarshal(data, b)
	return b
}

// Save writes the score board to disk.
func (b *Board) Save() error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scores: %w", err)
	}
	return os.WriteFile(filePath(), data, 0644)
}

// Add inserts a new score entry and keeps only the top N.
func (b *Board) Add(entry Entry) bool {
	b.Scores = append(b.Scores, entry)
	sort.Slice(b.Scores, func(i, j int) bool {
		if b.Scores[i].Kills != b.Scores[j].Kills {
			return b.Scores[i].Kills > b.Scores[j].Kills
		}
		return b.Scores[i].FreedMem > b.Scores[j].FreedMem
	})

	isHighScore := len(b.Scores) <= maxScores ||
		(len(b.Scores) > 0 && b.Scores[len(b.Scores)-1] != entry)

	if len(b.Scores) > maxScores {
		b.Scores = b.Scores[:maxScores]
	}

	return isHighScore
}

// HighScore returns the current top score (kills), or 0 if none.
func (b *Board) HighScore() int {
	if len(b.Scores) == 0 {
		return 0
	}
	return b.Scores[0].Kills
}

// PrintScores displays the high score table to stdout.
func (b *Board) PrintScores() {
	if len(b.Scores) == 0 {
		fmt.Println("\n  No high scores yet!")
		return
	}

	fmt.Println("\n  ╔════╦═══════╦════════════╦═══════╦════════════╗")
	fmt.Println("  ║  # ║ Kills ║   Freed    ║ Speed ║    Date    ║")
	fmt.Println("  ╠════╬═══════╬════════════╬═══════╬════════════╣")

	for i, s := range b.Scores {
		mem := game.FormatBytes(s.FreedMem)
		date := s.Date.Format("2006-01-02")
		fmt.Printf("  ║ %2d ║  %3d  ║ %8s   ║ %4.1fx ║ %s ║\n",
			i+1, s.Kills, mem, s.Speed, date)
	}

	fmt.Println("  ╚════╩═══════╩════════════╩═══════╩════════════╝")
}
