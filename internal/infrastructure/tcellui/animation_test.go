package tcellui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnimationFrameAtZeroProgress(t *testing.T) {
	a := animation{"a", "b", "c"}
	assert.Equal(t, "a", a.frame(0))
}

func TestAnimationFrameAtMidProgress(t *testing.T) {
	a := animation{"a", "b", "c"}
	assert.Equal(t, "b", a.frame(0.5))
}

func TestAnimationFrameClampsAtUpperBound(t *testing.T) {
	a := animation{"a", "b", "c"}
	assert.Equal(t, "c", a.frame(1.0))
}

func TestAnimationFrameClampsAtNegativeProgress(t *testing.T) {
	a := animation{"a", "b", "c"}
	assert.Equal(t, "a", a.frame(-0.5))
}
