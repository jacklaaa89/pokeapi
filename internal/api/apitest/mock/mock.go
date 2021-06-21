package mock

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/jacklaaa89/pokeapi/internal/api/format"
	"github.com/jacklaaa89/pokeapi/internal/api/format/json"
)

// Expectation represents a met expectation. Effectively the result of matching a path and HTTP method.
type Expectation interface {
	WithStatusCode(code int)               // WithStatusCode sets the status code for an expectation.
	WithResult(code int, body interface{}) // WithResult sets the status code and result body for an expectation.
}

// expectation the internal implementation of an expectation.
type expectation struct {
	code int         // code is the HTTP status code to respond with
	body interface{} // body is the response body to respond with.
}

// WithStatusCode sets the status code for an expectation.
func (e *expectation) WithStatusCode(code int) { e.code = code }

// WithResult sets the status code and result body for an expectation.
func (e *expectation) WithResult(code int, body interface{}) {
	e.WithStatusCode(code)
	e.body = body
}

// API represents a mock API.
type API interface {
	URL() string // URl retrieves the URL to connect to.
	Start()      // Start attempts to start the server
	Close()      // Close attempts to close the server

	Expect(path string, method string) Expectation // Expect adds a new expectation to the stack.
	AllExpectationsMet() error                     // AllExpectationsMet reports whether all expectations have been met or used.
}

// mockAPI represents the mock API implementation.
type mockAPI struct {
	sync.RWMutex
	*httptest.Server

	encoder      format.Encoder            // encoder the encoder to use for the response body
	serving      bool                      // serving is whether the server is running.
	expectations map[string][]*expectation // expectations is the set of expectations to run with
}

// URL attempts to get the URL of the running server.
func (m *mockAPI) URL() string {
	if !m.isServing() {
		panic("cannot retrieve endpoint from un-started server")
	}

	return m.Server.URL
}

// Start attempts to start the HTTP server
// this will panic if the server is already started.
func (m *mockAPI) Start() {
	if m.isServing() {
		return
	}

	m.Server.Start()
	m.Lock()
	defer m.Unlock()
	m.serving = true
}

// isServing reports if the server is marked as running
func (m *mockAPI) isServing() bool {
	m.RLock()
	defer m.RUnlock()
	return m.serving
}

// Close attempts to close the running HTTP server
func (m *mockAPI) Close() {
	if !m.isServing() {
		return
	}

	m.Server.Close()
	m.Lock()
	defer m.Unlock()
	m.serving = false
}

// pop given a http.Request, we attempt to find the next expectation in the
// stack for it, if an expectation is found it is popped from the top of the expectation slice for that path.
// if no expectation is found then a generic expectation which returns a 404 NOT FOUND error is used.
func (m *mockAPI) pop(req *http.Request) *expectation {
	path := fmt.Sprintf("%s-%s", strings.TrimSuffix(req.URL.Path, "/"), req.Method)
	m.Lock()
	s, ok := m.expectations[path]
	if !ok || len(s) == 0 {
		m.Unlock()
		return m.notFound()
	}
	next := s[0]
	m.expectations[path] = s[1:]
	if len(m.expectations[path]) == 0 {
		delete(m.expectations, path)
	}

	m.Unlock()

	if next == nil || next.code == 0 {
		return m.notFound()
	}
	return next
}

// notFound generates a 404 NOT FOUND expectation.
func (m *mockAPI) notFound() *expectation {
	return &expectation{code: http.StatusNotFound, body: nil}
}

// ServeHTTP implements the http.Handler interface
// allows us to handle HTTP requests by checking the expectation stack and
// responding with the assigned response.
func (m *mockAPI) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ex := m.pop(req)
	r, err := m.encoder.Encode(ex.body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := io.ReadAll(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(ex.code)
	w.Header().Set("Content-Type", m.encoder.ContentType())
	w.Write(b)
}

// Expect sets up an expectation on the HTTP server at the defined path
// optionally only allowing the supplied methods.
func (m *mockAPI) Expect(path string, method string) Expectation {
	if m.isServing() {
		panic("cannot add expectations once server is started")
	}

	path = strings.TrimSuffix(path, "/")
	key := fmt.Sprintf("%s-%s", path, method)
	ex := &expectation{code: http.StatusOK, body: nil}

	m.Lock()
	_, ok := m.expectations[key]
	if !ok {
		m.expectations[key] = make([]*expectation, 0)
	}

	m.expectations[key] = append(m.expectations[key], ex)
	m.Unlock()

	return ex
}

// AllExpectationsMet reports if there are any expectations which have not
// been popped from the stack.
func (m *mockAPI) AllExpectationsMet() (err error) {
	m.Lock()
	defer m.Unlock()
	if len(m.expectations) > 0 {
		err = errors.New("there are remaining expectations")
	}
	return
}

// NewMockAPI initialises a new mock api using the supplied encoder for the body.
func NewMockAPI(enc format.Encoder) API {
	if enc == nil {
		enc = json.New()
	}
	m := &mockAPI{
		expectations: make(map[string][]*expectation),
		encoder:      enc,
	}
	m.Server = httptest.NewUnstartedServer(m)
	return m
}
