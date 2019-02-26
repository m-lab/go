// Package rtx provides free functions that would be handy to have as part of
// the go standard runtime.
package rtx

import (
	"io"
	"log"
)

// errorLoggingCloser provides a wrapper for implementations of io.Closer that
// logs any errors happening when calling the Close() method.
//
// Its main purpose is to turn long and ugly error-checking code such as:
//    defer func() {
//        err := resource.Close()
//        if err != nil {
//            log.Printf("Helpful message (error: %v)", err)
//        }
//    }()
// into a simplified pattern of:
//    defer ErrorLoggingCloser(resource).Close()
type errorLoggingCloser struct {
	c io.Closer
}

// Close() wraps io.Closer.Close() and logs errors.
func (elc errorLoggingCloser) Close() error {
	err := elc.c.Close()
	if err != nil {
		log.Println(err)
	}
	return err
}

// ErrorLoggingCloser wraps any io.Closer into an errorLoggingCloser.
func ErrorLoggingCloser(c io.Closer) io.Closer {
	return errorLoggingCloser{c}
}
