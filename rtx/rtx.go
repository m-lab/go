// Package rtx provides free functions that would be handy to have as part of
// the go standard runtime.
package rtx

import (
	"fmt"
	"log"
	"os"
)

// Allow overriding of this variable to aid in whitebox testing.
var osExit = os.Exit

// Must will call log.Fatal if passed a non-nil error. The fatal message is
// specified as the prefix argument. If any further args are passed, then the
// prefix will be treated as a format string.
//
// The main purpose of this function is to turn the common pattern of:
//
//	err := Func()
//	if err != nil {
//	    log.Fatalf("Helpful message (error: %v)", err)
//	}
//
// into a simplified pattern of:
//
//	Must(Func(), "Helpful message")
//
// This has the benefit of using fewer lines, a common error path that has test
// coverage, and enabling code which switches to this package to have 100%
// coverage.
func Must(err error, prefix string, args ...interface{}) {
	if err != nil {
		suffix := fmt.Sprintf(" (error: %v)", err)
		if len(args) != 0 {
			prefix = fmt.Sprintf(prefix, args...)
		}
		log.Output(2, prefix+suffix)
		osExit(1)
	}
}

// ValueOrDie is like Must, but it returns a value instead of having no output.
//
// It is expected this will be used like Must, but for functions that return a
// value, error tuple. Code like:
//
//	value, err := Func()
//	if err != nil {
//	    log.Fatalf(err)
//	}
//
// can be simplified to:
//
//	value := ValueOrDie(Func())
//
// When Must was originally added, Go didn't have generics. Now that it does,
// ValueOrDie is the generic version of Must. Unfortunately, because of how tuples
// are unpacked and passed to other functions, we can't supply a prefix or any
// other arguments.
func ValueOrDie[T any](value T, err error) T {
	if err != nil {
		log.Output(2, err.Error())
		osExit(1)
	}
	return value
}

// PanicOnError will call panic if passed a non-nil error. The message to panic
// is the prefix argument. If further args are passed, the prefix is treated as
// a format string. If you don't know whether to use `Must` or `PanicOnError`,
// you should use `Must`.
//
// This provides a function like Must which causes a recoverable failure, for
// use in things like web handlers, where the crashing and failure of a single
// response should not crash the server as a whole. The use of `panic()` and
// `recover()` in Go code is discouraged, and best practices dictate that the
// call to `panic()` and any corresponding `recover()` should be in the same
// package. PanicOnError is a simple function, so we adopt the rule that every
// call to PanicOnError and its corresponding `recover()` should be in the
// same package. For an example of this, see the code in the `scamper` package
// in https://github.com/m-lab/traceroute-caller
func PanicOnError(err error, prefix string, args ...interface{}) {
	if err != nil {
		suffix := fmt.Sprintf(" (error: %v)", err)
		if len(args) != 0 {
			prefix = fmt.Sprintf(prefix, args...)
		}
		log.Output(2, prefix+suffix)
		panic(prefix + suffix)
	}
}

// ValueOrPanic is like ValueOrDie, but it panics instead of exiting.
func ValueOrPanic[T any](value T, err error) T {
	if err != nil {
		log.Output(2, err.Error())
		panic(err)
	}
	return value
}
