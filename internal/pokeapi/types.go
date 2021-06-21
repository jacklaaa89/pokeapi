package pokeapi

import "strings"

// The names of these types have been copied to be directly compatible with the documentation defined at
// https://pokeapi.co/docs/v2
//
// This have been stripped down to specifically what is required in regards to this challenge.

// NamedAPIResource a referenced resource
type NamedAPIResource struct {
	Name string `json:"name"` // Name the name of the referenced resource.
}

// FlavorText a flavor of text describing a resource, defined in the specified language.
type FlavorText struct {
	Text     string            `json:"flavor_text"` // Text the localized flavor text for an API resource in a specific language.
	Language *NamedAPIResource `json:"language"`    // Language the language this name is in.
}

// Species represents the resource for a pokeapi pokemon.
// see: https://pokeapi.co/docs/v2#pokemon-species for more information
type Species struct {
	Name        string            `json:"name"`                // Name the name for this resource.
	IsLegendary bool              `json:"is_legendary"`        // IsLegendary whether or not this is a legendary Pokémon.
	Habitat     *NamedAPIResource `json:"habitat"`             // Habitat habitat this Pokémon pokemon can be encountered in.
	FlavorText  []*FlavorText     `json:"flavor_text_entries"` // FlavorText a list of flavor text entries for this Pokémon pokemon.
}

// Description attempts to find the first description for the pokemon for the supplied language
// or returns an empty string if not found.
func (s *Species) Description(lang string) string {
	if lang == "" || s == nil || len(s.FlavorText) == 0 {
		return ""
	}

	for _, ft := range s.FlavorText {
		if strings.EqualFold(lang, ft.Language.Name) {
			return ft.Text
		}
	}

	return ""
}
