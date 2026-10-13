package score

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func scoreStoreMockSetup(persisted outbound.ScoreBoard, loadErr, saveErr error) *MockScoreStore {
	store := &MockScoreStore{}
	store.On("Load").Return(persisted, loadErr)
	store.On("Save", mock.Anything).Return(saveErr)
	return store
}

func boardWithScores(count int) any {
	return mock.MatchedBy(func(board outbound.ScoreBoard) bool {
		return len(board.Scores) == count
	})
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func reporterMockSetup() *MockScoreReporter {
	reporter := &MockScoreReporter{}
	reporter.On("Report", mock.Anything).Return()
	return reporter
}
