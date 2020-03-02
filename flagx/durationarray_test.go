package flagx_test

import (
	"flag"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/m-lab/go/flagx"
)

func TestDurationArray(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		expt    flagx.DurationArray
		repr    string
		wantErr bool
	}{
		{
			name: "okay",
			args: []string{"1s", "1m", "1h"},
			expt: flagx.DurationArray{time.Second, time.Minute, time.Hour},
			repr: `[1s 1m0s 1h0m0s]`,
		},
		{
			name: "okay-split-commas",
			args: []string{"1s,1m,1h"},
			expt: flagx.DurationArray{time.Second, time.Minute, time.Hour},
			repr: `[1s 1m0s 1h0m0s]`,
		},
		{
			name:    "empty",
			args:    []string{},
			expt:    flagx.DurationArray{},
			repr:    `[]`,
			wantErr: true,
		},
		{
			name:    "error-bad-format",
			args:    []string{"this-is-not-a-duration"},
			expt:    flagx.DurationArray{},
			repr:    `[]`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			da := &flagx.DurationArray{}
			for i := range tt.args {
				err := da.Set(tt.args[i])
				if (err != nil) != tt.wantErr {
					t.Errorf("DurationArray.Set() error = %v, want nil", err)
				}
			}
			v := (da.Get().(flagx.DurationArray))
			if diff := deep.Equal(v, tt.expt); diff != nil {
				t.Errorf("DurationArray.Get() unexpected differences %v", diff)
			}
			if tt.repr != da.String() {
				t.Errorf("DurationArray.String() want = %q, got %q", tt.repr, da.String())
			}
		})
	}
}

// Successful compilation of this function means that DurationArray implements the
// flag.Value interface. The function need not be called.
func assertFlagDurationArray(b flagx.DurationArray) {
	func(in flag.Value) {}(&b)
}
