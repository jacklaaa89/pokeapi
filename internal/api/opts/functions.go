package opts

import (
	"net/http"
	"time"

	"golang.org/x/text/language"

	"github.com/jacklaaa89/pokeapi/internal/api/auth"
	"github.com/jacklaaa89/pokeapi/internal/api/backoff"
	"github.com/jacklaaa89/pokeapi/internal/api/format"
	"github.com/jacklaaa89/pokeapi/internal/api/log"
)

// WithUserAgent updates the API options with the supplied user-agent.
func WithUserAgent(ua string) APIOption {
	return newAPIOption(func(o *Options) {
		o.UserAgent = ua
	})
}

// WithHTTPClient sets the http.Client to use for each request.
// this can be used to finely tweak TLS settings or add a custom http.RoundTripper
func WithHTTPClient(h *http.Client) APIOption {
	return newAPIOption(func(o *Options) {
		if h == nil {
			return
		}
		o.HTTPClient = h
	})
}

// WithEncoder overrides the request and response encoder for the client.
func WithEncoder(f format.Encoder) APIOption {
	return newAPIOption(func(o *Options) {
		if f == nil {
			return
		}
		o.Encoder = f
	})
}

// WithMaxNetworkRetries overrides the max number of retries to attempt.
func WithMaxNetworkRetries(n int64) APIOption {
	return newAPIOption(func(o *Options) {
		o.MaxNetworkRetries = n
	})
}

// WithCredentials overrides the credentials to use.
func WithCredentials(c auth.Credentials) APIOption {
	return newAPIOption(func(o *Options) {
		o.Credentials = c
	})
}

// WithBackoff overrides the backoff strategy to in between retries.
func WithBackoff(b backoff.Backoff) APIOption {
	return newAPIOption(func(o *Options) {
		if b == nil {
			return
		}
		o.Backoff = b
	})
}

// WithLogger overrides the logger to use.
func WithLogger(l log.Logger) APIOption {
	return newAPIOption(func(o *Options) {
		if l == nil {
			return
		}
		o.Logger = l
	})
}

// WithLanguage overrides the language to use.
func WithLanguage(l language.Tag) APIOption {
	return newAPIOption(func(o *Options) {
		o.Language = l
	})
}

// WithTimeout overrides the per-request timeout.
func WithTimeout(t time.Duration) APIOption {
	return newAPIOption(func(o *Options) {
		o.Timeout = t
	})
}
