package pscommand

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestMapError_MapsPsExitFailures(t *testing.T) {
	tests := newMapErrorTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := mapError(test.exitCode, test.stderr, test.cause)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func newMapErrorTestCase() []struct {
	name     string
	exitCode int
	stderr   []byte
	cause    error
	expected error
} {
	cause := errors.New("exit status 1")
	otherCause := errors.New("exit status 2")

	return []struct {
		name     string
		exitCode int
		stderr   []byte
		cause    error
		expected error
	}{
		{"silent exit 1 means no process matched", 1, nil, cause, outbound.NotFoundError{}},
		{"stderr text is appended to the failure", 1, []byte("ps: bad option\n"), cause, fmt.Errorf("%w: ps: bad option", cause)},
		{"stderr text is appended for other exit codes too", 2, []byte("ps: bad option"), otherCause, fmt.Errorf("%w: ps: bad option", otherCause)},
		{"other exit codes without stderr return the failure as is", 2, nil, otherCause, otherCause},
	}
}
