package flagx_test

import (
	"os"
	"path"
	"testing"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/testingx"
)

func TestStringFile(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		useFile bool
		wantErr bool
	}{
		{
			name:    "success-string",
			value:   "value12345",
			useFile: false,
		},
		{
			name:    "success-file",
			value:   "1234567890abcdef",
			useFile: true,
		},
		{
			name:    "error-file",
			value:   "@error-bad-filename",
			useFile: false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			value := tt.value

			if !tt.wantErr && tt.useFile {
				// This is a file read - so create a file in a temp directory.
				dir := t.TempDir()
				name := path.Join(dir, "file.txt")
				testingx.Must(t, os.WriteFile(name, []byte(tt.value), 0664), "failed to write test file")
				defer os.Remove(name)
				value = "@" + name // reset name to include directory.
			}

			fb := &flagx.StringFile{}
			if err := fb.Set(value); (err != nil) != tt.wantErr {
				t.Errorf("StringFile.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if tt.value != fb.Get() {
				t.Errorf("StringFile.Get() want = %q, got %q", tt.value, fb.Get())
			}
			if fb.String()[0] != '@' && tt.useFile {
				t.Errorf("StringFile.String() want = @<file>, got %q", fb.String())
			}
		})
	}
}
