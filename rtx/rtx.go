// Package rtx provides free functions that would be handy to have as part of
// the go standard runtime.
package rtx

import (
	"fmt"
	"log"
)

// Allow overriding of this variable to aid in whitebox testing.
var logFatal = log.Fatal

// Must will call log.Fatal if passed a non-nil error. The fatal message is
// specified as the prefix argument. If any further args are passed, then the
// prefix will be treated as a format string.
//
// The main purpose of this function is to turn the common pattern of:
//    err := Func()
//    if err != nil {
//        log.Fatalf("Helpful message (error: %v)", err)
//    }
// into a simplified pattern of:
//    Must(Func(), "Helpful message")
//
// This has the benefit of using fewer lines, a common error path that has test
// coverage, and enabling code which switches to this package to have 100%
// coverage.
func Must(err error, prefix string, args ...interface{}) {
	if err != nil {
		suffix := fmt.Sprintf("(error: %v)", err)
		if len(args) != 0 {
			prefix = fmt.Sprintf(prefix, args...)
		}
		logFatal(prefix, suffix)
	}
}

// Should will call panic if passed a non-nil error. The message to panic is the
// prefix argument. If further args are passed, the prefix is treated as a
// format string.
//
// This provides a version of Must which has a recoverable failure, for use in
// things like web handlers, where the crashing and failure of a single response
// should not crash the server as a whole.
func Should(err error, prefix string, args ...interface{}) {
	if err != nil {
		suffix := fmt.Sprintf(" (error: %v)", err)
		if len(args) != 0 {
			prefix = fmt.Sprintf(prefix, args...)
		}
		panic(prefix + suffix)
	}
}
