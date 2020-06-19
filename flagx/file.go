package flagx

import (
	"io/ioutil"
)

// File is a new flag type. For a given filename, File reads and saves the file
// content to Bytes and the original filename to Name. Errors opening or
// reading the file are handled during flag parsing.
type File struct {
	Bytes []byte
	Name  string
}

// Get retrieves the bytes read from the file.
func (fb *File) Get() []byte {
	return fb.Bytes
}

// Content retrieves the bytes read from the file as a string.
func (fb *File) Content() string {
	return string(fb.Bytes)
}

// Set accepts a file name. On success, the file content is saved to Bytes, and
// the original file name to Name.
func (fb *File) Set(s string) error {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		return err
	}
	fb.Name = s
	fb.Bytes = b
	return nil
}

// String reports the original file Name. NOTE: String is typically used by the
// flag help text and to report flag values. To return the file content as a
// string, see File.Content().
func (fb *File) String() string {
	return fb.Name
}
