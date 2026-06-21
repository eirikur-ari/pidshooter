package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ScoreEntry represents a single high score record.
type ScoreEntry struct {
	Kills    int       `json:"kills"`
	FreedMem int64     `json:"freed_mem"`
	Speed    float64   `json:"speed"`
	Time     int       `json:"time_limit"`
	Duration float64   `json:"duration_secs"`
	Date     time.Time `json:"date"`
}

// ScoreBoard holds all high scores.
type ScoreBoard struct {
	Scores []ScoreEntry `json:"scores"`
}

const maxScores = 10

// scoreFilePath returns the path to the high scores file.
func scoreFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".config", "pidshooter")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "highscores.json")
}

// LoadScores reads the score board from disk.
func LoadScores() *ScoreBoard {
	sb := &ScoreBoard{}
	data, err := os.ReadFile(scoreFilePath())
	if err != nil {
		return sb
	}
	_ = json.Unmarshal(data, sb)
	return sb
}

// Save writes the score board to disk.
func (sb *ScoreBoard) Save() error {
	data, err := json.MarshalIndent(sb, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scores: %w", err)
	}
	return os.WriteFile(scoreFilePath(), data, 0644)
}

// Add inserts a new score entry and keeps only the top N.
func (sb *ScoreBoard) Add(entry ScoreEntry) bool {
	sb.Scores = append(sb.Scores, entry)
	sort.Slice(sb.Scores, func(i, j int) bool {
		// Sort by kills descending, then by freed memory descending
		if sb.Scores[i].Kills != sb.Scores[j].Kills {
			return sb.Scores[i].Kills > sb.Scores[j].Kills
		}
		return sb.Scores[i].FreedMem > sb.Scores[j].FreedMem
	})

	isHighScore := len(sb.Scores) <= maxScores ||
		(len(sb.Scores) > 0 && sb.Scores[len(sb.Scores)-1] != entry)

	if len(sb.Scores) > maxScores {
		sb.Scores = sb.Scores[:maxScores]
	}

	return isHighScore
}

// HighScore returns the current top score (kills), or 0 if no scores exist.
func (sb *ScoreBoard) HighScore() int {
	if len(sb.Scores) == 0 {
		return 0
	}
	return sb.Scores[0].Kills
}

// PrintScores displays the high score table to stdout.
func (sb *ScoreBoard) PrintScores() {
	if len(sb.Scores) == 0 {
		fmt.Println("\n  No high scores yet!")
		return
	}

	fmt.Println("\n  ╔════╦═══════╦════════════╦═══════╦════════════╗")
	fmt.Println("  ║  # ║ Kills ║   Freed    ║ Speed ║    Date    ║")
	fmt.Println("  ╠════╬═══════╬════════════╬═══════╬════════════╣")

	for i, s := range sb.Scores {
		mem := formatBytes(s.FreedMem)
		date := s.Date.Format("2006-01-02")
		fmt.Printf("  ║ %2d ║  %3d  ║ %8s   ║ %4.1fx ║ %s ║\n",
			i+1, s.Kills, mem, s.Speed, date)
	}

	fmt.Println("  ╚════╩═══════╩════════════╩═══════╩════════════╝")
}
