package movement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpeed_Set_TracksCurrentAndLowestValue(t *testing.T) {
	tests := []struct {
		name    string
		speeds  []float64
		current float64
		lowest  float64
	}{
		{name: "starts with lowest equal to current", speeds: nil, current: 2.0, lowest: 2.0},
		{name: "a lower value becomes the new lowest", speeds: []float64{1.5}, current: 1.5, lowest: 1.5},
		{name: "a value equal to the lowest changes nothing", speeds: []float64{2.0}, current: 2.0, lowest: 2.0},
		{name: "a higher value than the start leaves the lowest alone", speeds: []float64{3.0}, current: 3.0, lowest: 2.0},
		{name: "raising after lowering keeps the earlier lowest", speeds: []float64{1.5, 1.8}, current: 1.8, lowest: 1.5},
		{name: "lowering twice keeps the smallest", speeds: []float64{1.5, 1.0}, current: 1.0, lowest: 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			s := newSpeed(2.0)

			// When
			for _, value := range tt.speeds {
				s.Set(value)
			}

			// Then
			assert.Equal(t, tt.current, s.Current())
			assert.Equal(t, tt.lowest, s.Lowest())
		})
	}
}
