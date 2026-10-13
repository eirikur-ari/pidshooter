package movement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVector_Rounded_RoundsEachComponentHalfAwayFromZero(t *testing.T) {
	tests := newVectorRoundedTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			x, y := test.input.Rounded()

			// Then
			assert.Equal(t, test.expectedX, x)
			assert.Equal(t, test.expectedY, y)
		})
	}
}

func newVectorRoundedTestCase() []struct {
	name      string
	input     Vector
	expectedX int
	expectedY int
} {
	return []struct {
		name      string
		input     Vector
		expectedX int
		expectedY int
	}{
		{"below half rounds down", Vector{X: 2.4, Y: 5.4}, 2, 5},
		{"above half rounds up", Vector{X: 10.6, Y: 5.6}, 11, 6},
		{"exactly half rounds up", Vector{X: 0.5, Y: 1.5}, 1, 2},
		{"negative half rounds away from zero", Vector{X: -0.5, Y: -1.5}, -1, -2},
		{"whole numbers are unchanged", Vector{X: 3, Y: 7}, 3, 7},
	}
}
