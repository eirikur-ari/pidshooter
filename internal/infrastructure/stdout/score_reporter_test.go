package stdout

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestReportPrintsGameOverSummary(t *testing.T) {
	var buf bytes.Buffer
	r := &ScoreReporter{writer: &buf}

	r.Report(outbound.ScoreSummary{Kills: 3, FreedMem: 4096, Duration: 7.5})

	assert.Contains(t, buf.String(), "Game Over!")
	assert.Contains(t, buf.String(), "Kills: 3")
}

func TestReportPrintsTrophyWhenNewHighScore(t *testing.T) {
	var buf bytes.Buffer
	r := &ScoreReporter{writer: &buf}

	r.Report(outbound.ScoreSummary{NewHighScore: true})

	assert.Contains(t, buf.String(), "New high score")
}

func TestReportOmitsTrophyWhenNotNewHighScore(t *testing.T) {
	var buf bytes.Buffer
	r := &ScoreReporter{writer: &buf}

	r.Report(outbound.ScoreSummary{NewHighScore: false})

	assert.NotContains(t, buf.String(), "New high score")
}

func TestReportPrintsNoScoresMessageWhenEmpty(t *testing.T) {
	var buf bytes.Buffer
	r := &ScoreReporter{writer: &buf}

	r.Report(outbound.ScoreSummary{})

	assert.Contains(t, buf.String(), "No high scores yet!")
}

func TestReportPrintsTable(t *testing.T) {
	var buf bytes.Buffer
	r := &ScoreReporter{writer: &buf}

	r.Report(outbound.ScoreSummary{Entries: []outbound.ScoreEntry{
		{Kills: 5, FreedMem: 100, Speed: 1.5, Date: time.Now()},
	}})

	assert.Contains(t, buf.String(), "Kills")
	assert.Contains(t, buf.String(), "Freed")
}

func TestReportPrintsDuration(t *testing.T) {
	var buf bytes.Buffer
	r := &ScoreReporter{writer: &buf}

	r.Report(outbound.ScoreSummary{Entries: []outbound.ScoreEntry{
		{Kills: 5, FreedMem: 100, Speed: 1.5, Duration: 12.3, Date: time.Now()},
	}})

	assert.Contains(t, buf.String(), "Time")
	assert.Contains(t, buf.String(), "12.3s")
}
