package apperror

import "errors"

func newErrorFixture() *Error {
	return NewError(CodeUnknown, SeverityUnknown, "message", errors.New("cause"))
}

func newErrorFixtureFor(severity Severity) *Error {
	err := newErrorFixture()
	err.Severity = severity
	return err
}
