// Package osx provides functionx which are extensions of the functionality provided in os.
package osx

import (
	"os"

	"github.com/m-lab/go/rtx"
)

// MustSetenv sets the environment variable named key to the passed-in value.
// Returns a function that, when run, restores the environment to its previous
// state.
func MustSetenv(key, value string) func() {
	oldVal, present := os.LookupEnv(key)
	rtx.Must(os.Setenv(key, value), "Could not set environment variable %q to %q", key, value)
	return func() {
		if present {
			rtx.Must(os.Setenv(key, oldVal), "Could not restore environment variable %q to %q", key, oldVal)

		} else {
			rtx.Must(os.Unsetenv(key), "Could not erase temporary env var %q", key)
		}
	}
}
