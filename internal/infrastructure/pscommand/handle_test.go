package pscommand

import (
	"errors"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestHandle_Kill_MapsSignalErrors(t *testing.T) {
	tests := newKillTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			process := &MockOsProcess{}
			process.On("Signal", syscall.SIGKILL).Return(test.signalErr)
			processHandle := &handle{process: process}

			// When
			err := processHandle.Kill()

			// Then
			assert.Equal(t, test.expected, err)
			process.AssertExpectations(t)
		})
	}
}

func TestHandle_Release_ReturnsReleaseResult(t *testing.T) {
	tests := newReleaseTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			process := &MockOsProcess{}
			process.On("Release").Return(test.expected)
			processHandle := &handle{process: process}

			// When
			err := processHandle.Release()

			// Then
			assert.Equal(t, test.expected, err)
			process.AssertExpectations(t)
		})
	}
}

func TestHandle_verify_ReportsWhetherProcessStillExists(t *testing.T) {
	tests := newVerifyTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			process := &MockOsProcess{}
			process.On("Signal", syscall.Signal(0)).Return(test.signalErr)
			if test.releases > 0 {
				process.On("Release").Return(nil)
			}
			processHandle := &handle{process: process}

			// When
			err := processHandle.verify()

			// Then
			assert.Equal(t, test.expected, err)
			process.AssertExpectations(t)
			process.AssertNumberOfCalls(t, "Release", test.releases)
		})
	}
}

func newKillTestCase() []struct {
	name      string
	signalErr error
	expected  error
} {
	otherErr := errors.New("operation not permitted")

	return []struct {
		name      string
		signalErr error
		expected  error
	}{
		{"process is terminated", nil, nil},
		{"process is already gone", os.ErrProcessDone, outbound.NotFoundError{}},
		{"other signal errors are returned as they are", otherErr, otherErr},
	}
}

func newReleaseTestCase() []struct {
	name     string
	expected error
} {
	return []struct {
		name     string
		expected error
	}{
		{"release succeeds", nil},
		{"release fails", errors.New("boom")},
	}
}

func newVerifyTestCase() []struct {
	name      string
	signalErr error
	releases  int
	expected  error
} {
	otherErr := errors.New("operation not permitted")

	return []struct {
		name      string
		signalErr error
		releases  int
		expected  error
	}{
		{"process exists", nil, 0, nil},
		{"process is gone and is released", os.ErrProcessDone, 1, outbound.NotFoundError{}},
		{"other signal errors do not mean the process is gone", otherErr, 0, nil},
	}
}
