package iox

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

	err := ErrorLoggingCloser(res).Close()

	if err != nil {
		t.Errorf("Unexpected error from Close(): %v\n", err)
	}

	err = ErrorLoggingCloser(resWithError).Close()
	if err == nil {
		t.Errorf("Expected error, got nil.")
	}
}
