package pokeapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jacklaaa89/pokeapi/internal/api/apitest/mock"
	"github.com/jacklaaa89/pokeapi/internal/api/format/json"
)

func TestNew(t *testing.T) {
	c := New()
	assert.NotNil(t, c)
	assert.NotNil(t, c.common)
	assert.NotNil(t, c.Pokemon)
}

func TestNewWithEndpoint(t *testing.T) {
	m := mock.NewMockAPI(json.New())

	// generate a simple expectation, just to check we set the endpoint correctly
	// when using NewWithEndpoint
	m.Expect("/pokemon-species/mewtwo", http.MethodGet).WithResult(http.StatusOK, &Species{
		Name:        "Mewtwo",
		IsLegendary: true,
		FlavorText: []*FlavorText{
			{
				Text:     "Test Description",
				Language: &NamedAPIResource{Name: "en"},
			},
		},
	})

	m.Start()
	defer m.Close()

	c := NewWithEndpoint(m.URL())
	s, err := c.Pokemon.Species(context.Background(), "mewtwo")
	assert.NotNil(t, s)
	assert.NoError(t, err)
	assert.NoError(t, m.AllExpectationsMet())
	assert.Equal(t, "Mewtwo", s.Name)
}
