package mathx

import (
	"math"
	"math/rand"
)

// Random is a source of random values.
type Random struct {
	Src *rand.Rand // Pseudo-random source seeded with a given value.
}

// NewRandom returns a new Random that uses the provided seed to generate
// random values.
func NewRandom(seed int64) Random {
	src := rand.New(rand.NewSource(seed))
	return Random{Src: src}
}

// GetRandomInt returns a non-negative pseudo-random number in the interval [0, max).
// It returns 0 if max <= 0.
func (r *Random) GetRandomInt(max int) int {
	if max <= 0 {
		return 0
	}
	return r.Src.Intn(max)
}

// GetExpDistributedInt returns a exponentially distributed number in the interval
// [0, +math.MaxFloat64), rounded to the nearest int. Callers can adjust the rate of the
// function through the rate parameter.
func (r *Random) GetExpDistributedInt(rate float64) int {
	f := r.Src.ExpFloat64() / rate
	index := int(math.Round(f))
	return index
}
