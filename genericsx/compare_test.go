package genericsx

import "testing"

func TestContainsString(t *testing.T) {
	tests := []struct {
		name string
		s    []string
		e    string
		want bool
	}{
		{
			name: "contains",
			s:    []string{"apple", "banana", "orange"},
			e:    "banana",
			want: true,
		},
		{
			name: "not-contains",
			s:    []string{"apple", "banana", "orange"},
			e:    "mango",
			want: false,
		},
		{
			name: "empty",
			s:    []string{},
			e:    "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Contains(tt.s, tt.e)
			if got != tt.want {
				t.Errorf("Compare() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestContainsInt(t *testing.T) {
	tests := []struct {
		name string
		s    []int
		e    int
		want bool
	}{
		{
			name: "contains",
			s:    []int{1, 2},
			e:    1,
			want: true,
		},
		{
			name: "not-contains",
			s:    []int{1, 2},
			e:    0,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Contains(tt.s, tt.e)
			if got != tt.want {
				t.Errorf("Compare() got = %v, want = %v", got, tt.want)
			}
		})
	}
}
