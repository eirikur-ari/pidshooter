package cli

import "github.com/stretchr/testify/mock"

type MockRunner struct{ mock.Mock }

func (m *MockRunner) Run() error {
	return m.Called().Error(0)
}
