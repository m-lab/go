// Package warnonfailure contains custom versions of common functions/methods
// that always log a warning in case of error.
package warnonfailure

import (
	"io"
	"log"
)

var logPrintf = log.Printf

// Close wraps a resource.Close() call and logs the error, if any.
//
// This function is intended to be used as a better alternative to completely
// ignoring an error in cases where there isn't any obvious reason to handle it
// explicitly.
//
// Its allows to turn ugly error-logging code such as:
//    defer func() {
//        err := resource.Close()
//        if err != nil {
//            log.Printf("Warning: ignoring error (%v)", err)
//        }
//    }()
// into a simplified pattern of:
//    defer Close(resource, "Warning: ignoring error")
func Close(c io.Closer, msg string) error {
	err := c.Close()
	if err != nil {
		logPrintf("%s (%s)\n", msg, err)
	}
	return err
}
