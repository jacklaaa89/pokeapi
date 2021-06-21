package backoff

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// defaultGrowthRate this is the default exponential growth rate to apply,
// it represents 50% growth.
const defaultGrowthRate float64 = 0.5

// exponentialBackoff represents a backoff policy
// which makes the delay get exponentially larger after
// each failed attempt starting from the initial value
// up to the max value.
type exponentialBackoff struct {
	// growthRate the percentage represented between 0 and 1
	// which represents the rate of change.
	growthRate   float64
	initial, max time.Duration
}

// Next implements Backoff interface.
// returns the next delay to apply taking into account how many retries
// have been attempts thus far. Jitter is applied to the determined value
// to make the values seem more random.
// if the defined delay is greater than the then the max is used
//
// exponential growth is defined typically as: f(x) = a(1+r)^x
// where:
// 		a is the initial value
//		r is the growth rate
//		x is the time interval (in our case our retry count)
//
// see: https://en.wikipedia.org/wiki/Exponential_growth
func (e *exponentialBackoff) Next(retries int) time.Duration {
	a := float64(e.initial)
	x := float64(retries)

	if x == 0 {
		return time.Duration(a) // f(0) = a
	}

	b := 1 + e.growthRate

	fx := a * math.Pow(b, x)
	delay := time.Duration(fx)

	// apply jitter to the determined value
	// making the actual value between 75% and 100% of the actual delay
	jitter := rand.Int63n(int64(delay / 4))
	delay -= time.Duration(jitter)

	if delay > e.max {
		delay = e.max
	}

	return delay
}

// Exponential applies an exponential type strategy from the initial interval
// up to the defined max using the default growth rate.
func Exponential(initial, max time.Duration) (Backoff, error) {
	return ExponentialWithGrowthRate(initial, max, defaultGrowthRate)
}

// ExponentialWithGrowthRate applies an exponential type strategy from the initial interval
// up to the defined max using the supplied growth rate which should be between 0.1 and 1.0 (1% - 100% growth)
func ExponentialWithGrowthRate(initial, max time.Duration, growth float64) (Backoff, error) {
	// check that min is greater than time.Nanosecond, but less than max
	// this also transitively checks that max is greater than min
	// and time.Nanosecond.
	//
	// we cannot use zero as the initial value, as then every single next calculation will
	// also be zero
	if !isWithinDurationRange(initial, time.Nanosecond, max) {
		return nil, errors.New("initial must be greater than one nanosecond, but less than max")
	}

	// the growth rate is a percentage from 1 - 100% so this has to be between 0.01 and 1.0
	// again using a growth rate of zero will also return zero, which is incorrect.
	if !isWithinRange(growth, 0.01, 1) {
		return nil, errors.New("growth rate has to be a percentage between 1 and 100%")
	}
	return &exponentialBackoff{growth, initial, max}, nil
}

// isWithinRange checks if the float64 d is within the bounds
// of min and max.
func isWithinRange(d, min, max float64) bool {
	if d >= min && d <= max {
		return true
	}
	return false
}

// isWithinDurationRange helper function to perform isWithinRange function with durations.
func isWithinDurationRange(d, min, max time.Duration) bool {
	return isWithinRange(float64(d), float64(min), float64(max))
}
