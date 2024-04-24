package mathx

import (
	"testing"
)

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

func TestMode(t *testing.T) {
	tests := []struct {
		name    string
		slice   []int64
		want    int64
		wantErr bool
	}{
		{
			name:    "empty",
			slice:   []int64{},
			want:    0,
			wantErr: true,
		},
		{
			name:    "single",
			slice:   []int64{1, 2, 1, 3},
			want:    1,
			wantErr: false,
		},
		{
			name:    "multiple",
			slice:   []int64{1, 2, 3, 1, 2},
			want:    1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Mode(tt.slice)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Mode() = %v, want %v", got, tt.want)
			}
		})
	}
}
