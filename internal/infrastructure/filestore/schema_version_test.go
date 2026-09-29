package filestore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaVersionValidateRejectsNegative(t *testing.T) {
	err := schemaVersion(1).validate("config file", -5)

	assert.ErrorContains(t, err, "config file")
	assert.ErrorContains(t, err, "-5")
	assert.ErrorContains(t, err, "invalid")
}

func TestSchemaVersionValidateRejectsNewerThanCurrent(t *testing.T) {
	err := schemaVersion(1).validate("score file", 2)

	assert.ErrorContains(t, err, "score file")
	assert.ErrorContains(t, err, "2")
	assert.ErrorContains(t, err, "newer")
}

func TestSchemaVersionValidateAcceptsZeroThroughCurrent(t *testing.T) {
	for got := 0; got <= 1; got++ {
		assert.NoError(t, schemaVersion(1).validate("config file", got))
	}
}
