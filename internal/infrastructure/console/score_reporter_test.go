package console

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestScoreReporter_Report_PrintsSummaryAndHighScores(t *testing.T) {
	for name, test := range newScoreReporterTestCases() {
		t.Run(name, func(t *testing.T) {
			// Given
			var output bytes.Buffer
			reporter := &ScoreReporter{writer: &output}

			// When
			reporter.Report(test.summary)

			// Then
			assert.Equal(t, test.expected, output.String())
		})
	}
}

type scoreReporterTestCase struct {
	summary  outbound.ScoreSummary
	expected string
}

func newScoreReporterTestCases() map[string]scoreReporterTestCase {
	const gameOver = "\n  Game Over! Kills: 5 | Duds: 1 | Freed: 2.0 KB | Time: 12.5s\n"
	const noScores = "\n  No high scores yet!\n"
	const tableTop = "\n  ╔════╦═══════╦═══════╦═══════╦════════╦════════════╦════════════╗\n" +
		"  ║  # ║ Kills ║ Duds  ║ Speed ║  Time  ║   Freed    ║    Date    ║\n" +
		"  ╠════╬═══════╬═══════╬═══════╬════════╬════════════╬════════════╣\n"
	const tableBottom = "  ╚════╩═══════╩═══════╩═══════╩════════╩════════════╩════════════╝\n"

	withEntries := newScoreSummaryFixture()
	withEntries.Entries = []outbound.ScoreEntry{
		newScoreEntryFixture(),
		{Kills: 12, Duds: 0, FreedMem: 0, Speed: 1.0, Duration: 3.0, Date: dateFixture().Add(24 * time.Hour)},
	}

	topScore := newScoreSummaryFixture()
	topScore.IsTopScore = true

	return map[string]scoreReporterTestCase{
		"no entries": {
			summary:  newScoreSummaryFixture(),
			expected: gameOver + noScores,
		},
		"top score with no entries": {
			summary:  topScore,
			expected: gameOver + "  🏆 New high score!\n" + noScores,
		},
		"entries are ranked in the given order": {
			summary: withEntries,
			expected: gameOver + tableTop +
				"  ║  1 ║    5  ║    1  ║  2.5x ║  12.5s ║   2.0 KB   ║ 2026-01-02 ║\n" +
				"  ║  2 ║   12  ║    0  ║  1.0x ║   3.0s ║      0 B   ║ 2026-01-03 ║\n" +
				tableBottom,
		},
	}
}
