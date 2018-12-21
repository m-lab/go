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

// Run calls the given function repeatedly, waiting a c.Expected amount of time
// between calls on average. The wait time is actually random and will generate
// a memoryless (Poisson) distribution of f() calls in time, ensuring that f()
// has the PASTA property (Poisson Arrivals See Time Averages). This statistical
// guarantee is subject to two caveats.
//
// Caveat 1 is that, in a nod to the realities of systems needing to have
// guarantees, we allow the random wait time to be clamped both above and below.
// This means that calls to f() should be at least c.Min and at most c.Max apart
// in time. This clamping causes bias in the timing. For use of this function to
// be statistically sensible, the clamping should not be too extreme. The exact
// mathematical meaning of "too extreme" depends on your situation, but a nice
// rule of thumb is c.Min should be at most 10% of expected and c.Max should be
// at least 250% of expected. These values mean that less than 10% of time you
// will be waiting c.Min and less than 10% of the time you will be waiting
// c.Max.
//
// Caveat 2 is that this assumes that the function f() takes negligible time to
// run when compared to the expected wait time. Technically memoryless events
// have the property that the times between successive event starts has the
// exponential distribution, and this code will not start a new call to f()
// before the old one has completed, which provides a lower bound on wait times.
func Run(ctx context.Context, f func(), c Config) error {
	if !(0 <= c.Min && c.Min <= c.Expected && (c.Max == 0 || c.Expected <= c.Max)) {
		return fmt.Errorf(
			"The arguments to Run make no sense. It should be true that Min <= Expected <= Max (or Min <= Expected and Max is 0), "+
				"but that is not true for Min(%v) Expected(%v) Max(%v).",
			c.Min, c.Expected, c.Max)
	}
	if c.Once {
		f()
		return nil
	}
	// When Done() is not closed and the Deadline has not been exceeded, the error
	// is nil.
	for ctx.Err() == nil {
		// Start the timer before the function call because the time between function
		// call *starts* should be exponentially distributed.
		t := time.NewTimer(c.waittime())
		f()
		// Wait until the timer is done or the context is canceled. If both conditions
		// are true, which case gets called is unspecified.
		select {
		case <-ctx.Done():
			// Clean up the timer.
			t.Stop()
			// Please don't put logic here that assumes that this code path will
			// definitely execute if the context is done. select {} doesn't promise that
			// multiple channels will get selected with equal probability, which means
			// that if f() takes a while and c.Max is low, then it could be true that the
			// timer is done AND the context is canceled, and we have no guarantee that
			// in that case the canceled context case will be the one that is selected.
		case <-t.C:
		}
	}
	return nil
}
