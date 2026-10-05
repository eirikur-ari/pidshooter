package apperror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestHandler_Handle_ReturnsOrAbsorbsBySeverity(t *testing.T) {
	tests := newHandleReturnedTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			handler := NewHandler(&testutil.FakeLogger{})

			// When
			err := handler.Handle(test.err)

			// Then
			assert.Equal(t, test.expected, err)
		})
	}
}

func TestHandler_Handle_LogsAsWarningBySeverity(t *testing.T) {
	tests := newHandleWarnedTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			logger := &testutil.FakeLogger{}
			handler := NewHandler(logger)

			// When
			_ = handler.Handle(test.err)

			// Then
			assert.Equal(t, test.expected, logger.Warned)
		})
	}
}

func TestHandler_Handle_LogsAsErrorBySeverity(t *testing.T) {
	tests := newHandleErroredTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			logger := &testutil.FakeLogger{}
			handler := NewHandler(logger)

			// When
			_ = handler.Handle(test.err)

			// Then
			assert.Equal(t, test.expected, logger.Errored)
		})
	}
}

func TestHandler_Handle_LogsAnErrorOnlyOnce(t *testing.T) {
	tests := []struct {
		name   string
		newErr func() error
	}{
		{"error", func() error { return newErrorFixtureFor(SeverityFatal) }},
		{"wrapped error", func() error { return fmt.Errorf("wrapped: %w", newErrorFixtureFor(SeverityFatal)) }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			err := test.newErr()
			logger := &testutil.FakeLogger{}
			handler := NewHandler(logger)
			first := handler.Handle(err)

			// When
			second := handler.Handle(err)

			// Then
			assert.Same(t, err, first)
			assert.Same(t, err, second)
			assert.Len(t, logger.Errored, 1)
		})
	}
}

func newHandleReturnedTestCase() []struct {
	name     string
	err      error
	expected error
} {
	warning := newErrorFixtureFor(SeverityWarning)
	nonFatal := newErrorFixtureFor(SeverityError)
	fatal := newErrorFixtureFor(SeverityFatal)
	unknown := newErrorFixture()
	plain := errors.New("plain error")
	wrappedWarning := fmt.Errorf("wrapped: %w", newErrorFixtureFor(SeverityWarning))
	wrappedFatal := fmt.Errorf("wrapped: %w", newErrorFixtureFor(SeverityFatal))

	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{"warning severity is absorbed", warning, nil},
		{"error severity is absorbed", nonFatal, nil},
		{"fatal severity is returned", fatal, fatal},
		{"unknown severity is returned", unknown, unknown},
		{"not an Error is returned", plain, plain},
		{"wrapped warning is absorbed", wrappedWarning, nil},
		{"wrapped fatal is returned", wrappedFatal, wrappedFatal},
		{"nil error is absorbed", nil, nil},
	}
	return tests
}

func newHandleWarnedTestCase() []struct {
	name     string
	err      error
	expected []string
} {
	warning := newErrorFixtureFor(SeverityWarning)
	wrappedWarning := fmt.Errorf("wrapped: %w", newErrorFixtureFor(SeverityWarning))

	tests := []struct {
		name     string
		err      error
		expected []string
	}{
		{"warning severity", warning, []string{warning.Error()}},
		{"wrapped warning", wrappedWarning, []string{wrappedWarning.Error()}},
		{"error severity", newErrorFixtureFor(SeverityError), nil},
		{"fatal severity", newErrorFixtureFor(SeverityFatal), nil},
		{"unknown severity", newErrorFixture(), nil},
		{"not an Error", errors.New("plain error"), nil},
		{"nil error", nil, nil},
	}
	return tests
}

func newHandleErroredTestCase() []struct {
	name     string
	err      error
	expected []string
} {
	nonFatal := newErrorFixtureFor(SeverityError)
	fatal := newErrorFixtureFor(SeverityFatal)
	unknown := newErrorFixture()
	plain := errors.New("plain error")
	wrappedFatal := fmt.Errorf("wrapped: %w", newErrorFixtureFor(SeverityFatal))

	tests := []struct {
		name     string
		err      error
		expected []string
	}{
		{"error severity", nonFatal, []string{nonFatal.Error()}},
		{"fatal severity", fatal, []string{fatal.Error()}},
		{"unknown severity", unknown, []string{unknown.Error()}},
		{"not an Error", plain, []string{plain.Error()}},
		{"wrapped fatal", wrappedFatal, []string{wrappedFatal.Error()}},
		{"warning severity", newErrorFixtureFor(SeverityWarning), nil},
		{"wrapped warning", fmt.Errorf("wrapped: %w", newErrorFixtureFor(SeverityWarning)), nil},
		{"nil error", nil, nil},
	}
	return tests
}
