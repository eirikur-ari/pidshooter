package runner

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/game"
)

// MockGamePlayer is a test double for gamePlayer.
type MockGamePlayer struct{ mock.Mock }

func (m *MockGamePlayer) Play(request game.PlayRequest) (game.PlayResult, error) {
	args := m.Called(request)
	return args.Get(0).(game.PlayResult), args.Error(1)
}
