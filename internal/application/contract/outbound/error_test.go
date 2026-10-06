package outbound

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundError_Error_ReturnsFixedText(t *testing.T) {
	assert.Equal(t, "not found", NotFoundError{}.Error())
}

func TestCorruptedDataError_Error_Texts(t *testing.T) {
	tests := newCorruptedDataErrorTextTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := test.err.Error()

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func newCorruptedDataErrorTextTestCases() []struct {
	name     string
	err      CorruptedDataError
	expected string
} {
	return []struct {
		name     string
		err      CorruptedDataError
		expected string
	}{
		{"message is empty", CorruptedDataError{}, "corrupted data"},
		{"message has text", CorruptedDataError{Message: "file is too large"}, "file is too large: corrupted data"},
	}
}
