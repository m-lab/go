package flagx_test

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/m-lab/go/flagx"
)

func TestKeyArrayValue(t *testing.T) {
	tests := []struct {
		name       string
		flags      []string
		want       map[string][]string
		wantErr    bool
		wantString string
	}{
		{
			name:  "success-single-pair-one-value",
			flags: []string{"key1=value1"},
			want: map[string][]string{
				"key1": []string{"value1"},
			},
			wantString: "key1:[value1]",
		},
		{
			name:  "success-single-pair-multiple-values",
			flags: []string{"key1=value1,value1.1"},
			want: map[string][]string{
				"key1": []string{"value1", "value1.1"},
			},
			wantString: "key1:[value1,value1.1]",
		},
	}

	for _, tt := range tests {
		kav := flagx.KeyArrayValue{}
		for _, f := range tt.flags {
			if err := kav.Set(f); (err != nil) != tt.wantErr {
				t.Errorf("KeyArrayValue.Set() error: %v, wantErr: %v", err, tt.wantErr)
			}
		}
		if tt.wantErr {
			return
		}

		got := kav.Get()
		if diff := deep.Equal(got, tt.want); diff != nil {
			t.Errorf("KeyArrayValue.Get() did not match; got: %v, want: %v", got, tt.want)
		}

		if kav.String() != tt.wantString {
			t.Errorf("KeyArrayValue.String() did not match; got: %s, want: %s",
				kav.String(), tt.wantString)
		}
	}
}
