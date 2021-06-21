package api

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
)

// the test file is solely used for initialisation of a mock HTTP server for use in other tests.

// callback the callback which is triggered on each request to the server.
type callback func(t *testing.T, attempts int, r *http.Request)

// dummyResponseBody a structure which emulates the response
// sent back in each request from the echo server.
type dummyResponseBody struct {
	Data string `json:"data" xml:"data"`
}

// echoHTTPHandler represents a http.HandlerFunc which
// given a path including a status code, it will respond with
// the same HTTP code.
//
// i.e /200 will return HTTP OK and /418 will return HTTP I'm a Teapot.
// if an invalid status code is given, then http.StatusBadRequest is sent.
//
// the callback fn is triggered on each request made to the handler with the HTTP request made
// and the amount of attempts made to this handler since generated.
func echoHTTPHandler(t *testing.T, fn callback) func(w http.ResponseWriter, req *http.Request) {
	var attempts = 0
	return func(w http.ResponseWriter, req *http.Request) {
		attempts++

		var code = http.StatusBadRequest
		v := mux.Vars(req)
		if c, ok := v["status"]; ok {
			if parsed, err := strconv.Atoi(c); err == nil {
				code = parsed
			}
		}

		// perform assertions.
		fn(t, attempts, req)
		// check the code is a valid HTTP status code
		if http.StatusText(code) == "" {
			code = http.StatusBadRequest
		}

		w.WriteHeader(code)
		json.NewEncoder(w).Encode(dummyResponseBody{Data: randomText(10)})
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// randomText generates random text of size n
func randomText(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// newEchoServer a function which initialises a test HTTP server
// using the echoHTTPHandler and returns the endpoint the server is listening
// on and a function to shut the server down.
//
// in the case that the route is not found then a 404 is returned.
func newEchoServer(t *testing.T, fn callback) (endpoint string, closer func()) {
	if fn == nil {
		fn = emptyCallback
	}
	m := mux.NewRouter()
	m.HandleFunc("/{status:[0-9]+}", echoHTTPHandler(t, fn))
	s := httptest.NewServer(m)
	return s.URL, s.Close
}

// emptyCallback a callback which performs no assertions.
func emptyCallback(*testing.T, int, *http.Request) {}
