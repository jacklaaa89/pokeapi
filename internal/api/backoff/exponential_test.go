package backoff

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const zero = time.Duration(0)

// growthRateIncrease testNext function which assumes that the newly computed
// delay is at least more than double the last
// giving the indication of exponential growth.
//
// if the expected value exceeds the expected max value, then the new delay has to be the max
//
// there is a point where it will flat-line and always return the max.
func growthRateIncrease(rate float64, max time.Duration) func(t *testing.T, i int, last, delay time.Duration) {
	return func(t *testing.T, i int, last, delay time.Duration) {
		var expected time.Duration
		if i == 0 {
			assert.Greater(t, delay, expected)
			return
		}
		// the expectation is that we added the growth rate percentage from the last
		// amount to get the new amount.
		expected = last + time.Duration(float64(last)/100*rate)

		// if our expected value (after adding the growth rate) is greater
		// than the actual value, then we must have reached our upper bound limit
		// assert that the returned delay matches our expected max upper bound.
		if expected > delay {
			assert.Equal(t, max, delay)
			return
		}

		// jitter is applied to the delay making the returned delay
		// anywhere from 75% up to 100% of the actual exponential growth value.
		expected -= time.Duration(float64(expected) / 100 * 0.75)
		assert.GreaterOrEqual(t, delay, expected)
	}
}

func TestExponentialWithGrowthRate(t *testing.T) {
	tt := []struct {
		Name     string
		Min, Max time.Duration
		Growth   float64
		Expected func(t *testing.T, bk Backoff, err error)
		// tests the next computation
		// this function gets the current retry count,
		// the last computed delay, and the newly generated one
		// to determine the difference.
		TestNext func(t *testing.T, i int, last, delay time.Duration)
	}{
		{
			Name:   "Valid",
			Min:    50 * time.Millisecond,
			Max:    5000 * time.Millisecond,
			Growth: defaultGrowthRate,
			Expected: func(t *testing.T, bk Backoff, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, bk)
			},
			TestNext: growthRateIncrease(defaultGrowthRate, 5000*time.Millisecond),
		},
		{
			Name:   "GrowthExceedsMax",
			Min:    50 * time.Millisecond,
			Max:    60 * time.Millisecond,
			Growth: defaultGrowthRate,
			Expected: func(t *testing.T, bk Backoff, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, bk)
			},
			TestNext: growthRateIncrease(defaultGrowthRate, 60*time.Millisecond),
		},
		{
			Name:   "MinGreaterThanMax",
			Min:    5000 * time.Millisecond,
			Max:    50 * time.Millisecond,
			Growth: defaultGrowthRate,
			Expected: func(t *testing.T, bk Backoff, err error) {
				assert.Error(t, err)
				assert.Nil(t, bk)
			},
		},
		{
			Name:   "MinLessThanZero",
			Min:    zero - 1,
			Max:    50 * time.Millisecond,
			Growth: defaultGrowthRate,
			Expected: func(t *testing.T, bk Backoff, err error) {
				assert.Error(t, err)
				assert.Nil(t, bk)
			},
		},
		{
			Name:   "MaxLessThanZero",
			Min:    50 * time.Millisecond,
			Max:    zero - 1,
			Growth: defaultGrowthRate,
			Expected: func(t *testing.T, bk Backoff, err error) {
				assert.Error(t, err)
				assert.Nil(t, bk)
			},
		},
		{
			Name:   "MinEqualZero",
			Min:    zero,
			Max:    5000 * time.Millisecond,
			Growth: defaultGrowthRate,
			Expected: func(t *testing.T, bk Backoff, err error) {
				assert.Error(t, err)
				assert.Nil(t, bk)
			},
		},
		{
			Name:   "InvalidGrowthRate",
			Min:    50 * time.Millisecond,
			Max:    5000 * time.Millisecond,
			Growth: 500.00,
			Expected: func(t *testing.T, bk Backoff, err error) {
				assert.Error(t, err)
				assert.Nil(t, bk)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			bk, err := ExponentialWithGrowthRate(tc.Min, tc.Max, tc.Growth)
			tc.Expected(st, bk, err)
			if tc.TestNext == nil || err != nil {
				return
			}

			last := zero
			for i := 0; i < 5; i++ {
				delay := bk.Next(i)
				tc.TestNext(st, i, last, delay)
				last = delay
			}
		})
	}
}

func TestExponential(t *testing.T) {
	bk, err := Exponential(50*time.Millisecond, 5000*time.Millisecond)
	assert.NotNil(t, bk)
	assert.NoError(t, err)
	assert.Equal(t, defaultGrowthRate, bk.(*exponentialBackoff).growthRate)
}
