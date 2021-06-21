package backoff

import "time"

// constantBackoff represents a backoff strategy which
// always waits the same amount of time between requests.
type constantBackoff struct {
	delay time.Duration
}

// Next implements Backoff interface
// returns the next delay to apply, this is always constant with this strategy.
func (c *constantBackoff) Next(_ int) time.Duration { return c.delay }

// Constant returns a backoff strategy which always waits the same amount of time between requests.
func Constant(delay time.Duration) Backoff { return &constantBackoff{delay} }

// Zero returns a backoff strategy which always allows
// no time for delay between attempts.
func Zero() Backoff { return Constant(0) }
