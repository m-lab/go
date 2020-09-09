package testingx

import (
	"fmt"
)

// FatalReporter defines the interface for reporting a fatal test.
type FatalReporter interface {
	Fatal(args ...interface{})
	Helper()
}

// Must allows the rtx.Must pattern within a unit test and will call t.Fatal if
// passed a non-nil error. The fatal message is specified as the prefix
// argument. If any further args are passed, then the prefix will be treated as
// a format string.
//
// The main purpose of this function is to turn the common pattern of:
//    err := Func()
//    if err != nil {
//        t.Fatalf("Helpful message (error: %v)", err)
//    }
// into a simplified pattern of:
//    Must(t, Func(), "Helpful message")
//
// This has the benefit of using fewer lines and verifying unit tests are
// "correct by inspection".
func Must(t FatalReporter, err error, prefix string, args ...interface{}) {
	t.Helper() // Excludes this function from the line reported by t.Fatal.
	if err != nil {
		suffix := fmt.Sprintf(" (error: %v)", err)
		if len(args) != 0 {
			prefix = fmt.Sprintf(prefix, args...)
		}
		t.Fatal(prefix + suffix)
	}
}
