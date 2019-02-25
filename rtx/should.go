// Package rtx provides free functions that would be handy to have as part of
// the go standard runtime.
package rtx

import (
	"fmt"
	"log"
)

// Allow overriding of this variable to aid in whitebox testing.
var logPrintln = log.Println

// Should will output a log message if passed a non-nil error. The message is
// specified as the prefix argument. If any further args are passed, then the
// prefix will be treated as a format string.
//
// The main purpose of this function is to turn an error-checking code like:
//    defer func() {
//        err := resource.Close()
//        if err != nil {
//            log.Printf("Helpful message (error: %v)", err)
//        }
//    }()
// into a simplified pattern of:
//    Should(resource.Close(), "Helpful message")
//
// This provides much more readable code and makes it easier to always check
// errors returned by deferred calls such as the resource.Close() above.
func Should(err error, prefix string, args ...interface{}) {
	if err != nil {
		suffix := fmt.Sprintf("(error: %v)", err)
		if len(args) != 0 {
			prefix = fmt.Sprintf(prefix, args...)
		}
		logPrintln(prefix, suffix)
	}
}
