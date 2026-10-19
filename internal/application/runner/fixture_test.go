package runner

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func newFatalError(message string) *apperror.Error {
	return apperror.NewError(apperror.CodeUnknown, apperror.SeverityFatal, message, nil)
}

func newWarning(message string) *apperror.Error {
	return apperror.NewError(apperror.CodeUnknown, apperror.SeverityWarning, message, nil)
}

func newPlayRequestFixture() inbound.PlayRequest {
	return inbound.PlayRequest{Patterns: []string{"target"}}
}

func newDispatcherFixture() (dispatcher *Dispatcher, play *MockPlayRunner, list *MockListRunner, kill *MockKillRunner) {
	play, list, kill = &MockPlayRunner{}, &MockListRunner{}, &MockKillRunner{}
	dispatcher = NewDispatcher(play, list, kill, apperror.NewHandler(&testutil.FakeLogger{}))
	return dispatcher, play, list, kill
}
