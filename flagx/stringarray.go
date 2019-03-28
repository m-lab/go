package flagx

import "fmt"

// StringArray is a new flag type. It appends the flag parameter to an
// `[]string` allowing the parameter to be specified multiple times.
type StringArray []string

// Get retrieves the value contained in the flag.
func (sa StringArray) Get() interface{} {
	return sa
}

// Set accepts a string parameter and appends it to the associated StringArray.
func (sa *StringArray) Set(s string) error {
	*sa = append(*sa, s)
	return nil
}

// String reports the StringArray as a Go value.
func (sa StringArray) String() string {
	return fmt.Sprintf("%#v", []string(sa))
}
