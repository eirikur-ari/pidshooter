package composition

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func newCreatorFixture() *Creator {
	handler, _ := newHandlerFixture()
	return &Creator{errHandler: handler}
}

func newLazyRunnerFixture(
	build func() (inbound.PlayRunner, error),
	handler *apperror.Handler,
) *lazyRunner[inbound.PlayRequest, inbound.PlayRunner] {
	return &lazyRunner[inbound.PlayRequest, inbound.PlayRunner]{build: build, errHandler: handler}
}

func newBuildFixture(runner inbound.PlayRunner, err error) func() (inbound.PlayRunner, error) {
	return func() (inbound.PlayRunner, error) { return runner, err }
}

func newCountingBuildFixture(runner inbound.PlayRunner) (func() (inbound.PlayRunner, error), *int) {
	builds := 0
	build := func() (inbound.PlayRunner, error) {
		builds++
		return runner, nil
	}
	return build, &builds
}

func newWarningFixture(message string) *apperror.Error {
	return apperror.NewError(apperror.CodeUnknown, apperror.SeverityWarning, message, nil)
}

func newHandlerFixture() (*apperror.Handler, *testutil.FakeLogger) {
	logger := &testutil.FakeLogger{}
	return apperror.NewHandler(logger), logger
}
