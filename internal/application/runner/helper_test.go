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

type playMockSetup struct {
	play         *Play
	config       *MockConfigLoader
	processes    *MockProcessFinder
	scores       *MockScoreKeeper
	game         *MockGamePlayer
	logger       *testutil.FakeLogger
	configResult config.Result
	found        []process.FindResult
	loaded       score.LoadResult
}

func newPlayMockSetup() *playMockSetup {
	setup := &playMockSetup{
		config:    &MockConfigLoader{},
		processes: &MockProcessFinder{},
		scores:    &MockScoreKeeper{},
		game:      &MockGamePlayer{},
		logger:    &testutil.FakeLogger{},
		found:     []process.FindResult{{PID: 200, Name: "target", Rss: 1024, UID: 1000}},
		loaded:    score.LoadResult{},
	}
	setup.play = NewPlay(setup.config, setup.processes, setup.scores, setup.game, apperror.NewHandler(setup.logger))
	return setup
}

// expectSessionAfterConfig expects every step after the configuration load to succeed.
func (m *playMockSetup) expectSessionAfterConfig(result game.PlayResult) {
	m.expectSession(result, nil, nil)
}

// expectRunWarningAt expects every step to run, with the given step returning
// err. Only a step whose error is not necessarily fatal can return a warning.
func (m *playMockSetup) expectRunWarningAt(warning stage, err error) {
	var configErr, loadErr, recordErr error
	switch warning {
	case stageConfigLoad:
		configErr = err
	case stageScoreBoardLoad:
		loadErr = err
	case stageScoreRecording:
		recordErr = err
	default:
		panic("stage cannot return a warning")
	}
	m.config.On("Load", mock.Anything).Return(m.configResult, configErr)
	m.expectSession(game.PlayResult{}, loadErr, recordErr)
}

// expectRunFailingAt expects every step before failing to succeed, failing to
// return err, and no step after it to be called.
func (m *playMockSetup) expectRunFailingAt(failing stage, err error) {
	if failing == stageConfigLoad {
		m.config.On("Load", mock.Anything).Return(config.Result{}, err)
		return
	}
	m.config.On("Load", mock.Anything).Return(m.configResult, nil)

	if failing == stageProcessSearch {
		m.processes.On("FindProcesses", mock.Anything).Return(nil, err)
		return
	}
	m.processes.On("FindProcesses", mock.Anything).Return(m.found, nil)

	if failing == stageScoreBoardLoad {
		m.scores.On("LoadScoreBoard").Return(m.loaded, err)
		return
	}
	m.scores.On("LoadScoreBoard").Return(m.loaded, nil)

	if failing == stageGamePlay {
		m.game.On("Play", mock.Anything).Return(game.PlayResult{}, err)
		return
	}
	if failing == stageGamePlayResult {
		m.game.On("Play", mock.Anything).Return(game.PlayResult{Errors: []error{err}}, nil)
		return
	}
	m.game.On("Play", mock.Anything).Return(game.PlayResult{}, nil)

	m.scores.On("RecordScore", mock.Anything).Return(score.RecordResult{}, err)
}

func (m *playMockSetup) expectSession(result game.PlayResult, loadErr, recordErr error) {
	m.processes.On("FindProcesses", mock.Anything).Return(m.found, nil)
	m.scores.On("LoadScoreBoard").Return(m.loaded, loadErr)
	m.game.On("Play", mock.Anything).Return(result, nil)
	m.scores.On("RecordScore", mock.Anything).Return(score.RecordResult{}, recordErr)
	m.scores.On("ReportResults", mock.Anything).Return()
}
