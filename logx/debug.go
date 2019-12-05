package logx

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
)

var (
	// LogxDebug controls whether the debug logger is enabled or not. In almost all cases
	// you should rely on the flag to change this value. If critical, you may change the
	// value using LogxDebug.Set("true") or LogxDebug.Set("false").
	LogxDebug = logxDebug(false)
)

// Debug is the package logger for debug messages. Debug messages are discarded if LogxDebug is set false.
var Debug = log.New(ioutil.Discard, "DEBUG: ", 0)

func init() {
	flag.Var(&LogxDebug, "logx.debug", "Enable logx debug logging.")
}

// setup configures the Debug logger output based on the LogxLevel value.
func setup() {
	if LogxDebug {
		// Set output of the debug logger to a debug writer that uses the log package.
		Debug.SetOutput(&debugWriter{})
	} else {
		Debug.SetOutput(ioutil.Discard)
	}
}

// logxDebug is a bool flag that runs `setup` when the value is set.
type logxDebug bool

func (d logxDebug) Get() bool {
	return bool(d)
}

func (d *logxDebug) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*d = logxDebug(v)
	setup()
	return err
}

func (d logxDebug) String() string {
	return fmt.Sprintf("%t", d)
}

// debugWriter implements the io.Writer interface.
type debugWriter struct{}

// Write uses the standard log package log.Output to preserve atomic line writes.
func (w *debugWriter) Write(p []byte) (int, error) {
	// NOTE: calldepth depends on the log package implementation. It should ensure
	// that log.Lshortfile and log.Llongfile match the call site for logx.Debug.Print*.
	return len(p), log.Output(4, string(p))
}
