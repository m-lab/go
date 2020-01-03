package flagx

import (
	"fmt"
	"strings"
)

// StringArray is a new flag type. It appends the flag parameter to an
// `[]string` allowing the parameter to be specified multiple times or using ","
// separated items. Unlike other Flag types, the default argument should almost
// always be the empty array, because there is no way to remove an element, only
// to add one.
type StringArray []string

// Get retrieves the value contained in the flag.
func (sa StringArray) Get() interface{} {
	return sa
}

// Set accepts a string parameter and appends it to the associated StringArray.
// Set attempts to split the given string on commas "," and appends each element
// to the StringArray.
func (sa *StringArray) Set(s string) error {
	f := strings.Split(s, ",")
	*sa = append(*sa, f...)
	return nil
}

// String reports the StringArray as a Go value.
func (sa StringArray) String() string {
	return fmt.Sprintf("%#v", []string(sa))
}
