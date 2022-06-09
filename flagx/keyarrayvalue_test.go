package flagx_test

import (
	"reflect"
	"testing"

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
			wantErr:    false,
			wantString: "key1:[value1]",
		},
		{
			name:  "success-single-pair-multiple-values",
			flags: []string{"key1=value1,value1.1"},
			want: map[string][]string{
				"key1": []string{"value1", "value1.1"},
			},
			wantErr:    false,
			wantString: "key1:[value1,value1.1]",
		},
		{
			name: "success-multiple-pairs",
			flags: []string{
				"key1=value1,value1.1",
				"key2=value2,value2.1,value2.2",
				"key3=value3",
			},
			want: map[string][]string{
				"key1": []string{"value1", "value1.1"},
				"key2": []string{"value2", "value2.1", "value2.2"},
				"key3": []string{"value3"},
			},
			wantErr:    false,
			wantString: "key1:[value1,value1.1],key2:[value2,value2.1,value2.2],key3:[value3]",
		},
		{
			name:       "invalid-input",
			flags:      []string{"key1=value1,key2=value2"},
			want:       map[string][]string{},
			wantErr:    true,
			wantString: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyArrayValue.Get() did not match; got: %v, want: %v", got, tt.want)
			}

			if kav.String() != tt.wantString {
				t.Errorf("KeyArrayValue.String() did not match; got: %s, want: %s",
					kav.String(), tt.wantString)
			}
		})
	}
}
