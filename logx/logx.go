// Package logx provides very useful utilities for logging.
// I hesitate to create it, for fear we will add too much stuff to it.  But LogEvery seems worth it.
// This packages uses the MIT license, to respect the corresponding license from kami-zh/go-capture
package logx

/*
Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the Software is furnished to
do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.
*/

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// This indirection allows breaking os.Pipe call for testing.
// Trying it out.  Seems a bit ugly.
var pipe = os.Pipe

// CaptureLog captures all output from log.Println, etc.
// Adapted from github.com/kami-zh
func CaptureLog(logger *log.Logger, f func()) (string, error) {
	r, w, err := pipe()
	if err != nil {
		return "", err
	}

	if logger != nil {
		writer := logger.Writer()
		defer func() {
			logger.SetOutput(writer)
		}()
		logger.SetOutput(w)
	} else {
		// Use the system default logger.
		// Unfortunately, we cannot get the current Writer from the log.std.  8-(
		defer func() {
			// NOTE: This may be troublesome if SetOutput has been called
			// elsewhere.
			log.SetOutput(os.Stderr)
		}()
		log.SetOutput(w)
	}

	f()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String(), nil
}

// A Logger has basic Printf and Println functions.
type Logger interface {
	Println(v ...interface{})
	Printf(fmt string, v ...interface{})
}

type logEvery struct {
	logger *log.Logger
	ticker *time.Ticker
}

// NewLogEvery creates a logger that will log not more than once every interval.
func NewLogEvery(logger *log.Logger, interval time.Duration) Logger {
	return &logEvery{logger: logger, ticker: time.NewTicker(interval)}
}

func (le *logEvery) ok() bool {
	select {
	case <-le.ticker.C:
		return true
	default:
		return false
	}
}

func (le *logEvery) Println(v ...interface{}) {
	if le.ok() {
		if le.logger != nil {
			le.logger.Output(2, fmt.Sprintln(v...))
		} else {
			log.Output(2, fmt.Sprintln(v...))
		}
	}
}

// LogEvery takes an interval and pointer to a time.Time, and determines whether to produce the log or not.
func (le *logEvery) Printf(format string, v ...interface{}) {
	if le.ok() {
		if le.logger != nil {
			le.logger.Output(2, fmt.Sprintf(format, v...))
		} else {
			log.Output(2, fmt.Sprintf(format, v...))
		}
	}
}
