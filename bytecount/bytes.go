// Package bytecount provides a single datatype ByteCount, designed to track
// counts of bytes, as well as some helper constants.  It also provides all the
// necessary functions to allow a ByteCount to be specified as a command-line
// argument, which should allow command-line arguments like
// `--cache-size=20MB`.
package bytecount

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	r "github.com/m-lab/go/runtimeext"
)

// ByteCount holds filesizes and the like.
type ByteCount int64

// Some constants to make working with ByteCounts easier.  If the difference
// between Kilobytes and Kibibytes matters to you, then this library is likely
// too unsophisticated for your needs.  Pull requests are welcome, however.
const (
	Byte     ByteCount = 1
	Kilobyte           = 1000 * Byte
	Megabyte           = 1000 * Kilobyte
	Gigabyte           = 1000 * Megabyte
)

// Get is used by the Flag library to get the value out of the ByteCount.
func (b ByteCount) Get() interface{} {
	return b
}

// String is used by the Flag library to turn the value into a string.  Because
// we output bytecounts in our logs, it is worth it to convert them into
// readable values.
func (b ByteCount) String() string {
	if b%Gigabyte == 0 {
		return fmt.Sprintf("%dGB", b/Gigabyte)
	} else if b%Megabyte == 0 {
		return fmt.Sprintf("%dMB", b/Megabyte)
	} else if b%Kilobyte == 0 {
		return fmt.Sprintf("%dKB", b/Kilobyte)
	}
	return fmt.Sprintf("%dB", b)
}

// Set is used by the Flag library to turn a string into a ByteCount.  This
// implementation parses on the quick and dirty using regular expressions.
func (b *ByteCount) Set(s string) error {
	bytesRegexpStr := `^(?P<quantity>[0-9]+)(?P<units>[KMG]?B?)?$`
	bytesRegexp := regexp.MustCompile(bytesRegexpStr)
	if !bytesRegexp.MatchString(s) {
		return fmt.Errorf("Invalid size format: %q", s)
	}
	for _, submatches := range bytesRegexp.FindAllStringSubmatchIndex(s, -1) {
		quantityBytes := bytesRegexp.ExpandString([]byte{}, "$quantity", s, submatches)
		quantityInt, err := strconv.ParseInt(string(quantityBytes), 10, 64)
		// If this check ever fails, it represents a bug in the code rather than a
		// normal response to bad input. A richer compiler would be able to prove
		// that this check always passes. Regrettably, that compiler does not exist.
		r.Must(err, "The string %q passed the regexp %q but did not have an int we could parse. This is a bug.", s, bytesRegexpStr)
		quantity := ByteCount(quantityInt)
		unitsBytes := bytesRegexp.ExpandString([]byte{}, "$units", s, submatches)
		units := Byte
		err = errors.New("No units found")
		switch string(unitsBytes) {
		case "B", "":
			units = Byte
			err = nil
		case "KB", "K":
			units = Kilobyte
			err = nil
		case "MB", "M":
			units = Megabyte
			err = nil
		case "GB", "G":
			units = Gigabyte
			err = nil
		}
		// If this check ever fails, it represents a bug in the code rather than a
		// normal response to bad input. A richer compiler would be able to prove
		// that this check always passes. Regrettably, that compiler does not exist.
		r.Must(err, "The string %q passed the regexp %q but did not have units we could parse. This is a bug.", s, bytesRegexpStr)
		*b = quantity * units
	}
	return nil
}
