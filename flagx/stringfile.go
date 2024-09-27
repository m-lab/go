package flagx

import (
	"fmt"
	"os"
)

// StringFile acts like the native flag.String by storing a string from the
// given argument. Additionally, StringFile may specify a file to read the string value from when
// prefixed with an '@', e.g. -flag=@value.txt
type StringFile struct {
	Value string
	file  string
}

// Set records the string in Value. When the first character of the parameter is
// prefixed with "@", i.e. "@file1", Set reads the file content for the value.
func (fs *StringFile) Set(v string) error {
	if len(v) > 0 && v[0] == '@' {
		fname := v[1:]
		b, err := os.ReadFile(fname)
		if err != nil {
			return err
		}
		*fs = StringFile{Value: string(b), file: fname}
	} else {
		*fs = StringFile{Value: v}
	}
	return nil
}

// String returns the flags in a form similiar to how they were added from the
// command line.
func (fs *StringFile) String() string {
	if fs.file != "" {
		return fmt.Sprintf("@%s", fs.file)
	} else {
		return fs.Value
	}
}

// Get returns the flag value.
func (fs *StringFile) Get() string {
	return fs.Value
}
