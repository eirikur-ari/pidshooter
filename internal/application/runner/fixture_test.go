package runner

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

// stage is a step of Run.
type stage int

const (
	stageConfigLoad stage = iota
	stageProcessSearch
	stageScoreBoardLoad
	stageGamePlay
	stageGamePlayResult
	stageScoreRecording
)

type runFixture struct {
	service      *Service
	config       *MockConfigLoader
	processes    *MockProcessFinder
	scores       *MockScoreKeeper
	game         *MockGamePlayer
	logger       *testutil.FakeLogger
	configResult config.Result
	found        []process.FindResult
	board        score.LoadResult
}

func newRunFixture() *runFixture {
	fixture := &runFixture{
		config:    &MockConfigLoader{},
		processes: &MockProcessFinder{},
		scores:    &MockScoreKeeper{},
		game:      &MockGamePlayer{},
		logger:    &testutil.FakeLogger{},
		found:     []process.FindResult{{PID: 200, Name: "target", Rss: 1024, UID: 1000}},
		board:     score.LoadResult{},
	}
	fixture.service = NewService(fixture.config, fixture.processes, fixture.scores, fixture.game, apperror.NewHandler(fixture.logger))
	return fixture
}

// expectSessionAfterConfig expects every step after the configuration load to succeed.
func (f *runFixture) expectSessionAfterConfig(result game.PlayResult) {
	f.processes.On("FindProcesses", mock.Anything).Return(f.found, nil)
	f.scores.On("LoadScoreBoard").Return(f.board, nil)
	f.game.On("Play", mock.Anything).Return(result, nil)
	f.scores.On("RecordScore", mock.Anything, mock.Anything).Return(score.RecordResult{}, nil)
	f.scores.On("ReportResults", mock.Anything).Return()
}

// expectRunFailingAt expects every step before failing to succeed, failing to
// return err, and no step after it to be called.
func (f *runFixture) expectRunFailingAt(failing stage, err error) {
	if failing == stageConfigLoad {
		f.config.On("Load").Return(config.Result{}, err)
		return
	}
	f.config.On("Load").Return(f.configResult, nil)

	if failing == stageProcessSearch {
		f.processes.On("FindProcesses", mock.Anything).Return(nil, err)
		return
	}
	f.processes.On("FindProcesses", mock.Anything).Return(f.found, nil)

	if failing == stageScoreBoardLoad {
		f.scores.On("LoadScoreBoard").Return(f.board, err)
		return
	}
	f.scores.On("LoadScoreBoard").Return(f.board, nil)

	if failing == stageGamePlay {
		f.game.On("Play", mock.Anything).Return(game.PlayResult{}, err)
		return
	}
	if failing == stageGamePlayResult {
		f.game.On("Play", mock.Anything).Return(game.PlayResult{Errors: []error{err}}, nil)
		return
	}
	f.game.On("Play", mock.Anything).Return(game.PlayResult{}, nil)

	f.scores.On("RecordScore", mock.Anything, mock.Anything).Return(score.RecordResult{}, err)
}

func newFatalError(message string) *apperror.Error {
	return apperror.NewError(apperror.CodeUnknown, apperror.SeverityFatal, message, nil)
}

func newWarning(message string) *apperror.Error {
	return apperror.NewError(apperror.CodeUnknown, apperror.SeverityWarning, message, nil)
}
