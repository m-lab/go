package flagx

import (
	"fmt"
)

// Enum is a new flag type. An Enum contains an array of Options and a selected Value.
type Enum struct {
	Options []string
	Value   string
}

// Get retrieves the value contained in the flag.
func (e Enum) Get() string {
	return e.Value
}

// Set selects the Enum.Value if s equals one of the values in Enum.Options.
func (e *Enum) Set(s string) error {
	for i := range e.Options {
		if s == e.Options[i] {
			e.Value = s
			return nil
		}
	}
	return fmt.Errorf("%q is not a valid option: %+v", s, e.Options)
}

// String reports the set Enum value.
func (e Enum) String() string {
	return e.Value
}
