package flagx

import (
	"fmt"
	"strings"
	"time"
)

// DurationArray collects time.Durations so a duration flag can be specified
// multiple times or using "," separated items. The default argument should
// almost always be the empty array, because there is no way to remove an
// element, only to add one.
type DurationArray []time.Duration

// Get retrieves the value contained in the flag.
func (da DurationArray) Get() interface{} {
	return da
}

// Set appends the given time.Duration to the DurationArray. Set accepts
// multiple durations separated by commas "," and appends each element to the
// DurationArray.
func (da *DurationArray) Set(s string) error {
	f := strings.Split(s, ",")
	for _, d := range f {
		dur, err := time.ParseDuration(d)
		if err != nil {
			return err
		}
		*da = append(*da, dur)
	}
	return nil
}

// String reports the DurationArray as a Go value.
func (da DurationArray) String() string {
	return fmt.Sprintf("%+v", []time.Duration(da))
}
