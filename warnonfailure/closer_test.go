package warnonfailure

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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

func callPrintf(f string, args ...interface{}) {
	fmt.Printf(f, args...)
}

func TestClose(t *testing.T) {
	logPrintf = callPrintf
	defer func() { logPrintf = log.Printf }()

	res := closeableResource{
		shouldErr: false,
	}

	resWithError := closeableResource{
		shouldErr: true,
	}

	// Technique from https://stackoverflow.com/questions/10473800/in-go-how-do-i-capture-stdout-of-a-function-into-a-string
	// Intercept stdout
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	outC := make(chan string)

	// Copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// Call the function which causes the output
	err := Close(resWithError, "Warning: ignored error")
	if err == nil {
		t.Errorf("Expected error, got nil.")
	}

	// Return back to normal state
	w.Close()
	os.Stdout = old

	// Find out what got read
	out := <-outC

	if string(out) != "Warning: ignored error (Error while closing resource)\n" {
		t.Errorf("%q != \"Warning: ignored error (Error while closing resource)\"", out)
	}

	err = Close(res, "Warning: ignored error")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

}
