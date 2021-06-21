package pokemon

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/jacklaaa89/pokeapi/internal/api"
	"github.com/jacklaaa89/pokeapi/internal/pokeapi"
	"github.com/jacklaaa89/pokeapi/internal/translation"
)

// varFunc a function which is used to retrieve URL parameters from a request.
type varFunc func(*http.Request) map[string]string

// getVars the function to use to get the URL parameters.
// its impossible to set-up mux for tests as it uses an internal
// const for the context key, so we can override the function here.
var getVars varFunc = mux.Vars

// cfgTranslationAPIKey the environment variable key to
// use to set the translation API key. defaults to an empty string
// if not defined.
const cfgTranslationAPIKey = "TRANSLATION_API_KEY"

// initialisations of the API clients to use.
var (
	pokemonAPI     = pokeapi.New()
	translationAPI = translation.New(os.Getenv(cfgTranslationAPIKey))
)

// SpeciesResponse the response from the /pokemon/{name} and
// /pokemon/{name}/translated
type SpeciesResponse struct {
	// Name the name of the Pokemon
	Name string `json:"name"`
	// Description is the first english description of the pokemon
	// this will be translated if required.
	Description string `json:"description"`
	// Habibat is the habitat in which the pokemon can be found.
	Habitat string `json:"habitat"`
	// IsLegendary determines if the pokemon is classed as a legendary pokemon.
	IsLegendary bool `json:"is_legendary"`
}

// fromSpecies takes the result from the poke-api and converts it into a structure
// which is encoded using JSON.
//
// this also attempts to find the first english description and normalise the text.
func fromSpecies(s *pokeapi.Species) *SpeciesResponse {
	return &SpeciesResponse{
		Name:        s.Name,
		Description: api.Normalise(s.Description("en")),
		Habitat:     s.Habitat.Name,
		IsLegendary: s.IsLegendary,
	}
}
