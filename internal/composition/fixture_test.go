package composition

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func newCreatorFixture() *Creator {
	handler, _ := newHandlerFixture()
	return &Creator{errHandler: handler}
}

func newWarning(message string) *apperror.Error {
	return apperror.NewError(apperror.CodeUnknown, apperror.SeverityWarning, message, nil)
}

func newHandlerFixture() (*apperror.Handler, *testutil.FakeLogger) {
	logger := &testutil.FakeLogger{}
	return apperror.NewHandler(logger), logger
}
