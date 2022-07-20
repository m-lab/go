package mathx

import (
	"math"
	"testing"
)

func TestGetHaversineDistance(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64
	}{
		{
			name:     "US-UK",
			lat1:     37.09024,
			lon1:     -95.712891,
			lat2:     55.378051,
			lon2:     -3.435973,
			expected: 6830.40,
		},
		{
			name:     "US-US",
			lat1:     37.09024,
			lon1:     -95.712891,
			lat2:     37.09024,
			lon2:     -95.712891,
			expected: 0,
		},
		{
			name:     "0-UK",
			lat1:     0,
			lon1:     0,
			lat2:     55.378051,
			lon2:     -3.435973,
			expected: 6165.67,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetHaversineDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)

			if !compareFloats(got, tt.expected) {
				t.Errorf("GetHaversineDistance() = %f, want %f", got, tt.expected)
			}

			got = GetHaversineDistance(tt.lat2, tt.lon2, tt.lat1, tt.lon1)

			if !compareFloats(got, tt.expected) {
				t.Errorf("GetHaversineDistance() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func compareFloats(f1 float64, f2 float64) bool {
	diff := math.Abs(f1 - f2)
	return diff < 0.01
}
