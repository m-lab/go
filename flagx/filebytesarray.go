package flagx

import (
	"io/ioutil"
	"strings"
)

// FileBytesArray is a new flag type that combines the semantics of StringArray
// for multiple filenames and FileBytes for reading the content of each file.
// Every filename is read as a `[]byte` and the contents appended to an
// FileBytesArray of type `[][]byte`.
//
// Like StringArray, the flag parameter may be specified multiple times or using
// "," separated items. Unlike other Flag types, the default argument should
// almost always be the empty array, because there is no way to remove an
// element, only to add one.
type FileBytesArray struct {
	// Bytes read from file names.
	bytes [][]byte
	// Names of files passed to Set. Preserved for meaningful String() output.
	names []string
}

// Get retrieves the bytes read from the file (or the default bytes).
func (fb *FileBytesArray) Get() [][]byte {
	return fb.bytes
}

// Set accepts a filename (or filenames separated by comma) and reads the bytes
// associated with the file and appends the bytes the FileBytesArray.
func (fb *FileBytesArray) Set(s string) error {
	f := strings.Split(s, ",")
	fb.names = append(fb.names, f...)
	for i := range f {
		b, err := ioutil.ReadFile(f[i])
		if err != nil {
			return err
		}
		fb.bytes = append(fb.bytes, b)
	}
	return nil
}

// String reports the original FileBytesArray filenames a string.
func (fb FileBytesArray) String() string {
	return strings.Join(fb.names, ",")
}
