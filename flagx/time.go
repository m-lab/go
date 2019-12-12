package flagx

import (
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	"src/github.com/m-lab/go/rtx"
)

// ErrBadTimeFormat is returned when failing to parse a Time value.
var ErrBadTimeFormat = fmt.Errorf("ErrBadTimeFormat: unsupported time format")

// Local prototype time formats.
var timeFormat = "15:04:05"

// DateTime is a flag type for accepting date parameters.
type DateTime struct {
	time.Time
	Raw string // The original string parameter.
}

// Get retrieves the current date value contained in the flag as a string.
func (t DateTime) Get() string {
	return t.Raw
}

// Set parses and assigns the DateTime value. DateTime accepts all formats
// supported by the "github.com/araddon/dateparse" package.
func (t *DateTime) Set(s string) error {
	(*t).Raw = s
	_, err := dateparse.ParseStrict(s)
	if err != nil {
		return err
	}
	f, err := dateparse.ParseIn(s, time.UTC)
	// If ParseStrict succeeds, then ParseIn is always expected to succeed.
	rtx.Must(err, "Failed to infer format from %q", s)
	(*t).Time = f
	return nil
}

// String reports the Raw value used to derive the DateTime value. The raw value
// is provided because not all formats supported by parsedate are supported by
// time.Format (e.g. unix timestamps). The DateTime type relies on
// dateparse.ParseStrict errors to enforce equivalence between the .Raw and
// .Time values.
func (t DateTime) String() string {
	return t.Get()
}

// Time is a flag type for accepting time parameters formatted as HH:MM:SS. If
// you need sub-second resolution, consider using one of the unix timestamp
// formats (ms, usec, or ns) supported by DateTime.
type Time struct {
	Hour   int
	Minute int
	Second int
}

// Get retrieves the value contained in the flag.
func (t Time) Get() string {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second)
}

// Set parses and assigns the Time Hour, Minute, and Second values.
func (t *Time) Set(s string) error {
	var format string
	switch {
	case len(s) == len(timeFormat):
		format = timeFormat
	default:
		return ErrBadTimeFormat
	}
	tmp, err := time.Parse(format, s)
	if err != nil {
		return err
	}
	(*t).Hour = tmp.Hour()
	(*t).Minute = tmp.Minute()
	(*t).Second = tmp.Second()
	return nil
}

// String reports the set Time value.
func (t Time) String() string {
	return t.Get()
}
