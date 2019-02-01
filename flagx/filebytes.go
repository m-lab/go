package flagx

import (
	"encoding/hex"
	"io/ioutil"
)

// FileBytes holds the file bytes.
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

// String reports the FileBytes as a raw string. Unprintable characters will
// still be unprintable.
func (fb FileBytes) String() string {
	return hex.Dump(fb)
}
