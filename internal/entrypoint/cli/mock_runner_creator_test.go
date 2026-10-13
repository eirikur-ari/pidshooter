package cli

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

type MockRunnerCreator struct{ mock.Mock }

func (m *MockRunnerCreator) Create(patterns []string, options config.Options) (inbound.Runner, error) {
	args := m.Called(patterns, options)
	runner, _ := args.Get(0).(inbound.Runner)
	return runner, args.Error(1)
}

func (m *MockRunnerCreator) ErrHandler() *apperror.Handler {
	return apperror.NewHandler(&testutil.FakeLogger{})
}
