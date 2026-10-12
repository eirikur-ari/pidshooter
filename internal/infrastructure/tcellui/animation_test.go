package tcellui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnimation_frame_SelectsFrameByProgress(t *testing.T) {
	tests := newFrameTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			anim := animation{"a", "b", "c"}

			// When
			frame := anim.frame(test.progress)

			// Then
			assert.Equal(t, test.expected, frame)
		})
	}
}

func newFrameTestCase() []struct {
	name     string
	progress float64
	expected string
} {
	return []struct {
		name     string
		progress float64
		expected string
	}{
		{"zero progress is the first frame", 0, "a"},
		{"mid progress is the middle frame", 0.5, "b"},
		{"just before the end is the last frame", 0.99, "c"},
		{"full progress clamps to the last frame", 1.0, "c"},
		{"progress past the end clamps to the last frame", 2.0, "c"},
		{"negative progress clamps to the first frame", -0.5, "a"},
	}
}
