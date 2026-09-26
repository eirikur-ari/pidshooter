package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{1610612736, "1.5 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatBytes(tt.input))
		})
	}
}

func TestClonePtrNilReturnsNil(t *testing.T) {
	var p *int
	assert.Nil(t, ClonePtr(p))
}

func TestClonePtrReturnsDistinctPointerWithSameValue(t *testing.T) {
	v := 42
	p := &v

	clone := ClonePtr(p)

	require.NotNil(t, clone)
	assert.Equal(t, *p, *clone)
	assert.NotSame(t, p, clone, "the clone must not alias the original")
}
