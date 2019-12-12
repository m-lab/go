package flagx

import (
	"fmt"
	"strings"
	"time"
)

// ErrBadFormat is a static error returned when attempting to parse an unsupported flag format.
var ErrBadFormat = fmt.Errorf("ErrBadFormat: unsupported time format")

// Local prototype time formats.
var timeFormat = "15:04:05"
var dateFormat = "2006-01-02"
var dateTimeFormat = "2006-01-02T15:04:05"

// DateTime is a new flag type.
type DateTime struct {
	Date
	Time
}

// Get retrieves the value contained in the flag.
func (t DateTime) Get() string {
	return t.Date.String() + "T" + t.Time.String()
}

// Set parses and assigns the DateTime value. As a convenience, DateTime accepts
// abbreviated formats for date. For example: 2019-01-30 or 2019-01-30T01:01:34
func (t *DateTime) Set(s string) error {
	switch {
	case len(s) == len(dateFormat):
		return (*t).Date.Set(s)
	case len(s) == len(dateTimeFormat):
		f := strings.Split(s, "T")
		if len(f) != 2 {
			return ErrBadFormat
		}
		err := (*t).Date.Set(f[0])
		if err != nil {
			return err
		}
		return (*t).Time.Set(f[1])
	default:
		return ErrBadFormat
	}
}

// String reports the set Time value.
func (t DateTime) String() string {
	return t.Get()
}

// Date is a new flag type.
type Date struct {
	Year  int
	Month int
	Day   int
}

// Get retrieves the value contained in the flag.
func (t Date) Get() string {
	return fmt.Sprintf("%04d-%02d-%02d", t.Year, t.Month, t.Day)
}

// Set parses and assigns the Date Year, Month, and Day values.
func (t *Date) Set(s string) error {
	var format string
	switch {
	case len(s) == len(dateFormat):
		format = dateFormat
	default:
		return ErrBadFormat
	}
	tmp, err := time.Parse(format, s)
	if err != nil {
		return err
	}
	(*t).Year = tmp.Year()
	(*t).Month = int(tmp.Month())
	(*t).Day = tmp.Day()
	return nil
}

// String reports the set Time value.
func (t Date) String() string {
	return t.Get()
}

// Time is a new flag type.
type Time struct {
	Hour   int
	Minute int
	Second int
}

// Get retrieves the value contained in the flag.
func (t Time) Get() string {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second)
}

// Set parses and assigns the Date Year, Month, and Day values.
func (t *Time) Set(s string) error {
	var format string
	switch {
	case len(s) == len(timeFormat):
		format = timeFormat
	default:
		return ErrBadFormat
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
