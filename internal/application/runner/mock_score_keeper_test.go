package runner

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/score"
)

// MockScoreKeeper is a test double for scoreKeeper.
type MockScoreKeeper struct{ mock.Mock }

func (m *MockScoreKeeper) LoadScoreBoard() (score.LoadResult, error) {
	args := m.Called()
	return args.Get(0).(score.LoadResult), args.Error(1)
}

func (m *MockScoreKeeper) RecordScore(request score.RecordRequest) (score.RecordResult, error) {
	args := m.Called(request)
	return args.Get(0).(score.RecordResult), args.Error(1)
}

func (m *MockScoreKeeper) ReportResults(request score.ReportRequest) {
	m.Called(request)
}
