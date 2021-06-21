package pokeapi

import (
	"context"
	"net/http"

	"github.com/jacklaaa89/pokeapi/internal/api"
)

// PokemonService service which performs actions against pokemon resources.
type PokemonService service

// Species retrieves the pokemon information for a pokemon. The reference can either be the id or name
// of the pokemon to query.
//
// A Pokémon Species forms the basis for at least one Pokémon. Attributes of a Pokémon pokemon
// are shared across all varieties of Pokémon within the pokemon. A good example is Wormadam;
// Wormadam is the pokemon which can be found in three different varieties, Wormadam-Trash,
// Wormadam-Sandy and Wormadam-Plant.
//
// see: https://pokeapi.co/docs/v2#pokemon-species
func (p *PokemonService) Species(ctx context.Context, reference string) (s *Species, err error) {
	s = new(Species)
	path := api.FormatURLPath("/pokemon-species/%s/", reference)
	err = p.c.Call(ctx, http.MethodGet, path, nil, s)
	return
}
