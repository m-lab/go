package memoryless_test

import (
	"context"
	"testing"
	"time"

	"github.com/m-lab/go/memoryless"
	"github.com/m-lab/go/rtx"
)

func TestBadArgs(t *testing.T) {
	f := func() { panic("should not be called") }
	for _, c := range []memoryless.Config{
		{Expected: -1},
		{Min: -1},
		{Max: -1},
		{Min: -3, Expected: -2, Max: -1},
		{Min: 1},
		{Min: 2, Max: 1},
		{Expected: 2, Max: 1},
		{Min: 2, Expected: 1},
	} {
		err := c.Errors()
		if err == nil {
			t.Errorf("Should have had an error with config %+v", c)
		}
		err = memoryless.Run(context.Background(), f, c)
		if err == nil {
			t.Errorf("Should have had an error running config %+v", c)
		}
		_, err = memoryless.NewTicker(context.Background(), c)
		if err == nil {
			t.Errorf("Should have had an error running config %+v", c)
		}
		_, err = memoryless.NewTimer(c)
		if err == nil {
			t.Errorf("Should have had an error running config %+v", c)
		}
	}
}

func TestRunOnce(t *testing.T) {
	count := 0
	f := func() { count++ }
	rtx.Must(
		memoryless.Run(context.Background(), f, memoryless.Config{Once: true}),
		"Bad time config")
	if count != 1 {
		t.Errorf("Once should mean once, not %d.", count)
	}
}

func TestRunForever(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// We use count rather than a waitgroup because an extra call to f() shouldn't
	// cause the test to fail - cancel() races with the timer, and that's both
	// fundamental and okay. Contexts can be canceled() multiple times no problem,
	// but if you ever call .Done() on a WaitGroup more times than you .Add(), you
	// get a panic.
	count := 1000
	f := func() {
		if count < 0 {
			cancel()
		} else {
			count--
		}
	}
	wt := time.Duration(1 * time.Microsecond)
	go memoryless.Run(ctx, f, memoryless.Config{Expected: wt, Min: wt, Max: wt})
	<-ctx.Done()
	// If this does not run forever, then f() was called at least 100 times and
	// then the context was canceled.
}

func TestLongRunningFunctions(t *testing.T) {
	// Make a ticker that fires many many times.
	wt := time.Duration(1 * time.Microsecond)
	ticker, err := memoryless.NewTicker(context.Background(), memoryless.Config{Expected: wt, Min: wt, Max: wt})
	rtx.Must(err, "Could not make ticker")
	time.Sleep(time.Millisecond)
	ticker.Stop()
	// Once ticker.Stop is called, lose all races.
	time.Sleep(100 * time.Millisecond)
	// Verify that no events are queued.
	count := 0
	for range ticker.C {
		count++
	}
	if count > 0 {
		t.Errorf("There should have been nothing in the channel, but instead there were %d items", count)
	}
}

func TestNewTimer(t *testing.T) {
	wt := time.Duration(1 * time.Millisecond)
	start := time.Now()
	timer, err := memoryless.NewTimer(memoryless.Config{Expected: wt, Min: wt, Max: wt})
	rtx.Must(err, "Could not make timer")
	channelTime := <-timer.C
	end := time.Now()
	diff := end.Sub(start)
	if diff < 1*time.Millisecond {
		t.Error("Did not wait at least 1ms:", diff)
	}
	if diff > 1*time.Second {
		// This check is potentially flaky if a cloud machine turns a 1ms sleep
		// into a 1s sleep for some reason. This seems unlikely, but every other
		// check in this function is a mathematical guarantee, so noting the
		// distant potential for flakiness with this check is a good idea.
		t.Error("Waited WAY more than 1ms:", diff)
	}
	if start.After(channelTime) || end.Before(channelTime) {
		t.Error("It should be:", start, "<=", channelTime, "<=", end)
	}
}
