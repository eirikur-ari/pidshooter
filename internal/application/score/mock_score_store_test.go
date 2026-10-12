package score

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

type MockScoreStore struct{ mock.Mock }

func (m *MockScoreStore) Load() (outbound.ScoreBoard, error) {
	args := m.Called()
	board, _ := args.Get(0).(outbound.ScoreBoard)
	return board, args.Error(1)
}

func (m *MockScoreStore) Save(board outbound.ScoreBoard) error {
	return m.Called(board).Error(0)
}
