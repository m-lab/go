package logx_test

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/m-lab/go/logx"
	"github.com/m-lab/go/rtx"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func TestCaptureLog(t *testing.T) {
	out, err := logx.CaptureLog(nil, func() {
		logger := logx.NewLogEvery(nil, time.Millisecond)
		start := time.Now()
		log.Println("key phrase")
		for time.Since(start) < 10*time.Millisecond {
			logger.Println(time.Now())
		}
	})
	rtx.Must(err, "Error capturing log")

	if !strings.Contains(out, "logx_test.go:") {
		t.Error("Missing short filename")
	}

	if !strings.Contains(out, "key phrase") {
		t.Error("Missing key phrase")
	}
}

func TestLogEvery_Println(t *testing.T) {
	logger := log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
	out, err := logx.CaptureLog(logger, func() {
		logger := logx.NewLogEvery(logger, time.Millisecond)
		start := time.Now()
		for ; time.Since(start) < 10*time.Millisecond; time.Sleep(100 * time.Microsecond) {
			logger.Println("foobar")
		}
	})
	rtx.Must(err, "Error capturing log")

	if !strings.Contains(out, "logx.go:") {
		t.Error("Missing short filename")
	}
	lines := strings.Split(out, "\n")
	// Should be 10 or 11.  Error if more than 12.
	if len(lines) > 12 {
		t.Error("Too many logs", len(lines))
	}
	if len(lines) < 9 {
		t.Error("Too few logs", len(lines))
	}
}

func TestLogEvery_Printf(t *testing.T) {
	logger := log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
	out, err := logx.CaptureLog(logger, func() {
		logger := logx.NewLogEvery(logger, time.Millisecond)
		start := time.Now()
		for ; time.Since(start) < 10*time.Millisecond; time.Sleep(100 * time.Microsecond) {
			logger.Printf("%s\n", "foobar")
		}
	})
	rtx.Must(err, "Error capturing log")

	if !strings.Contains(out, "logx.go:") {
		t.Error("Missing short filename")
	}
	lines := strings.Split(out, "\n")
	// Should be 10 or 11.  Error if more than 12.
	if len(lines) > 12 {
		t.Error("Too many logs", len(lines))
	}
	if len(lines) < 9 {
		t.Error("Too few logs", len(lines))
	}

	// Test with log.std
	out, err = logx.CaptureLog(nil, func() {
		logger := logx.NewLogEvery(nil, time.Millisecond)
		start := time.Now()
		for ; time.Since(start) < 10*time.Millisecond; time.Sleep(100 * time.Microsecond) {
			logger.Printf("%s\n", "foobar")
		}
	})
	rtx.Must(err, "Error capturing log")

	if !strings.Contains(out, "logx.go:") {
		t.Error("Missing short filename")
	}
	lines = strings.Split(out, "\n")
	// Should be 10 or 11.  Error if more than 12.
	if len(lines) > 12 {
		t.Error("Too many logs", len(lines))
	}
	if len(lines) < 9 {
		t.Error("Too few logs", len(lines))
	}
}
