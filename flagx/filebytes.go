package flagx

import (
	"encoding/hex"
	"io/ioutil"
)

// FileBytes is a new flag type. It automatically reads the content of the
// given filename as a `[]byte`, handling errors during flag parsing.
type FileBytes []byte

// Get retrieves the value contained in the flag.
func (fb FileBytes) Get() interface{} {
	return fb
}

// Set accepts a filename and reads the bytes associated with that file.
func (fb *FileBytes) Set(s string) error {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		return err
	}
	*fb = b
	return nil
}

// String reports the FileBytes content as a hexdump.
func (fb FileBytes) String() string {
	return hex.Dump(fb)
}
