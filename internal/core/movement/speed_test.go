package movement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSpeed(t *testing.T) {
	s := newSpeed(2.0)
	assert.Equal(t, 2.0, s.Current())
	assert.Equal(t, 2.0, s.Lowest())
}

func TestSpeedSetTracksLowest(t *testing.T) {
	s := newSpeed(2.0)
	s.Set(1.5)
	assert.Equal(t, 1.5, s.Current())
	assert.Equal(t, 1.5, s.Lowest())
}

func TestSpeedSetAboveLowestDoesNotRaiseLowest(t *testing.T) {
	s := newSpeed(2.0)
	s.Set(1.5)
	s.Set(1.8)
	assert.Equal(t, 1.8, s.Current())
	assert.Equal(t, 1.5, s.Lowest(), "expected lowest to remain the smallest value ever set")
}

func TestSpeedSetAboveStartNeverLowersLowest(t *testing.T) {
	s := newSpeed(2.0)
	s.Set(3.0)
	assert.Equal(t, 3.0, s.Current())
	assert.Equal(t, 2.0, s.Lowest())
}
