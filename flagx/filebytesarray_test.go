package flagx_test

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/m-lab/go/rtx"

	"github.com/m-lab/go/flagx"
)

func TestFileBytesArray(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "success",
			content: "this is a test",
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
				f, err = ioutil.TempFile("", "-filebytes")
				rtx.Must(err, "Failed to create tempfile")
				defer os.Remove(f.Name())
				f.WriteString(tt.content)
				fname = f.Name()
				f.Close()
			} else {
				fname = "this-is-not-a-file"
			}

			fb := &flagx.FileBytesArray{}
			if err := fb.Set(fname); (err != nil) != tt.wantErr {
				t.Errorf("FileBytesArray.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			data := fb.Get()
			if !tt.wantErr && tt.content != string(data[0]) {
				t.Errorf("FileBytesArray.Get() want = %q, got %q", tt.content, string(data[0]))
			}
			if !tt.wantErr && fname != fb.String() {
				t.Errorf("FileBytesArray.String() want = %q, got %q", tt.content, fb.String())
			}
		})
	}
}

// Successful compilation of this function means that FileBytes implements the
// flag.Getter interface. The function need not be called.
func assertFlagGetterFileBytesArray(b flagx.FileBytesArray) {
	func(in flag.Value) {}(&b)
}
