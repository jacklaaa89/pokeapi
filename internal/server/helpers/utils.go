package helpers

import (
	"context"
	"net/http"
	"sync"

	"github.com/jacklaaa89/pokeapi/internal/api/format"
	"github.com/jacklaaa89/pokeapi/internal/api/format/json"
	"github.com/jacklaaa89/pokeapi/internal/server/middleware"
)

// formatter the formatter to use when responding
// allows us to change the format by just changing the formatter.
// we default to responding with JSON.
// the mutex is used to allow thread-safe changes of the formatter.
var (
	mu        sync.RWMutex
	formatter = json.New()
)

// response the response from the server.
type response struct {
	RequestID string         `json:"request_id"`      // RequestID is the generated id for the request
	Error     *errorResponse `json:"error,omitempty"` // Error is any error that occurred, if applicable
	Data      interface{}    `json:"data,omitempty"`  // Data is the response data.
}

// WithEncoder function which allows us to change the encoder to use
// when responding to requests.
func WithEncoder(f format.Encoder) {
	if f == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	formatter = f
}

// RespondOK responds by writing the supplied data to the supplied http.ResponseWriter with a http.StatusOK
func RespondOK(ctx context.Context, w http.ResponseWriter, r interface{}) {
	write(ctx, w, http.StatusOK, &response{RequestID: middleware.RequestID(ctx), Data: r})
}

// write helper function to write the supplied http response using the
// encoder. This is thread-safe.
func write(ctx context.Context, w http.ResponseWriter, code int, r interface{}) {
	l := middleware.Logger(ctx)

	mu.RLock()
	defer mu.RUnlock()

	w.Header().Set("Content-Type", formatter.ContentType())
	w.WriteHeader(code)
	if err := formatter.EncodeTo(w, r); err != nil {
		l.Errorf("could not encode receiver into response: %v", err)
	}
}
