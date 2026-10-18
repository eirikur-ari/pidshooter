package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
)

func TestPlay_Run_ReturnsFatalErrorAndStopsAtFailingStage(t *testing.T) {
	tests := newFailingStageTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			setup := newPlayMockSetup()
			request := newPlayRequestFixture()
			fatal := newFatalError(test.name)
			setup.expectRunFailingAt(test.failingStage, fatal)

			// When
			err := setup.play.Run(request)

			// Then
			assert.ErrorIs(t, err, fatal)
			assert.Equal(t, []string{test.name}, setup.logger.Errored)
		})
	}
}

func TestPlay_Run_ContinuesAfterNonFatalErrorAtStage(t *testing.T) {
	tests := newWarningStageTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			setup := newPlayMockSetup()
			request := newPlayRequestFixture()
			setup.expectRunWarningAt(test.warningStage, newWarning(test.name))

			// When
			err := setup.play.Run(request)

			// Then
			require.NoError(t, err)
			assert.Equal(t, []string{test.name}, setup.logger.Warned)
			setup.scores.AssertCalled(t, "ReportResults", mock.Anything)
		})
	}
}

func TestPlay_Run_PassesEachStepsResultToTheNext(t *testing.T) {
	// Given
	setup := newPlayMockSetup()
	request := inbound.PlayRequest{
		Patterns: []string{"target"},
		Game:     inbound.GameOptions{Speed: new(3.5)},
		Process:  inbound.ProcessOptions{IncludeRoot: new(true)},
	}
	options := config.Options{
		Game:    config.GameOptions{Speed: new(3.5)},
		Process: config.ProcessOptions{IncludeRoot: new(true)},
	}
	setup.configResult = config.Result{
		Process: config.ProcessResult{IncludeRoot: true},
		Game:    config.GameResult{ConfirmMode: true, Speed: 3.5, TimeLimit: 45},
	}
	loaded := score.LoadResult{Entries: []score.BoardEntry{{Kills: 4}}, HighScore: 4}
	result := game.PlayResult{Duration: 12.5, LowestSpeed: 2.5, Kills: 3, Duds: 1, FreedMem: 8192}
	recorded := score.RecordResult{Entries: []score.BoardEntry{{Kills: 4}, {Kills: 3}}, NewHighScore: true}
	setup.config.On("Load", options).Return(setup.configResult, nil)
	setup.processes.On("FindProcesses", process.FindRequest{Patterns: []string{"target"}, IncludeRoot: true}).Return(setup.found, nil)
	setup.scores.On("LoadScoreBoard").Return(loaded, nil)
	setup.game.On("Play", game.PlayRequest{
		ConfirmMode: true, Speed: 3.5, TimeLimit: 45,
		Processes: []game.ProcessRequest{{PID: 200, Name: "target", Rss: 1024, UID: 1000}},
		HighScore: 4,
	}).Return(result, nil)
	setup.scores.On("RecordScore", score.RecordRequest{
		Entries: loaded.Entries, Kills: 3, Duds: 1, FreedMem: 8192, LowestSpeed: 2.5, TimeLimit: 45, Duration: 12.5,
	}).Return(recorded, nil)
	setup.scores.On("ReportResults", score.ReportRequest{
		Duration: 12.5, Kills: 3, Duds: 1, FreedMem: 8192, Entries: recorded.Entries, NewHighScore: true,
	}).Return()

	// When
	err := setup.play.Run(request)

	// Then
	require.NoError(t, err)
	setup.config.AssertExpectations(t)
	setup.processes.AssertExpectations(t)
	setup.scores.AssertExpectations(t)
	setup.game.AssertExpectations(t)
}

func TestPlay_Run_HandlesEachPlayErrorWithoutFailing(t *testing.T) {
	// Given
	setup := newPlayMockSetup()
	request := newPlayRequestFixture()
	setup.config.On("Load", mock.Anything).Return(setup.configResult, nil)
	setup.expectSessionAfterConfig(game.PlayResult{
		Errors: []error{newWarning("could not kill stubborn (PID 200)"), newWarning("gone (PID 300) ran away")},
	})

	// When
	err := setup.play.Run(request)

	// Then
	require.NoError(t, err)
	assert.Equal(t, []string{"could not kill stubborn (PID 200)", "gone (PID 300) ran away"}, setup.logger.Warned)
	assert.Empty(t, setup.logger.Errored)
}

func newFailingStageTestCase() []struct {
	name         string
	failingStage stage
} {
	return []struct {
		name         string
		failingStage stage
	}{
		{"config load", stageConfigLoad},
		{"process search", stageProcessSearch},
		{"score board load", stageScoreBoardLoad},
		{"game play", stageGamePlay},
		{"game play result", stageGamePlayResult},
		{"score recording", stageScoreRecording},
	}
}

func newWarningStageTestCase() []struct {
	name         string
	warningStage stage
} {
	return []struct {
		name         string
		warningStage stage
	}{
		{"config load", stageConfigLoad},
		{"score board load", stageScoreBoardLoad},
		{"score recording", stageScoreRecording},
	}
}
