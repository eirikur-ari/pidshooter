package console

import (
	"fmt"
	"io"
	"os"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

// ScoreReporter reports the outcome of a finished session and the recorded high scores.
type ScoreReporter struct {
	writer io.Writer
}

// NewScoreReporter returns a ScoreReporter that writes to standard output.
func NewScoreReporter() *ScoreReporter {
	return &ScoreReporter{writer: os.Stdout}
}

// Report writes the summary of the finished session, whether it set a top
// score, and the recorded high scores.
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
