package flagx_test

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/m-lab/go/flagx"
)

func TestKeyValueArray(t *testing.T) {
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
				"key2=value2.3",
			},
			want: map[string][]string{
				"key1": []string{"value1", "value1.1"},
				"key2": []string{"value2", "value2.1", "value2.2", "value2.3"},
				"key3": []string{"value3"},
			},
			wantErr:    false,
			wantString: "key1:[value1,value1.1],key2:[value2,value2.1,value2.2,value2.3],key3:[value3]",
		},
		{
			name:       "invalid-input",
			flags:      []string{"key1"},
			want:       map[string][]string{},
			wantErr:    true,
			wantString: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kva := flagx.KeyValueArray{}
			for _, f := range tt.flags {
				if err := kva.Set(f); (err != nil) != tt.wantErr {
					t.Errorf("KeyValueArrray.Set() error: %v, wantErr: %v", err, tt.wantErr)
				}
			}
			if tt.wantErr {
				return
			}

			got := kva.Get()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyValueArray.Get() did not match; got: %v, want: %v", got, tt.want)
			}

			// Sort because order is not guaranteed.
			strFields := strings.Split(kva.String(), ",")
			sort.Strings(strFields)
			kvsFields := strings.Split(tt.wantString, ",")
			sort.Strings(kvsFields)
			if diff := deep.Equal(strFields, kvsFields); diff != nil {
				t.Errorf("KeyValueArray.String() did not match; got = %v, want %v, diff %v", strFields, kvsFields, diff)
			}
		})
	}
}
