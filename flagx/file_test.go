package flagx_test

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/m-lab/go/rtx"

	"github.com/m-lab/go/flagx"
)

func TestFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "okay",
			content: "1234567890abcdef",
		},
		{
			name:    "error-bad-filename",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fname string
			var f *os.File
			var err error

			if !tt.wantErr {
				f, err = ioutil.TempFile("", "filebytes-*")
				rtx.Must(err, "Failed to create tempfile")
				defer os.Remove(f.Name())
				f.WriteString(tt.content)
				fname = f.Name()
				f.Close()
			} else {
				fname = "this-is-not-a-file"
			}

			fb := flagx.File{}
			if err := fb.Set(fname); (err != nil) != tt.wantErr {
				t.Errorf("File.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.content != string(fb.Get()) {
				t.Errorf("File.Get() want = %q, got %q", tt.content, string(fb.Get()))
			}
			if !tt.wantErr && tt.content != fb.Content() {
				t.Errorf("File.Get() want = %q, got %q", tt.content, fb.Content())
			}
			if !tt.wantErr && fname != fb.String() {
				t.Errorf("File.String() want = %q, got %q", fname, fb.String())
			}
		})
	}
}

// Successful compilation of this function means that File implements the
// flag.Value interface. The function need not be called.
func assertFlagValue(b flagx.File) {
	func(in flag.Value) {}(&b)
}
