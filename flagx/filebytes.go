package flagx

import (
	"io/ioutil"
)

// FileBytes is a new flag type. It automatically reads the content of the
// given filename as a `[]byte`, handling errors during flag parsing.
type FileBytes []byte

// Get retrieves the bytes read from the file (or the default bytes).
func (fb FileBytes) Get() interface{} {
	return fb
}

// Set accepts a filename and reads the bytes associated with that file into the
// FileBytes storage.
func (fb *FileBytes) Set(s string) error {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		return err
	}
	*fb = b
	return nil
}

// String reports the FileBytes content as a string.
//
// FileBytes are awkward to represent in help text, and such help text is the
// main use of the Stringer interface for this flag. Help text like:
//   "Sets the file containing the prefix string. The default file contents are: " + fb.String()
// is recommended.
func (fb FileBytes) String() string {
	return string([]byte(fb))
}
