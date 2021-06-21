package auth

import (
	"context"
	"net/http"
)

type Credentials interface {
	set(r *http.Request)
}

// NewRequest wraps the standard http.NewRequestWithContext but also applies credentials when applicable.
func NewRequest(ctx context.Context, method, endpoint string, c Credentials) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil || c == nil {
		return req, err
	}

	c.set(req)
	return req, nil
}
