package mathx

import (
	"math"
	"math/rand"
)

// GetRandomInt returns a non-negative pseudo-random number in the interval [0, max).
// It returns 0 if max <= 0.
func GetRandomInt(max int) int {
	if max <= 0 {
		return 0
	}
	return rand.Intn(max)
}

// GetExpDistributedInt returns a exponentially distributed number in the interval
// [0, +math.MaxFloat64), rounded to the nearest int. Callers can adjust the rate of the
// function through the rate parameter.
func GetExpDistributedInt(rate float64) int {
	f := rand.ExpFloat64() / rate
	index := int(math.Round(f))
	return index
}
