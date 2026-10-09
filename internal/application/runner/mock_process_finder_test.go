package runner

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/process"
)

// MockProcessFinder is a test double for processFinder.
type MockProcessFinder struct{ mock.Mock }

func (m *MockProcessFinder) FindProcesses(request process.FindRequest) ([]process.FindResult, error) {
	args := m.Called(request)
	found, _ := args.Get(0).([]process.FindResult)
	return found, args.Error(1)
}
