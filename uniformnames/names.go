// Package uniformnames allows experiments and utilities to check whether names
// conform the requirements for experiment and datatype names in M-Lab.
package uniformnames

import (
	"fmt"
	"regexp"
)

var (
	nameRE = regexp.MustCompile("^[a-z][a-z0-9]*$")
)

// Check whether a passed-in name conforms to the specs set out in the "M-Lab
// Unform Naming" doc.
func Check(name string) error {
	if nameRE.MatchString(name) {
		return nil
	}
	return fmt.Errorf("%q did not match the regexp %q", name, nameRE.String())
}
