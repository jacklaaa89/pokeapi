// Package opts allows us to customise the settings to use with a payment provider API
//
// the settings defined here are platform agnostic and can be applied to all providers.
// applying settings in this fashion also allows us to set reasonable defaults in tests.
package opts

import (
	"net/http"
	"time"

	"golang.org/x/text/language"

	"github.com/jacklaaa89/pokeapi/internal/api/auth"
	"github.com/jacklaaa89/pokeapi/internal/api/backoff"
	"github.com/jacklaaa89/pokeapi/internal/api/format"
	"github.com/jacklaaa89/pokeapi/internal/api/format/json"
	"github.com/jacklaaa89/pokeapi/internal/api/log"
	"github.com/jacklaaa89/pokeapi/internal/api/log/fmt"
)

// zeroTimeout defines a zero timeout.
const zeroTimeout time.Duration = 0

// Options defines the options to use with the API.
type Options struct {
	HTTPClient        *http.Client     // HTTPClient is the http client to use.
	MaxNetworkRetries int64            // MaxNetworkRetries the maximum number of retries on API failure.
	Encoder           format.Encoder   // Encoder is the encoder to use to encode / decode requests and responses.
	UserAgent         string           // UserAgent the user agent to send with each request.
	Credentials       auth.Credentials // Credentials the method in which to authenticate if applicable.
	Backoff           backoff.Backoff  // Backoff allows us to customise the function to determine the backoff for retries.
	Logger            log.Logger       // Logger the logger to use when performing requests.
	Language          language.Tag     // Language language used to set the Accept-Language header.

	// Timeout defines the timeout which is applied to each request, a timeout of zero
	// represents no timeout is applied.
	Timeout time.Duration
}

// APIOption configures how we set up the API.
type APIOption interface {
	apply(*Options) // apply allows us to apply settings to an Options instance.
}

// funcAPIOption wraps a function that modifies apiOption into an
// implementation of the APIOption interface.
type funcAPIOption struct {
	f func(*Options) // f is the wrapped function.
}

// apply implements APIOption interface.
func (fdo *funcAPIOption) apply(do *Options) {
	fdo.f(do)
}

// newAPIOption generates a new APIOption from a function.
func newAPIOption(f func(*Options)) APIOption {
	return &funcAPIOption{
		f: f,
	}
}

// Apply applies the options and returns the generated options.
func Apply(opts ...APIOption) *Options {
	o := &Options{
		HTTPClient:        http.DefaultClient,
		MaxNetworkRetries: 2,
		Encoder:           json.New(),
		UserAgent:         "",
		Credentials:       nil,
		Backoff:           backoff.Zero(),
		Logger:            fmt.New(fmt.LevelNone),
		Language:          language.BritishEnglish,
		Timeout:           zeroTimeout,
	}

	for _, opt := range opts {
		opt.apply(o)
	}
	return o
}
