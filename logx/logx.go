// Package logx provides very useful utilities for logging.
// I hesitate to create it, for fear we will add too much stuff to it.  But LogEvery seems worth it.
package logx

import (
	"sync/atomic"
	"time"
	"unsafe"
)

// A Logger has basic Printf and Println functions.
type Logger interface {
	Println(v ...interface{})
	Printf(fmt string, v ...interface{})
}

type logEvery struct {
	lastTime unsafe.Pointer
	interval time.Duration
}

// Should be handled with atomic update
var lastDecodeLogTime = unsafe.Pointer(&time.Time{})

// NewLogEvery creates a logger that will log not more than once every interval.
func NewLogEvery(interval time.Duration) Logger {
	return &logEvery{unsafe.Pointer(&time.Time{}), interval}
}

func (log *logEvery) Println(v ...interface{}) {
}

// LogEvery takes an interval and pointer to a time.Time, and determines whether to produce the log or not.
func (log *logEvery) Printf(fmt string, v ...interface{}) {
	now := time.Now()
	oldPtr := atomic.LoadPointer(&log.lastTime)
	last := *(*time.Time)(oldPtr)
	if now.Sub(last) < log.interval {
		return
	}
	if atomic.CompareAndSwapPointer(&log.lastTime, oldPtr, unsafe.Pointer(&now)) {
		log.Println(v...)
	}
}
