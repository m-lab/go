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

func TestCaptureLog(t *testing.T) {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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

func TestCaptureLog_BadPipe(t *testing.T) {
	logx.BadPipeForTest()
	defer logx.RestorePipeForTest()
	_, err := logx.CaptureLog(nil, func() {
		logger := logx.NewLogEvery(nil, time.Millisecond)
		start := time.Now()
		log.Println("key phrase")
		for time.Since(start) < 10*time.Millisecond {
			logger.Println(time.Now())
		}
	})
	if err == nil {
		t.Fatal("Should have gracefully handled error")
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

	if !strings.Contains(out, "logx_test.go:") {
		t.Error("Missing short filename", out)
	}
	lines := strings.Split(out, "\n")
	// Should be 10 or 11.  Error if more than 12.
	if len(lines) > 12 {
		t.Error("Too many logs", len(lines))
	}
	if len(lines) == 0 {
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

	if !strings.Contains(out, "logx_test.go:") {
		t.Error("Missing short filename")
	}
	lines := strings.Split(out, "\n")
	// Should be 10 or 11.  Error if more than 12.
	if len(lines) > 12 {
		t.Error("Too many logs", len(lines))
	}
	if len(lines) == 0 {
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

	if !strings.Contains(out, "logx_test.go:") {
		t.Error("Missing short filename")
	}
	lines = strings.Split(out, "\n")
	// Should be 10 or 11.  Error if more than 12.
	if len(lines) > 12 {
		t.Error("Too many logs", len(lines))
	}
	if len(lines) == 0 {
		t.Error("Too few logs", len(lines))
	}
}

// cpu: Intel(R) Core(TM) i7-7920HQ CPU @ 3.10GHz
// Before using ticker:    33018	     37712 ns/op	   40000 B/op	    2000 allocs/op
// After using ticker: 	  113064	     10403 ns/op	   16000 B/op	    1000 allocs/op
func BenchmarkLogEvery(b *testing.B) {
	b.ReportAllocs()
	le := logx.NewLogEvery(nil, time.Minute)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {

			for i := 0; i < 1000; i++ {
				le.Println("foobar")
			}
		}
	})
}
