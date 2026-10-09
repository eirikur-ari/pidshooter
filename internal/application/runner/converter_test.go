package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
)

func TestToPlayRequest_MapsGameConfigFoundProcessesAndHighScore(t *testing.T) {
	tests := newPlayRequestTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toPlayRequest(test.cfg, test.found, 7)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToFindRequest_MapsProcessConfig(t *testing.T) {
	tests := newFindRequestTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toFindRequest(test.cfg)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToRecordRequest_MapsLoadedEntriesAndPlayedSession(t *testing.T) {
	// Given
	board := score.LoadResult{Entries: []score.BoardEntry{{Kills: 4}}, HighScore: 4}
	result := game.PlayResult{Duration: 12.5, LowestSpeed: 2.5, Kills: 3, Duds: 1, FreedMem: 8192}

	// When
	actual := toRecordRequest(board, result, 45)

	// Then
	assert.Equal(t, score.RecordRequest{
		Entries: []score.BoardEntry{{Kills: 4}}, Kills: 3, Duds: 1, FreedMem: 8192,
		LowestSpeed: 2.5, TimeLimit: 45, Duration: 12.5,
	}, actual)
}

func TestToReportRequest_MapsPlayedSessionAndRecordedEntries(t *testing.T) {
	// Given
	result := game.PlayResult{Duration: 12.5, Kills: 3, Duds: 1, FreedMem: 8192}
	recorded := score.RecordResult{Entries: []score.BoardEntry{{Kills: 4}, {Kills: 3}}, NewHighScore: true}

	// When
	actual := toReportRequest(result, recorded)

	// Then
	assert.Equal(t, score.ReportRequest{
		Duration: 12.5, Kills: 3, Duds: 1, FreedMem: 8192,
		Entries: []score.BoardEntry{{Kills: 4}, {Kills: 3}}, NewHighScore: true,
	}, actual)
}

func newPlayRequestTestCase() []struct {
	name     string
	cfg      config.GameResult
	found    []process.FindResult
	expected game.PlayRequest
} {
	cfg := config.GameResult{ConfirmMode: true, Speed: 3.5, TimeLimit: 45}

	return []struct {
		name     string
		cfg      config.GameResult
		found    []process.FindResult
		expected game.PlayRequest
	}{
		{
			name: "several found processes",
			cfg:  cfg,
			found: []process.FindResult{
				{PID: 1, Name: "a", Rss: 100, UID: 1000},
				{PID: 2, Name: "b", Rss: 200, UID: 2000},
			},
			expected: game.PlayRequest{
				ConfirmMode: true, Speed: 3.5, TimeLimit: 45,
				Processes: []game.ProcessRequest{
					{PID: 1, Name: "a", Rss: 100, UID: 1000},
					{PID: 2, Name: "b", Rss: 200, UID: 2000},
				},
				HighScore: 7,
			},
		},
		{
			name:  "no found processes",
			cfg:   cfg,
			found: nil,
			expected: game.PlayRequest{
				ConfirmMode: true, Speed: 3.5, TimeLimit: 45,
				Processes: []game.ProcessRequest{},
				HighScore: 7,
			},
		},
	}
}

func newFindRequestTestCase() []struct {
	name     string
	cfg      config.ProcessResult
	expected process.FindRequest
} {
	return []struct {
		name     string
		cfg      config.ProcessResult
		expected process.FindRequest
	}{
		{"include root only", config.ProcessResult{IncludeRoot: true}, process.FindRequest{IncludeRoot: true}},
		{"allow root only", config.ProcessResult{AllowRoot: true}, process.FindRequest{AllowRoot: true}},
	}
}
