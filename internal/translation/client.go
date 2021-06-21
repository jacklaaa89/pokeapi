package translation

import (
	"context"
	"errors"
	"net/http"

	"github.com/jacklaaa89/pokeapi/internal/api"
	"github.com/jacklaaa89/pokeapi/internal/api/auth"
	"github.com/jacklaaa89/pokeapi/internal/api/format/json"
	"github.com/jacklaaa89/pokeapi/internal/api/opts"
)

const (
	userAgent  = "translate/sdk-test"                        // userAgent is the User-Agent to send with each request.
	defaultURL = "https://api.funtranslations.com/translate" // defaultURL is the default endpoint to use.
	authHeader = "X-Funtranslations-Api-Secret"              // authHeader is the header to use for authorization, if applicable
)

// Client the translations client.
type Client struct{ c api.Client }

// Translate takes the input and performs the translation based on the method supplied.
func (c *Client) Translate(ctx context.Context, input string, method Method) (output string, err error) {
	r := new(response)
	formatted := method.String()
	if formatted == "" {
		return "", errors.New("invalid translation method provided: " + formatted)
	}
	path := api.FormatURLPath("/%s.json", formatted)
	err = c.c.Call(ctx, http.MethodGet, path, &request{input}, r)
	return r.Contents.Translated, err
}

// New initialises a new client using the default URL.
// token can be supplied as an empty string to use no authentication, this will restrict usage to
// the free plan.
func New(token string, o ...opts.APIOption) *Client { return NewWithEndpoint(defaultURL, token, o...) }

// NewWithEndpoint initialises a new client using the supplied URL.
// token can be supplied as an empty string to use no authentication, this will restrict usage to
// the free plan.
func NewWithEndpoint(endpoint, token string, o ...opts.APIOption) *Client {
	o = append(o,
		opts.WithEncoder(json.New()),
		opts.WithUserAgent(userAgent),
		opts.WithCredentials(auth.FromHeader(authHeader, token)),
	)
	c := &Client{c: api.New(endpoint, o...)}
	return c
}
