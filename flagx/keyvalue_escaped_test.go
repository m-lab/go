package flagx_test

import (
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/m-lab/go/flagx"
)

func TestKeyValueEscaped(t *testing.T) {
	d := t.TempDir()
	f := path.Join(d, "args.txt")
	tests := []struct {
		name    string
		kvs     string
		file    string
		want    map[string]string
		wantErr bool
	}{
		// KeyValue tests.
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
		// Escaping tests.
		{
			name: "success-single-value-escaped",
			kvs:  `a\/=b\,`,
			want: map[string]string{
				`a\/`: `b\,`,
			},
		},
		{
			name: "success-single-value-separtors-escaped",
			kvs:  `a=b\,c\=d`,
			want: map[string]string{
				`a`: `b\,c\=d`,
			},
		},
		{
			name: "success-multiple-value-escaped",
			kvs:  `a\,=b\,,c\,=d\,`,
			want: map[string]string{
				`a\,`: `b\,`,
				`c\,`: `d\,`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.file != "" {
				// write an args file to a tempdir.
				err := ioutil.WriteFile(f, []byte(tt.file), os.ModePerm)
				if err != nil {
					t.Fatalf("FlagsFromFile write file failed: %v", err)
				}
			}

			// Create kve flag.
			kve := &flagx.KeyValueEscaped{}
			if err := kve.Set(tt.kvs); (err != nil) != tt.wantErr {
				t.Errorf("KeyValueEscaped.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			// Get returns a map with expected values.
			got := kve.Get()
			if !mapsMatch(got, tt.want) {
				t.Errorf("KeyValueEscaped.Get() did not match; got = %v, want %v", got, tt.want)
			}

			// String returns the same fields given. Parse b/c order is not guaranteed.
			strFields := strings.Split(kve.String(), ",")
			sort.Strings(strFields)
			kvsFields := strings.Split(tt.kvs, ",")
			sort.Strings(kvsFields)
			if diff := deep.Equal(strFields, kvsFields); diff != nil {
				t.Errorf("KeyValueEscaped.String() did not match; got = %v, want %v, diff %v", strFields, kvsFields, diff)
			}
		})
	}
}
