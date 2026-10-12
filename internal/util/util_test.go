package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatBytes_UsesLargestFittingUnit(t *testing.T) {
	tests := newFormatBytesTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			formatted := FormatBytes(test.bytes)

			// Then
			assert.Equal(t, test.expected, formatted)
		})
	}
}

func TestClonePointer_ReturnsNilWhenGivenNil(t *testing.T) {
	// Given
	var original *int

	// When
	clone := ClonePointer(original)

	// Then
	assert.Nil(t, clone)
}

func TestClonePointer_ReturnsDistinctPointerWithSameValue(t *testing.T) {
	// Given
	value := 42
	original := &value

	// When
	clone := ClonePointer(original)

	// Then
	require.NotNil(t, clone)
	assert.Equal(t, *original, *clone)
	assert.NotSame(t, original, clone, "the clone must not alias the original")
}

func newFormatBytesTestCase() []struct {
	name     string
	bytes    int64
	expected string
} {
	return []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero", 0, "0 B"},
		{"below a kilobyte", 512, "512 B"},
		{"one byte short of a kilobyte", 1023, "1023 B"},
		{"exactly a kilobyte", 1024, "1.0 KB"},
		{"fractional kilobytes", 1536, "1.5 KB"},
		{"last size that still rounds within kilobytes", 1048524, "1023.9 KB"},
		{"first size that would round to 1024.0 KB moves up to megabytes", 1048525, "1.0 MB"},
		{"fractional megabytes", 1572864, "1.5 MB"},
		{"last size that still rounds within megabytes", 1073689395, "1023.9 MB"},
		{"first size that would round to 1024.0 MB moves up to gigabytes", 1073689396, "1.0 GB"},
		{"fractional gigabytes", 1610612736, "1.5 GB"},
		{"sizes above gigabytes stay in gigabytes", 1099511627776, "1024.0 GB"},
	}
}
