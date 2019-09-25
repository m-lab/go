// Package memoryless helps repeated calls to a function be distributed across
// time in a memoryless fashion.
package memoryless

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Config represents the time we should wait between runs of the function.
//
// A valid config will have:
//  0 <= Min <= Expected <= Max (or 0 <= Min <= Expected and Max is 0)
// If Max is zero or unset, it will be ignored. If Min is zero or unset, it will
// be ignored.
type Config struct {
	// Expected records the expected/mean/average amount of time between runs.
	Expected time.Duration
	// Min provides clamping of the randomly produced value. All timers will wait
	// at least Min time.
	Min time.Duration
	// Max provides clamping of the randomly produced value. All timers will take
	// at most Max time.
	Max time.Duration

	// Once is provided as a helper, because frequently for unit testing and
	// integration testing, you only want the "Forever" loop to run once.
	//
	// The zero value of this struct has Once set to false, which means the value
	// only needs to be set explicitly in codepaths where it might be true.
	Once bool
}

func (c Config) waittime() time.Duration {
	wt := time.Duration(rand.ExpFloat64() * float64(c.Expected))
	if wt < c.Min {
		wt = c.Min
	}
	if c.Max != 0 && wt > c.Max {
		wt = c.Max
	}
	return wt
}

// Ticker is a struct that waits a config.Expected amount of time on average
// between sends down the channel C. It has the same interface and requirements
// as time.Ticker. Every Ticker created must have its Stop() method called or it
// will leak a goroutine.
//
// The inter-send time is a random variable governed by the exponential
// distribution and will generate a memoryless (Poisson) distribution of channel
// reads over time, ensuring that a measurement scheme using this ticker has the
// PASTA property (Poisson Arrivals See Time Averages). This statistical
// guarantee is subject to two caveats:
//
// Caveat 1 is that, in a nod to the realities of systems needing to have
// guarantees, we allow the random wait time to be clamped both above and below.
// This means that channel events will be at least config.Min and at most
// config.Max apart in time. This clamping causes bias in the timing. For use of
// Ticker to be statistically sensible, the clamping should not be too extreme.
// The exact mathematical meaning of "too extreme" depends on your situation,
// but a nice rule of thumb is config.Min should be at most 10% of expected and
// config.Max should be at least 250% of expected. These values mean that less
// than 10% of time you will be waiting config.Min and less than 10% of the time
// you will be waiting config.Max.
//
// Caveat 2 is that this assumes that the actions performed between channel
// reads take negligible time when compared to the expected wait time.
// Memoryless sequences have the property that the times between successive
// event starts has the exponential distribution, and the exponential
// distribution can generate numbers arbitrarily close to zero (albeit
// exponentially infrequently). This code will not send on the channel if the
// other end is not ready to receive, which provides another lower bound on
// inter-event times. The only other option if the other side of the channel is
// not ready to receive would be queueing events in the channel, and that has
// some pathological cases we would like to avoid. In particular, queuing can
// cause long-term correlations if the queue gets big, which is the exact
// opposite of what a memoryless system is trying to achieve.
type Ticker struct {
	C         <-chan time.Time // The channel on which the ticks are delivered.
	config    Config
	writeChan chan<- time.Time
	cancel    func()
}

func (t *Ticker) runTicker(ctx context.Context) {
	// No matter what, when this function exits the channel should never be written to again.
	defer close(t.writeChan)

	if t.config.Once {
		if ctx.Err() == nil {
			t.writeChan <- time.Now()
		}
		return
	}

	// When Done() is not closed and the Deadline has not been exceeded, the error
	// is nil.
	for ctx.Err() == nil {
		timer := time.NewTimer(t.config.waittime())
		// Just like time.Ticker, writes to the channel are non-blocking. If a user of
		// this module can't keep up with the timer they set, that's on them. There
		// are some potential pathological cases associated with queueing events in
		// the channel, and we want to avoid them.
		select {
		case t.writeChan <- time.Now():
		default:
		}
		// Wait until the timer is done or the context is canceled. If both conditions
		// are true, which case gets called is unspecified.
		select {
		case <-ctx.Done():
			// Clean up the timer. The timer will expire on its own and be garbage
			// collected if we don't explicitly stop it here, so this is not strictly
			// required, but it is nice to do.
			timer.Stop()
			// Please don't put code here that assumes that this code path will
			// definitely execute if the context is done. select {} doesn't promise that
			// multiple channels will get selected with equal probability, which means
			// that if the channel write takes long enough and config.Max is low enough,
			// then it could be true that the timer is done AND the context is canceled,
			// and we have no guarantee that in that case the canceled context case will
			// be the one that is selected.
		case <-timer.C:
		}
	}
}

// Stop the ticker goroutine.
func (t *Ticker) Stop() {
	t.cancel()
}

// MakeTicker creates a new memoryless ticker. The returned struct is compatible
// with the time.Ticker struct interface, and everywhere you use a time.Ticker,
// you can use a memoryless.Ticker.
func MakeTicker(ctx context.Context, config Config) (*Ticker, error) {
	if !(0 <= config.Min && config.Min <= config.Expected && (config.Max == 0 || config.Expected <= config.Max)) {
		return nil, fmt.Errorf(
			"The arguments to Run make no sense. It should be true that Min <= Expected <= Max (or Min <= Expected and Max is 0), "+
				"but that is not true for Min(%v) Expected(%v) Max(%v).",
			config.Min, config.Expected, config.Max)
	}
	c := make(chan time.Time)
	ctx, cancel := context.WithCancel(ctx)
	ticker := &Ticker{
		C:         c,
		config:    config,
		writeChan: c,
		cancel:    cancel,
	}
	go ticker.runTicker(ctx)
	return ticker, nil
}

// Run calls the given function repeatedly, using a memoryless.Ticker to wait
// between function calls. It is a convenience function for code that does not
// want to use the channel interface.
func Run(ctx context.Context, f func(), c Config) error {
	ticker, err := MakeTicker(ctx, c)
	if err != nil {
		return err
	}
	defer ticker.Stop()
	for range ticker.C {
		f()
	}
	return nil
}
