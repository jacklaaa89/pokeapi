package pokeapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecies_Description(t *testing.T) {
	tt := []struct {
		Name     string
		Language string
		Species  *Species
		Expected func(t *testing.T, out string)
	}{
		{
			Name:     "Valid",
			Language: "en",
			Species: &Species{
				FlavorText: []*FlavorText{
					{
						Text:     "this is a description",
						Language: &NamedAPIResource{Name: "en"},
					},
				},
			},
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, "this is a description", out)
			},
		},
		{
			Name:     "MultipleEnglish",
			Language: "en",
			Species: &Species{
				FlavorText: []*FlavorText{
					{
						Text:     "this is a description",
						Language: &NamedAPIResource{Name: "en"},
					},
					{
						Text:     "this is another description",
						Language: &NamedAPIResource{Name: "en"},
					},
				},
			},
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, "this is a description", out)
			},
		},
		{
			Name:     "NoEnglish",
			Language: "en",
			Species: &Species{
				FlavorText: []*FlavorText{
					{
						Text:     "dies ist eine Beschreibung",
						Language: &NamedAPIResource{Name: "de"},
					},
				},
			},
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, "", out)
			},
		},
		{
			Name:     "NoLanguageCode",
			Language: "",
			Species: &Species{
				FlavorText: []*FlavorText{
					{
						Text:     "this is a description",
						Language: &NamedAPIResource{Name: "en"},
					},
				},
			},
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, "", out)
			},
		},
		{
			Name:     "NoFlavourText",
			Language: "en",
			Species: &Species{
				FlavorText: nil,
			},
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, "", out)
			},
		},
		{
			Name:     "NoSpecies",
			Language: "en",
			Species:  nil,
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, "", out)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			tc.Expected(st, tc.Species.Description(tc.Language))
		})
	}
}
