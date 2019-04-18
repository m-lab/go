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
	"io"
	"log"
	"os"
	"sync/atomic"
	"time"
	"unsafe"
)

// CaptureLog captures all output from log.Println, etc.
// Adapted from github.com/kami-zh
func CaptureLog(f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	log.SetOutput(w)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	f()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

// A Logger has basic Printf and Println functions.
type Logger interface {
	Println(v ...interface{})
	Printf(fmt string, v ...interface{})
}

type logEvery struct {
	lastTime unsafe.Pointer
	interval time.Duration
}

// NewLogEvery creates a logger that will log not more than once every interval.
func NewLogEvery(interval time.Duration) Logger {
	return &logEvery{unsafe.Pointer(&time.Time{}), interval}
}

func (le *logEvery) ok() bool {
	now := time.Now()
	oldPtr := atomic.LoadPointer(&le.lastTime)
	last := *(*time.Time)(oldPtr)
	if now.Sub(last) < le.interval {
		return false
	}
	// If this fails, then some other thread won the race.
	return atomic.CompareAndSwapPointer(&le.lastTime, oldPtr, unsafe.Pointer(&now))
}

func (le *logEvery) Println(v ...interface{}) {
	if le.ok() {
		log.Println(v...)
	}
}

// LogEvery takes an interval and pointer to a time.Time, and determines whether to produce the log or not.
func (le *logEvery) Printf(fmt string, v ...interface{}) {
	if le.ok() {
		log.Printf(fmt, v...)
	}
}
