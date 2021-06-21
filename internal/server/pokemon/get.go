package pokemon

import (
	"context"
	"errors"
	"net/http"

	"github.com/jacklaaa89/pokeapi/internal/server/helpers"
	"github.com/jacklaaa89/pokeapi/internal/server/middleware"
)

// Get http.HandlerFunc which handles /pokemon/{name}
func Get(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	l := middleware.Logger(ctx)

	v := getVars(req)
	res, err := get(ctx, v["name"])
	if err != nil {
		l.Errorf(err.Error())
		helpers.RespondError(ctx, w, err)
		return
	}

	helpers.RespondOK(ctx, w, res)
}

// get attempts to retrieve details for a pokemon based on the name supplied.
func get(ctx context.Context, name string) (*SpeciesResponse, error) {
	if name == "" {
		return nil, helpers.InvalidRequest(errors.New("pokemon name is required"))
	}

	s, err := pokemonAPI.Pokemon.Species(ctx, name)
	if err != nil {
		return nil, err
	}

	return fromSpecies(s), nil
}
