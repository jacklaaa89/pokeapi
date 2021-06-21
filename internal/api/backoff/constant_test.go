package backoff

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstant(t *testing.T) {
	b := Constant(time.Second)

	// assert that every time we call Next it returns the same duration.
	for i := 0; i < randomBetweenRange(10, 20); i++ {
		assert.Equal(t, time.Second, b.Next(i))
	}
}

func TestZero(t *testing.T) {
	b := Zero()

	// assert that every time we call Next it returns zero
	for i := 0; i < randomBetweenRange(10, 20); i++ {
		assert.Equal(t, time.Duration(0), b.Next(i))
	}
}

// randomBetweenRange generates a random number between the bounds
// of min and max inclusively.
func randomBetweenRange(min, max int) int {
	return min + rand.Intn(max-min+1)
}
