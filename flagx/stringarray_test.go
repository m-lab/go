package flagx_test

import (
	"flag"
	"testing"

	"github.com/go-test/deep"
	"github.com/m-lab/go/flagx"
)

func TestStringArray(t *testing.T) {
	tests := []struct {
		name string
		args []string
		expt flagx.StringArray
		repr string
	}{
		{
			name: "okay",
			args: []string{"a", "b"},
			expt: flagx.StringArray{"a", "b"},
			repr: `[]string{"a", "b"}`,
		},
		{
			name: "okay-split-commas",
			args: []string{"a", "b", "c,d"},
			expt: flagx.StringArray{"a", "b", "c", "d"},
			repr: `[]string{"a", "b", "c", "d"}`,
		},
		{
			name: "empty",
			args: []string{},
			expt: flagx.StringArray{},
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
			if diff := deep.Equal(v, tt.expt); diff != nil {
				t.Errorf("StringArray.Get() unexpected differences %v", diff)
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
