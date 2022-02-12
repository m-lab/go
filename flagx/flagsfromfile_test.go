package flagx_test

import (
	"flag"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/m-lab/go/flagx"
)

type flagVals struct {
	i int
	s string
	l flagx.KeyValue
}

func (f *flagVals) Match(v *flagVals) bool {
	if f.i != v.i {
		return false
	}
	if f.s != v.s {
		return false
	}
	fm := f.l.Get()
	vm := v.l.Get()
	for key, val := range fm {
		if vm[key] != val {
			return false
		}
	}
	for key, val := range vm {
		if fm[key] != val {
			return false
		}
	}
	return true
}

func newKeyValue(m map[string]string) flagx.KeyValue {
	kv := flagx.KeyValue{}
	for k, v := range m {
		kv.Set(k + "=" + v)
	}
	return kv
}

func TestFlagsFromFile(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		wantVals *flagVals
		rmFile   bool
		wantErr  bool
	}{
		{
			name: "success-one-line",
			file: "-s test -i 10 -l a=b",
			wantVals: &flagVals{
				i: 10,
				s: "test",
				l: newKeyValue(map[string]string{"a": "b"}),
			},
		},
		{
			name: "success-multi-flag-multi-line",
			file: "-s first -s second\n-i 30\n-i 10\n-l a=b\n-l c=d",
			wantVals: &flagVals{
				i: 10,       // last value.
				s: "second", // last value.
				l: newKeyValue(map[string]string{"a": "b", "c": "d"}),
			},
		},
		{
			name:    "error-no-file",
			rmFile:  true,
			wantErr: true,
		},
		{
			name:    "error-parse-undefined-flag",
			file:    "-undefined=string",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a flagset with flags.
			v := flagVals{}
			fs := flag.NewFlagSet("", flag.ContinueOnError)
			fs.StringVar(&v.s, "s", "", "")
			fs.IntVar(&v.i, "i", 0, "")
			fs.Var(&v.l, "l", "")

			// write an args file to a tempdir.
			d := t.TempDir()
			f := path.Join(d, "args.txt")
			err := ioutil.WriteFile(f, []byte(tt.file), os.ModePerm)
			if err != nil {
				t.Fatalf("FlagsFromFile write file failed: %v", err)
			}
			if tt.rmFile {
				os.Remove(f)
			}
			// set the flag by passing the file name.
			ff := &flagx.FlagsFromFile{CommandLine: fs}
			if err := ff.Set(f); (err != nil) != tt.wantErr {
				t.Fatalf("FlagsFromFile.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			// verify the expected values match the parsed values.
			if !v.Match(tt.wantVals) {
				t.Fatalf("FlagsFromFile values do not match: got %#v, want %#v", v, tt.wantVals)
			}
			if ff.Get() != f {
				t.Fatalf("FlagsFromFile.Get() wrong name; got %q, want %q", ff.Get(), f)
			}
			if ff.String() != f {
				t.Fatalf("FlagsFromFile.String() wrong name; got %q, want %q", ff.Get(), f)
			}
		})
	}
}
