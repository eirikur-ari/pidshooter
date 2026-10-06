package console

import (
	"fmt"
	"io"
	"os"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

// ScoreReporter writes score reports to an io.Writer.
type ScoreReporter struct {
	writer io.Writer
}

// NewScoreReporter returns a ScoreReporter that writes to os.Stdout.
func NewScoreReporter() *ScoreReporter {
	return &ScoreReporter{writer: os.Stdout}
}

// Report writes the game-over summary, a congratulation line if the session
// achieved the top score, and the high-score table (or a "no high scores
// yet" message) to the underlying writer.
func (r *ScoreReporter) Report(summary outbound.ScoreSummary) {
	_, _ = fmt.Fprintf(r.writer, "\n  Game Over! Kills: %d | Duds: %d | Freed: %s | Time: %.1fs\n",
		summary.Kills, summary.Duds, util.FormatBytes(summary.FreedMem), summary.Duration)

	if summary.IsTopScore {
		_, _ = fmt.Fprintln(r.writer, "  🏆 New high score!")
	}

	if len(summary.Entries) == 0 {
		_, _ = fmt.Fprintln(r.writer, "\n  No high scores yet!")
		return
	}

	_, _ = fmt.Fprintln(r.writer, "\n  ╔════╦═══════╦═══════╦═══════╦════════╦════════════╦════════════╗")
	_, _ = fmt.Fprintln(r.writer, "  ║  # ║ Kills ║ Duds  ║ Speed ║  Time  ║   Freed    ║    Date    ║")
	_, _ = fmt.Fprintln(r.writer, "  ╠════╬═══════╬═══════╬═══════╬════════╬════════════╬════════════╣")

	for i, entry := range summary.Entries {
		mem := util.FormatBytes(entry.FreedMem)
		date := entry.Date.Format("2006-01-02")
		_, _ = fmt.Fprintf(r.writer, "  ║ %2d ║  %3d  ║  %3d  ║ %4.1fx ║ %5.1fs ║ %8s   ║ %s ║\n",
			i+1, entry.Kills, entry.Duds, entry.Speed, entry.Duration, mem, date)
	}

	_, _ = fmt.Fprintln(r.writer, "  ╚════╩═══════╩═══════╩═══════╩════════╩════════════╩════════════╝")
}
