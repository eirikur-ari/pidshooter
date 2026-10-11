package filestore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileEncoder_encodeJSON_IndentsWithTwoSpaces(t *testing.T) {
	// Given
	encoder := fileEncoder{}
	value := map[string]int{"kills": 5}
	expected := "{\n  \"kills\": 5\n}"

	// When
	data, err := encoder.encodeJSON(value)

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, string(data))
}

func TestFileEncoder_encodeYAML_EncodesValueAsYAML(t *testing.T) {
	// Given
	encoder := fileEncoder{}
	value := map[string]int{"kills": 5}
	expected := "kills: 5\n"

	// When
	data, err := encoder.encodeYAML(value)

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, string(data))
}
