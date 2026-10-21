package apperror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCode_In_ReportsWhetherErrorIsAnAppErrorWithCode(t *testing.T) {
	nested := NewError(CodeKillFailed, SeverityWarning, "could not kill target",
		NewError(CodeInvalidConfig, SeverityFatal, "invalid configuration", nil))

	tests := []struct {
		name     string
		err      error
		code     Code
		expected bool
	}{
		{"matching code", NewError(CodeInvalidConfig, SeverityFatal, "invalid configuration", nil), CodeInvalidConfig, true},
		{"different code", NewError(CodeGameFailed, SeverityFatal, "game session failed", nil), CodeInvalidConfig, false},
		{"wrapped matching error", fmt.Errorf("wrapped: %w", NewError(CodeInvalidConfig, SeverityFatal, "invalid configuration", nil)), CodeInvalidConfig, true},
		{"plain error", errors.New("boom"), CodeInvalidConfig, false},
		{"nil error", nil, CodeInvalidConfig, false},
		{"outer code of nested errors", nested, CodeKillFailed, true},
		{"cause code of nested errors", nested, CodeInvalidConfig, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := test.code.In(test.err)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}
