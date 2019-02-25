// Package rtx provides free functions that would be handy to have as part of
// the go standard runtime.
package rtx

import (
	"bytes"
	"io"
	"log"
	"os"
	"testing"
)

func TestShouldSuccess(t *testing.T) {
	Should(success(), "It works!")
}

func TestShouldFail(t *testing.T) {
	logPrintln = callPrintln
	defer func() { logPrintln = log.Println }()

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

	Should(fail(), "Should fail")

	// Return back to normal state
	w.Close()
	os.Stdout = old

	// Find out what got read
	out := <-outC

	if string(out) != "Should fail (error: A failure for testing)\n" {
		t.Errorf("%q != \"Should fail (error: An error for testing)\"", out)
	}
}

func TestShouldFailWithFormatting(t *testing.T) {
	logPrintln = callPrintln
	defer func() { logPrintln = log.Println }()

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

	Should(fail(), "Should fail with arg %d", 5)

	// Return back to normal state
	w.Close()
	os.Stdout = old

	// Find out what got read
	out := <-outC

	if string(out) != "Should fail with arg 5 (error: A failure for testing)\n" {
		t.Errorf("%q != \"Should fail with arg 5 (error: An error for testing)\"", out)
	}
}
