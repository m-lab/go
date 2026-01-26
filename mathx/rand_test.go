package mathx

import (
	"testing"
)

func TestGetRandomInt(t *testing.T) {
	tests := []struct {
		name string
		max  int
	}{
		{
			name: "random",
			max:  10,
		},
		{
			name: "zero",
			max:  0,
		},
		{
			name: "negative",
			max:  -10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1 := GetRandomInt(tt.max)
			got2 := GetRandomInt(tt.max)

			if tt.max <= 0 {
				if got1 != 0 || got2 != 0 {
					t.Errorf("GetRandomInt(%d) should return 0, got %d and %d", tt.max, got1, got2)
				}
			} else {
				if got1 < 0 || got1 >= tt.max {
					t.Errorf("GetRandomInt(%d) = %d, want value in [0, %d)", tt.max, got1, tt.max)
				}
				if got2 < 0 || got2 >= tt.max {
					t.Errorf("GetRandomInt(%d) = %d, want value in [0, %d)", tt.max, got2, tt.max)
				}
			}
		})
	}
}

func TestGetExpDistributedInt(t *testing.T) {
	tests := []struct {
		name string
		rate float64
	}{
		{
			name: "rate-1",
			rate: 1,
		},
		{
			name: "rate-0.1",
			rate: 0.1,
		},
		{
			name: "rate-5",
			rate: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1 := GetExpDistributedInt(tt.rate)
			got2 := GetExpDistributedInt(tt.rate)

			// Exponential distribution should return non-negative integers.
			if got1 < 0 {
				t.Errorf("GetExpDistributedInt(%f) = %d, want non-negative value", tt.rate, got1)
			}
			if got2 < 0 {
				t.Errorf("GetExpDistributedInt(%f) = %d, want non-negative value", tt.rate, got2)
			}
		})
	}
}
