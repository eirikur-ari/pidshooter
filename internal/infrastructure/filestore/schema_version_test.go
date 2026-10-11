package filestore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaVersion_validate_ReturnsErrorForNegativeOrNewerVersion(t *testing.T) {
	tests := newInvalidSchemaVersionTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			current := schemaVersion(1)

			// When
			err := current.validate(test.fileType, test.version)

			// Then
			assert.EqualError(t, err, test.expected)
		})
	}
}

func TestSchemaVersion_validate_ReturnsNoErrorFromZeroThroughCurrentVersion(t *testing.T) {
	tests := newValidSchemaVersionTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			current := schemaVersion(1)

			// When
			err := current.validate("config file", test.version)

			// Then
			assert.NoError(t, err)
		})
	}
}

func newInvalidSchemaVersionTestCases() []struct {
	name     string
	fileType string
	version  int
	expected string
} {
	return []struct {
		name     string
		fileType string
		version  int
		expected string
	}{
		{"negative", "config file", -5, "config file schema version -5 is invalid"},
		{"newer than current", "score file", 2, "score file schema version 2 is newer than the 1 this build supports"},
	}
}

func newValidSchemaVersionTestCases() []struct {
	name    string
	version int
} {
	return []struct {
		name    string
		version int
	}{
		{"zero", 0},
		{"current", 1},
	}
}
