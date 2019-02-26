// Package rtx provides free functions that would be handy to have as part of
// the go standard runtime.
package rtx

import (
	"errors"
	"testing"
)

type closeableResource struct {
	shouldErr bool
}

func (c closeableResource) Close() error {
	if c.shouldErr {
		return errors.New("Error while closing resource")
	}
	return nil
}

func TestErrorLoggingCloser(t *testing.T) {
	res := closeableResource{
		shouldErr: false,
	}

	resWithError := closeableResource{
		shouldErr: true,
	}

	// Call the function which causes the output
	err := ErrorLoggingCloser(res).Close()

	if err != nil {
		t.Errorf("Unexpected error from Close(): %v\n", err)
	}

	err = ErrorLoggingCloser(resWithError).Close()
	if err == nil {
		t.Errorf("Expected error, got nil.")
	}
}
