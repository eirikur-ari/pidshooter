package process

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

type MockProcessReporter struct{ mock.Mock }

func (m *MockProcessReporter) Report(matches []outbound.ProcessInfo, patterns []string) {
	m.Called(matches, patterns)
}
