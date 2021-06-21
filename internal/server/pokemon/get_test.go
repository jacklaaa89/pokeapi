package pokemon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jacklaaa89/pokeapi/internal/api/apitest/mock"
	"github.com/jacklaaa89/pokeapi/internal/api/format/json"
	"github.com/jacklaaa89/pokeapi/internal/api/log/fmt"
	"github.com/jacklaaa89/pokeapi/internal/pokeapi"
	"github.com/jacklaaa89/pokeapi/internal/server/middleware"
	"github.com/jacklaaa89/pokeapi/internal/translation"
)

// formatter the formatter to use for the mock API and when we decode.
var formatter = json.New()

type partialMockAPI interface {
	Close()
	AllExpectationsMet() error
}

// setup performs the test setup, sorting out the API clients
// to use mock API's.
func setup(fn func(m mock.API)) partialMockAPI {
	m := mock.NewMockAPI(formatter)
	fn(m)
	m.Start()

	pokemonAPI = pokeapi.NewWithEndpoint(m.URL())
	translationAPI = translation.NewWithEndpoint(m.URL(), "")
	return m
}

// withMiddleware wraps the supplied http.handler with all of the middlewares
func withMiddleware(h http.Handler, m ...mux.MiddlewareFunc) http.Handler {
	for _, md := range m {
		h = md.Middleware(h)
	}
	return h
}

// expectedResponse the expected response from the API
// on a valid response.
type expectedResponse struct {
	Data *SpeciesResponse `json:"data"`
}

// decode helper function to decode the response using the
func decode(t *testing.T, w *httptest.ResponseRecorder) *expectedResponse {
	r := new(expectedResponse)
	require.NoError(t, formatter.Decode(w.Body, r))
	return r
}

func TestGet(t *testing.T) {
	tt := []struct {
		Name     string
		Vars     map[string]string // URL variables.
		Setup    func(m mock.API)
		Expected func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			Name:  "NoNameSupplied",
			Setup: func(m mock.API) {},
			Expected: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
		{
			Name: "ErrorFromAPI",
			Vars: map[string]string{
				"name": "unknown",
			},
			Setup: func(m mock.API) {
				m.Expect("/pokemon-species/unknown", http.MethodGet).
					WithStatusCode(http.StatusNotFound)
			},
			Expected: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, w.Code)
			},
		},
		{
			Name: "Valid",
			Vars: map[string]string{
				"name": "mewtwo",
			},
			Setup: func(m mock.API) {
				m.Expect("/pokemon-species/mewtwo", http.MethodGet).
					WithResult(http.StatusOK, &pokeapi.Species{
						Name:        "mewtwo",
						IsLegendary: true,
						Habitat:     &pokeapi.NamedAPIResource{Name: "rare"},
						FlavorText: []*pokeapi.FlavorText{
							{Text: "a test description", Language: &pokeapi.NamedAPIResource{Name: "en"}},
						},
					})
			},
			Expected: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
				res := decode(t, w)
				assert.NotNil(t, res)
				assert.Equal(t, "a test description", res.Data.Description)
				assert.Equal(t, "rare", res.Data.Habitat)
				assert.True(t, res.Data.IsLegendary)
				assert.Equal(t, "mewtwo", res.Data.Name)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			p := setup(tc.Setup)
			defer p.Close()

			v := tc.Vars
			if v == nil {
				v = make(map[string]string)
			}

			// set up the function to get the URL parameters for tests.
			getVars = func(*http.Request) map[string]string {
				return v
			}

			w := httptest.NewRecorder()
			l := fmt.New(fmt.LevelNone)

			req, err := http.NewRequest(http.MethodGet, "/get", nil)
			require.NoError(st, err)

			ml := []mux.MiddlewareFunc{
				middleware.WithLogger(l),
				middleware.WithRequestID(),
			}

			withMiddleware(http.HandlerFunc(Get), ml...).ServeHTTP(w, req)

			tc.Expected(st, w)
			assert.NoError(st, p.AllExpectationsMet())
		})
	}
}
