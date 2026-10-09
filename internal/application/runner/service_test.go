package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
)

func TestService_Run_ReturnsFatalErrorAndStopsAtFailingStage(t *testing.T) {
	tests := newFailingStageTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			fixture := newRunFixture()
			fatal := newFatalError(test.name)
			fixture.expectRunFailingAt(test.failingStage, fatal)

			// When
			err := fixture.service.Run()

			// Then
			assert.ErrorIs(t, err, fatal)
			assert.Equal(t, []string{test.name}, fixture.logger.Errored)
		})
	}
}

func TestService_Run_ContinuesAfterNonFatalConfigLoadError(t *testing.T) {
	// Given
	fixture := newRunFixture()
	loadErr := newWarning("config store not loaded")
	fixture.config.On("Load").Return(fixture.configResult, loadErr)
	fixture.expectSessionAfterConfig(game.PlayResult{})

	// When
	err := fixture.service.Run()

	// Then
	require.NoError(t, err)
	assert.Equal(t, []string{"config store not loaded"}, fixture.logger.Warned)
	fixture.game.AssertCalled(t, "Play", mock.Anything)
}

func TestService_Run_ContinuesAfterNonFatalRecordScoreError(t *testing.T) {
	// Given
	fixture := newRunFixture()
	recordErr := newWarning("score board not saved")
	fixture.config.On("Load").Return(fixture.configResult, nil)
	fixture.processes.On("FindProcesses", mock.Anything).Return(fixture.found, nil)
	fixture.scores.On("LoadScoreBoard").Return(fixture.loaded, nil)
	fixture.game.On("Play", mock.Anything).Return(game.PlayResult{}, nil)
	fixture.scores.On("RecordScore", mock.Anything).Return(score.RecordResult{}, recordErr)
	fixture.scores.On("ReportResults", mock.Anything).Return()

	// When
	err := fixture.service.Run()

	// Then
	require.NoError(t, err)
	assert.Equal(t, []string{"score board not saved"}, fixture.logger.Warned)
	fixture.scores.AssertCalled(t, "ReportResults", mock.Anything)
}

func TestService_Run_PassesEachStepsResultToTheNext(t *testing.T) {
	// Given
	fixture := newRunFixture()
	fixture.configResult = config.Result{
		Process: config.ProcessResult{IncludeRoot: true},
		Game:    config.GameResult{ConfirmMode: true, Speed: 3.5, TimeLimit: 45},
	}
	loaded := score.LoadResult{Entries: []score.BoardEntry{{Kills: 4}}, HighScore: 4}
	result := game.PlayResult{Duration: 12.5, LowestSpeed: 2.5, Kills: 3, Duds: 1, FreedMem: 8192}
	recorded := score.RecordResult{Entries: []score.BoardEntry{{Kills: 4}, {Kills: 3}}, NewHighScore: true}
	fixture.config.On("Load").Return(fixture.configResult, nil)
	fixture.processes.On("FindProcesses", process.FindRequest{IncludeRoot: true}).Return(fixture.found, nil)
	fixture.scores.On("LoadScoreBoard").Return(loaded, nil)
	fixture.game.On("Play", game.PlayRequest{
		ConfirmMode: true, Speed: 3.5, TimeLimit: 45,
		Processes: []game.ProcessRequest{{PID: 200, Name: "target", Rss: 1024, UID: 1000}},
		HighScore: 4,
	}).Return(result, nil)
	fixture.scores.On("RecordScore", score.RecordRequest{
		Entries: loaded.Entries, Kills: 3, Duds: 1, FreedMem: 8192, LowestSpeed: 2.5, TimeLimit: 45, Duration: 12.5,
	}).Return(recorded, nil)
	fixture.scores.On("ReportResults", score.ReportRequest{
		Duration: 12.5, Kills: 3, Duds: 1, FreedMem: 8192, Entries: recorded.Entries, NewHighScore: true,
	}).Return()

	// When
	err := fixture.service.Run()

	// Then
	require.NoError(t, err)
	fixture.config.AssertExpectations(t)
	fixture.processes.AssertExpectations(t)
	fixture.scores.AssertExpectations(t)
	fixture.game.AssertExpectations(t)
}

func TestService_Run_ContinuesAfterNonFatalScoreBoardLoadError(t *testing.T) {
	// Given
	fixture := newRunFixture()
	loadErr := newWarning("score board not loaded")
	fixture.config.On("Load").Return(fixture.configResult, nil)
	fixture.processes.On("FindProcesses", mock.Anything).Return(fixture.found, nil)
	fixture.scores.On("LoadScoreBoard").Return(fixture.loaded, loadErr)
	fixture.game.On("Play", mock.Anything).Return(game.PlayResult{}, nil)
	fixture.scores.On("RecordScore", mock.Anything).Return(score.RecordResult{}, nil)
	fixture.scores.On("ReportResults", mock.Anything).Return()

	// When
	err := fixture.service.Run()

	// Then
	require.NoError(t, err)
	assert.Equal(t, []string{"score board not loaded"}, fixture.logger.Warned)
	fixture.scores.AssertExpectations(t)
}

func TestService_Run_HandlesEachPlayErrorWithoutFailing(t *testing.T) {
	// Given
	fixture := newRunFixture()
	fixture.config.On("Load").Return(fixture.configResult, nil)
	fixture.expectSessionAfterConfig(game.PlayResult{
		Errors: []error{newWarning("could not kill stubborn (PID 200)"), newWarning("gone (PID 300) ran away")},
	})

	// When
	err := fixture.service.Run()

	// Then
	require.NoError(t, err)
	assert.Equal(t, []string{"could not kill stubborn (PID 200)", "gone (PID 300) ran away"}, fixture.logger.Warned)
	assert.Empty(t, fixture.logger.Errored)
}

func newFailingStageTestCases() []struct {
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
