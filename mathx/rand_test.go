package mathx

import (
	"math/rand"
	"testing"
)

const seed = 1658340109320624211

func TestGetRandomInt(t *testing.T) {
	tests := []struct {
		name      string
		max       int
		expected1 int
		expected2 int
	}{
		{
			name:      "random",
			max:       10,
			expected1: 6,
			expected2: 8,
		},
		{
			name:      "zero",
			max:       0,
			expected1: 0,
			expected2: 0,
		},
		{
			name:      "negative",
			max:       -10,
			expected1: 0,
			expected2: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rand.Seed(seed)
			got := GetRandomInt(tt.max)

			if got != tt.expected1 {
				t.Errorf("GetRandomInt() = %d, want %d", got, tt.expected1)
			}

			got = GetRandomInt(tt.max)

			if got != tt.expected2 {
				t.Errorf("GetRandomInt() = %d, want %d", got, tt.expected2)
			}
		})
	}
}

func TestGetExpDistributedInt(t *testing.T) {
	tests := []struct {
		name      string
		rate      float64
		expected1 int
		expected2 int
	}{
		{
			name:      "rate-1",
			rate:      1,
			expected1: 1,
			expected2: 0,
		},
		{
			name:      "rate-0.1",
			rate:      0.1,
			expected1: 5,
			expected2: 2,
		},
		{
			name:      "rate-5",
			rate:      5,
			expected1: 0,
			expected2: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rand.Seed(seed)
			got := GetExpDistributedInt(tt.rate)

			if got != tt.expected1 {
				t.Errorf("GetExpDistributedInt() = %d, want %d", got, tt.expected1)
			}

			got = GetExpDistributedInt(tt.rate)

			if got != tt.expected2 {
				t.Errorf("GetExpDistributedInt() = %d, want %d", got, tt.expected2)
			}
		})
	}
}
