package pokemon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jacklaaa89/pokeapi/internal/translation"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jacklaaa89/pokeapi/internal/api/apitest/mock"
	"github.com/jacklaaa89/pokeapi/internal/api/log/fmt"
	"github.com/jacklaaa89/pokeapi/internal/pokeapi"
	"github.com/jacklaaa89/pokeapi/internal/server/middleware"
)

// these three types are copied directly from the translation/types.go
// this is because these types are not exported in that package
// but I need the structure to mock the response.

// responseContents the response data for the translation.
type responseContents struct {
	Translated  string             `json:"translated"`  // Translated is the translated output.
	Text        string             `json:"text"`        // Text is the original input text
	Translation translation.Method `json:"translation"` // Translation is the translation method.
}

// responseSuccessData the data from the result which
// informs us how many total successful translations occurred.
type responseSuccessData struct {
	Total int64 `json:"total"`
}

// response this is the JSON structure which in which
// the translation API returns.
type response struct {
	Success  responseSuccessData `json:"success"`
	Contents responseContents    `json:"contents"`
}

func TestTranslated(t *testing.T) {
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
			Name: "ValidWithLegendary",
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

				// we apply yoda translation on legendary pokemon.
				m.Expect("/yoda.json", http.MethodGet).
					WithResult(http.StatusOK, &response{
						Success: responseSuccessData{
							Total: 1,
						},
						Contents: responseContents{
							Translated:  "a translated description",
							Translation: translation.Yoda,
							Text:        "a test description",
						},
					})
			},
			Expected: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
				res := decode(t, w)
				assert.NotNil(t, res)
				assert.Equal(t, "a translated description", res.Data.Description)
				assert.Equal(t, "rare", res.Data.Habitat)
				assert.True(t, res.Data.IsLegendary)
				assert.Equal(t, "mewtwo", res.Data.Name)

			},
		},
		{
			Name: "ValidWithCaveHabitat",
			Vars: map[string]string{
				"name": "mewtwo",
			},
			Setup: func(m mock.API) {
				m.Expect("/pokemon-species/mewtwo", http.MethodGet).
					WithResult(http.StatusOK, &pokeapi.Species{
						Name:        "mewtwo",
						IsLegendary: false,
						Habitat:     &pokeapi.NamedAPIResource{Name: "cave"},
						FlavorText: []*pokeapi.FlavorText{
							{Text: "a test description", Language: &pokeapi.NamedAPIResource{Name: "en"}},
						},
					})

				// we apply yoda translation on pokemon with the cave habitat.
				m.Expect("/yoda.json", http.MethodGet).
					WithResult(http.StatusOK, &response{
						Success: responseSuccessData{
							Total: 1,
						},
						Contents: responseContents{
							Translated:  "a translated description",
							Translation: translation.Yoda,
							Text:        "a test description",
						},
					})
			},
			Expected: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
				res := decode(t, w)
				assert.NotNil(t, res)
				assert.Equal(t, "a translated description", res.Data.Description)
				assert.Equal(t, "cave", res.Data.Habitat)
				assert.False(t, res.Data.IsLegendary)
				assert.Equal(t, "mewtwo", res.Data.Name)
			},
		},
		{
			Name: "ValidWithShakespeareTranslation",
			Vars: map[string]string{
				"name": "mewtwo",
			},
			Setup: func(m mock.API) {
				m.Expect("/pokemon-species/mewtwo", http.MethodGet).
					WithResult(http.StatusOK, &pokeapi.Species{
						Name:        "mewtwo",
						IsLegendary: false,
						Habitat:     &pokeapi.NamedAPIResource{Name: "rare"},
						FlavorText: []*pokeapi.FlavorText{
							{Text: "a test description", Language: &pokeapi.NamedAPIResource{Name: "en"}},
						},
					})

				// as the pokemon is not rare or a cave type, we apply the default.
				m.Expect("/shakespeare.json", http.MethodGet).
					WithResult(http.StatusOK, &response{
						Success: responseSuccessData{
							Total: 1,
						},
						Contents: responseContents{
							Translated:  "a translated description",
							Translation: translation.Shakespeare,
							Text:        "a test description",
						},
					})
			},
			Expected: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
				res := decode(t, w)
				assert.NotNil(t, res)
				assert.Equal(t, "a translated description", res.Data.Description)
				assert.Equal(t, "rare", res.Data.Habitat)
				assert.False(t, res.Data.IsLegendary)
				assert.Equal(t, "mewtwo", res.Data.Name)
			},
		},
		{
			Name: "ErrorOnTranslation",
			Vars: map[string]string{
				"name": "mewtwo",
			},
			Setup: func(m mock.API) {
				m.Expect("/pokemon-species/mewtwo", http.MethodGet).
					WithResult(http.StatusOK, &pokeapi.Species{
						Name:        "mewtwo",
						IsLegendary: false,
						Habitat:     &pokeapi.NamedAPIResource{Name: "rare"},
						FlavorText: []*pokeapi.FlavorText{
							{Text: "a test description", Language: &pokeapi.NamedAPIResource{Name: "en"}},
						},
					})

				// assert we can handle an error from the translation API.
				m.Expect("/shakespeare.json", http.MethodGet).WithStatusCode(http.StatusInternalServerError)
			},
			Expected: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
				res := decode(t, w)
				assert.NotNil(t, res)
				assert.Equal(t, "a test description", res.Data.Description)
				assert.Equal(t, "rare", res.Data.Habitat)
				assert.False(t, res.Data.IsLegendary)
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

			withMiddleware(http.HandlerFunc(Translated), ml...).ServeHTTP(w, req)

			tc.Expected(st, w)
			assert.NoError(st, p.AllExpectationsMet())
		})
	}
}
