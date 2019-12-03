package logx

import (
	"flag"
	"io/ioutil"
	"log"
)

var (
	// LogxDebug controls whether the debug logger is enabled or not.
	LogxDebug = false
)

// Debug is the package logger for debug messages. Before calling Setup(), debug messages are discarded.
var Debug = log.New(ioutil.Discard, "DEBUG: ", 0)

func init() {
	flag.BoolVar(&LogxDebug, "logx.debug", false, "Enable logx debug logging.")
}

// Setup should be called after flag.Parse(). If debug logging is enabled, Setup
// configures the package loggers based on the LogxLevel value.
func Setup() error {
	if LogxDebug {
		// Set output of the debug logger to a debug writer that uses the log package.
		Debug.SetOutput(&debugWriter{})
	} else {
		Debug.SetOutput(ioutil.Discard)
	}
	return nil
}

// debugWriter implements the io.Writer interface.
type debugWriter struct{}

// Write uses the standard log package log.Output to preserve atomic line writes.
func (w *debugWriter) Write(p []byte) (int, error) {
	// NOTE: calldepth depends on the log package implementation. It should ensure
	// that log.Lshortfile and log.Llongfile match the call site for logx.Debug.Print*.
	return len(p), log.Output(4, string(p))
}
