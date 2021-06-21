package opts

import (
	"io"
	"net/http"
	"testing"
	"time"

	"golang.org/x/text/language"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_zap "go.uber.org/zap"

	"github.com/jacklaaa89/pokeapi/internal/api/auth"
	"github.com/jacklaaa89/pokeapi/internal/api/backoff"
	"github.com/jacklaaa89/pokeapi/internal/api/format"
	"github.com/jacklaaa89/pokeapi/internal/api/format/json"
	"github.com/jacklaaa89/pokeapi/internal/api/log"
	"github.com/jacklaaa89/pokeapi/internal/api/log/fmt"
	"github.com/jacklaaa89/pokeapi/internal/api/log/zap"
)

var config = _zap.NewProductionConfig()

type emptyEncoder struct{}

func (emptyEncoder) ContentType() string                   { return "" }
func (emptyEncoder) Accept() string                        { return "" }
func (emptyEncoder) Decode(io.Reader, interface{}) error   { return nil }
func (emptyEncoder) Encode(interface{}) (io.Reader, error) { return nil, nil }
func (emptyEncoder) EncodeTo(io.Writer, interface{}) error { return nil }

type emptyBackoff struct{}

func (emptyBackoff) Next(int) time.Duration { return 0 }

func TestWithBackoff(t *testing.T) {
	tt := []struct {
		Name     string
		Backoff  backoff.Backoff
		Expected func(t *testing.T, o *Options)
	}{
		{
			Name:    "NewBackoff",
			Backoff: &emptyBackoff{},
			Expected: func(t *testing.T, o *Options) {
				assert.IsType(t, (*emptyBackoff)(nil), o.Backoff)
				assert.NotNil(t, o.Backoff)
			},
		},
		{
			Name:    "Nil",
			Backoff: nil,
			Expected: func(t *testing.T, o *Options) {
				assert.IsType(t, backoff.Zero(), o.Backoff)
				assert.NotNil(t, o.Backoff)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			o := Apply(WithBackoff(tc.Backoff))
			tc.Expected(st, o)
		})
	}
}

func TestWithCredentials(t *testing.T) {
	c := auth.BasicAuth("username", "password")
	tt := []struct {
		Name        string
		Credentials auth.Credentials
		Expected    func(t *testing.T, o *Options)
	}{
		{
			Name:        "NewCredentials",
			Credentials: c,
			Expected: func(t *testing.T, o *Options) {
				assert.IsType(t, c, o.Credentials)
				assert.NotNil(t, o.Credentials)
			},
		},
		{
			Name:        "Nil",
			Credentials: nil,
			Expected: func(t *testing.T, o *Options) {
				assert.Nil(t, o.Credentials)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			o := Apply(WithCredentials(tc.Credentials))
			tc.Expected(st, o)
		})
	}
}

func TestWithEncoder(t *testing.T) {
	tt := []struct {
		Name     string
		Encoder  format.Encoder
		Expected func(t *testing.T, o *Options)
	}{
		{
			Name:    "NewEncoder",
			Encoder: &emptyEncoder{},
			Expected: func(t *testing.T, o *Options) {
				assert.IsType(t, (*emptyEncoder)(nil), o.Encoder)
				assert.NotNil(t, o.Encoder)
			},
		},
		{
			Name:    "Nil",
			Encoder: nil,
			Expected: func(t *testing.T, o *Options) {
				assert.IsType(t, json.New(), o.Encoder)
				assert.NotNil(t, o.Encoder)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			o := Apply(WithEncoder(tc.Encoder))
			tc.Expected(st, o)
		})
	}
}

func TestWithHTTPClient(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}
	tt := []struct {
		Name     string
		Client   *http.Client
		Expected func(t *testing.T, o *Options)
	}{
		{
			Name:   "NewClient",
			Client: client,
			Expected: func(t *testing.T, o *Options) {
				assert.Equal(t, client, o.HTTPClient)
				assert.Equal(t, 30*time.Second, o.HTTPClient.Timeout)
			},
		},
		{
			Name:   "Nil",
			Client: nil,
			Expected: func(t *testing.T, o *Options) {
				assert.Equal(t, http.DefaultClient, o.HTTPClient)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			o := Apply(WithHTTPClient(tc.Client))
			tc.Expected(st, o)
		})
	}
}

func TestWithLogger(t *testing.T) {
	zl, err := zap.New(&config)
	require.NoError(t, err)

	tt := []struct {
		Name     string
		Logger   log.Logger
		Expected func(t *testing.T, o *Options)
	}{
		{
			Name:   "ZapLogger",
			Logger: zl,
			Expected: func(t *testing.T, o *Options) {
				assert.NotNil(t, o.Logger)
				assert.IsType(t, zl, o.Logger)
			},
		},
		{
			Name:   "Nil",
			Logger: nil,
			Expected: func(t *testing.T, o *Options) {
				assert.NotNil(t, o.Logger)
				assert.IsType(t, fmt.New(fmt.LevelNone), o.Logger)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			o := Apply(WithLogger(tc.Logger))
			tc.Expected(st, o)
		})
	}
}

func TestWithMaxNetworkRetries(t *testing.T) {
	opts := Apply(WithMaxNetworkRetries(5))
	assert.Equal(t, int64(5), opts.MaxNetworkRetries)
}

func TestWithUserAgent(t *testing.T) {
	opts := Apply(WithUserAgent("user-agent"))
	assert.Equal(t, "user-agent", opts.UserAgent)
}

func TestWithLanguage(t *testing.T) {
	opts := Apply(WithLanguage(language.German))
	assert.Equal(t, language.German, opts.Language)
}

func TestWithTimeout(t *testing.T) {
	opts := Apply(WithTimeout(15 * time.Second))
	assert.Equal(t, 15*time.Second, opts.Timeout)
}
