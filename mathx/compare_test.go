package mathx

import "testing"

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "pos-pos",
			a:        2,
			b:        3,
			expected: 2,
		},
		{
			name:     "neg-neg",
			a:        -2,
			b:        -3,
			expected: -3,
		},
		{
			name:     "pos-neg",
			a:        2,
			b:        -3,
			expected: -3,
		},
		{
			name:     "same",
			a:        2,
			b:        2,
			expected: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Min(tt.a, tt.b)

			if got != tt.expected {
				t.Errorf("Min() = %d, want %d", got, tt.expected)
			}

			got = Min(tt.b, tt.a)

			if got != tt.expected {
				t.Errorf("Min() = %d, want %d", got, tt.expected)
			}
		})
	}
}
