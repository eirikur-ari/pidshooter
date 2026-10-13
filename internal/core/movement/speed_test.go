package movement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpeed_Set_TracksCurrentAndLowestValue(t *testing.T) {
	tests := newSpeedSetTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			s := newSpeed(2.0)

			// When
			for _, value := range test.speeds {
				s.Set(value)
			}

			// Then
			assert.Equal(t, test.expectedCurrent, s.current)
			assert.Equal(t, test.expectedLowest, s.lowest)
		})
	}
}

func newSpeedSetTestCase() []struct {
	name            string
	speeds          []float64
	expectedCurrent float64
	expectedLowest  float64
} {
	return []struct {
		name            string
		speeds          []float64
		expectedCurrent float64
		expectedLowest  float64
	}{
		{"starts with lowest equal to current", nil, 2.0, 2.0},
		{"a lower value becomes the new lowest", []float64{1.5}, 1.5, 1.5},
		{"a value equal to the lowest changes nothing", []float64{2.0}, 2.0, 2.0},
		{"a higher value than the start leaves the lowest alone", []float64{3.0}, 3.0, 2.0},
		{"raising after lowering keeps the earlier lowest", []float64{1.5, 1.8}, 1.8, 1.5},
		{"lowering twice keeps the smallest", []float64{1.5, 1.0}, 1.0, 1.0},
	}
}
