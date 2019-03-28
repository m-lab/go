package flagx_test

import (
	"flag"
	"testing"

	"github.com/m-lab/go/flagx"
)

func TestStringArray(t *testing.T) {
	tests := []struct {
		name string
		args []string
		repr string
	}{
		{
			name: "okay",
			args: []string{"a", "b"},
			repr: `[]string{"a", "b"}`,
		},
		{
			name: "empty",
			args: []string{},
			repr: `[]string{}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := &flagx.StringArray{}
			for i := range tt.args {
				if err := sa.Set(tt.args[i]); err != nil {
					t.Errorf("StringArray.Set() error = %v, want nil", err)
				}
			}
			v := (sa.Get().(flagx.StringArray))
			for i := range v {
				if v[i] != tt.args[i] {
					t.Errorf("StringArray.Get() want[%d] = %q, got[%d] %q",
						i, tt.args[i], i, v[i])
				}
			}
			if tt.repr != sa.String() {
				t.Errorf("StringArray.String() want = %q, got %q", tt.repr, sa.String())
			}
		})
	}
}

// Successful compilation of this function means that StringArray implements the
// flag.Getter interface. The function need not be called.
func assertFlagGetterStringArray(b flagx.StringArray) {
	func(in flag.Getter) {}(&b)
}
