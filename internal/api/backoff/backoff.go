package backoff

import "time"

// Backoff allows us to customise the amount of delay (or backoff) between
// HTTP request retries.
type Backoff interface {
	// Next returns the next amount of backoff to wait before retrying
	// based on the amount of retries already attempted.
	Next(retries int) time.Duration
}
