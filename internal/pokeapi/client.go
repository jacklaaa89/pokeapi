package pokeapi

import (
	"github.com/jacklaaa89/pokeapi/internal/api"
	"github.com/jacklaaa89/pokeapi/internal/api/format/json"
	"github.com/jacklaaa89/pokeapi/internal/api/opts"
)

// defaultEndpoint the default endpoint to the pokeapi REST API.
const defaultEndpoint = "https://pokeapi.co/api/v2"

// userAgent the user-agent to send with each request.
const userAgent = "pokeapi/sdk-test"

// service a generic service type which all services inherit from.
//
// the api.Client here does all of the heavy lifting, i.e generating constructive helpers
// retries etc, we can simply wrap this service above for the specifics of the poke-api.
type service struct{ c api.Client }

// Client represents a api client to perform actions against the pokeapi V2.
type Client struct {
	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services used for talking to different parts of pokeapi V2.
	Pokemon *PokemonService
	// ... add more resource endpoints here when required.
}

// NewWithEndpoint initialises a new pokeapi client with a defined endpoint and a set of options.
func NewWithEndpoint(endpoint string, o ...opts.APIOption) *Client {
	o = append(o, opts.WithEncoder(json.New()), opts.WithUserAgent(userAgent))
	c := &Client{common: service{api.New(endpoint, o...)}}
	c.Pokemon = (*PokemonService)(&c.common)
	return c
}

// New initialises a new client with the default endpoint.
func New(o ...opts.APIOption) *Client { return NewWithEndpoint(defaultEndpoint, o...) }
