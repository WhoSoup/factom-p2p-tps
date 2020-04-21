package app

import (
	"math"
	"testing"
)

func TestGenerator_WeightedRandomType(t *testing.T) {
	const RUNS = 10000

	tests := []struct {
		name string
		prct map[byte]float64
	}{
		{"even", map[byte]float64{0: .5, 1: .5}},
		{"even, non 1 prob", map[byte]float64{0: 2, 1: 2}},
		{"33% to 66%", map[byte]float64{0: .5, 1: 1}},
		{"33% to 66%, non 1 prob", map[byte]float64{0: 6, 1: 9}},
		{"even 3", map[byte]float64{0: .5, 1: .5, 2: .5}},
		{"10% 15% 75%", map[byte]float64{0: 10, 1: 15, 2: 75}},
		{"75% 15% 10%", map[byte]float64{0: 75, 1: 15, 2: 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gen := NewGenerator(tt.prct)

			results := make(map[byte]int)
			for i := 0; i < RUNS; i++ {
				results[gen.WeightedRandomType()]++
			}

			for k, v := range tt.prct {
				rate := float64(results[k]) / RUNS
				want := v / gen.entryRange
				diff := math.Abs(want - rate)
				if diff > 0.01 {
					t.Errorf("Results[%v] = %v, want %v, diff %v", k, rate, want, diff)
				}
			}

		})
	}
}
