package server

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/jacklaaa89/pokeapi/internal/server/pokemon"
	"github.com/jacklaaa89/pokeapi/internal/server/status"
)

// Handler defines the http.Handler to use with the server.
func Handler(middleware ...mux.MiddlewareFunc) http.Handler {
	m := mux.NewRouter()

	m.Use(middleware...)

	// === pokemon resource endpoints ===
	m.HandleFunc("/pokemon/{name}", pokemon.Get).
		Methods(http.MethodGet)
	m.HandleFunc("/pokemon/{name}/translated", pokemon.Translated).
		Methods(http.MethodGet)

	// === miscellaneous resource endpoints ===
	m.HandleFunc("/status", status.Get).
		Methods(http.MethodGet)

	return m
}
