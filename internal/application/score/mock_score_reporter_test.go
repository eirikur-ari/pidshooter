package score

import (
	"github.com/stretchr/testify/mock"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

type MockScoreReporter struct{ mock.Mock }

func (m *MockScoreReporter) Report(summary outbound.ScoreSummary) {
	m.Called(summary)
}
