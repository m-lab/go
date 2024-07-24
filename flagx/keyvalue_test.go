package flagx_test

import (
	"flag"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/go-test/deep"

	"github.com/m-lab/go/flagx"
)

func mapsMatch(src, dst map[string]string) bool {
	for key, val := range src {
		if dst[key] != val {
			return false
		}
	}
	for key, val := range dst {
		if src[key] != val {
			return false
		}
	}
	return true
}

func TestKeyValue(t *testing.T) {
	d := t.TempDir()
	f := path.Join(d, "args.txt")
	tests := []struct {
		name        string
		kvs         string
		file        string
		want        map[string]string
		ignoreError bool
		wantErr     bool
	}{
		{
			name: "success-single-keyvalue",
			kvs:  "a=b",
			want: map[string]string{
				"a": "b",
			},
		},
		{
			name: "success-multiple-keyvalue",
			kvs:  "a=b,c=d",
			want: map[string]string{
				"a": "b",
				"c": "d",
			},
		},
		{
			name: "success-single-from-file",
			kvs:  "c=@" + f,
			file: "d",
			want: map[string]string{
				"c": "d",
			},
		},
		{
			name: "success-multiple-from-file",
			kvs:  "a=b,c=@" + f,
			file: "d",
			want: map[string]string{
				"a": "b",
				"c": "d",
			},
		},
		{
			name: "success-multiple-from-file-2",
			kvs:  "a=@" + f + ",c=@" + f,
			file: "d",
			want: map[string]string{
				"a": "d",
				"c": "d",
			},
		},
		{
			name: "success-strip-whitespace-newlines",
			kvs:  "a=@" + f,
			file: "\t d    \n\n ",
			want: map[string]string{
				"a": "d",
			},
		},
		{
			name:        "success-missing-file",
			kvs:         "a=@this-file-does-not-exist",
			ignoreError: true,
			want: map[string]string{
				"a": "",
			},
		},
		{
			name:    "error-bad-key-value",
			kvs:     "a=b,c",
			wantErr: true,
		},
		{
			name:    "error-missing-file",
			kvs:     "a=b,c=@",
			wantErr: true,
		},
		{
			name:    "error-not-a-file",
			kvs:     "a=b,c=@/",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.file != "" {
				// write an args file to a tempdir.
				err := os.WriteFile(f, []byte(tt.file), os.ModePerm)
				if err != nil {
					t.Fatalf("FlagsFromFile write file failed: %v", err)
				}
			}

			// Create kv flag.
			kv := &flagx.KeyValue{IgnoreFileError: tt.ignoreError}
			if err := kv.Set(tt.kvs); (err != nil) != tt.wantErr {
				t.Errorf("KeyValue.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			// Get returns a map with expected values.
			got := kv.Get()
			if !mapsMatch(got, tt.want) {
				t.Errorf("KeyValue.Get() did not match; got = %v, want %v", got, tt.want)
			}

			// String returns the same fields given. Parse b/c order is not guaranteed.
			strFields := strings.Split(kv.String(), ",")
			sort.Strings(strFields)
			kvsFields := strings.Split(tt.kvs, ",")
			sort.Strings(kvsFields)
			if diff := deep.Equal(strFields, kvsFields); diff != nil {
				t.Errorf("KeyValue.String() did not match; got = %v, want %v, diff %v", strFields, kvsFields, diff)
			}
		})
	}
}

// If this compiles successfully then flagx.KeyValue conforms to the flag.Value
// interface.
func AssertKeyValueIsFlagValue(kv *flagx.KeyValue) {
	func(f flag.Value) {}(kv)
}

func Example() {
	metadata := flagx.KeyValue{}
	flag.Var(&metadata, "metadata", "Key-value pairs to be added to the metadata (flag may be repeated)")
	// Commandline flags should look like: -metadata key1=val1 -metadata key2=val2
	flag.Parse()

	log.Println(metadata.Get())
}
