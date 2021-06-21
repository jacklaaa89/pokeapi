package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jacklaaa89/pokeapi/internal/api/errors"
)

type testRequestIDHandler struct {
	foundID string
}

func (h *testRequestIDHandler) handle(t *testing.T) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		assert.NotPanics(t, func() {
			parsed, err := uuid.Parse(RequestID(req.Context()))
			assert.NoError(t, err)
			h.foundID = parsed.String() // set the found value to assert the set header is the same.
		})
	}
}

func TestRequestID(t *testing.T) {
	m := WithRequestID()
	th := &testRequestIDHandler{foundID: ""} // set the found ID to empty
	h := m.Middleware(http.HandlerFunc(th.handle(t)))

	r := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet,
		"/status", nil,
	)
	require.NoError(t, err)

	// run the test handler, also performs assertions on the
	// generated context assigned to the request.
	// this will set the th.foundID variable.
	h.ServeHTTP(r, req)
	parsed, err := uuid.Parse(r.Header().Get(errors.RequestIDHeader))
	assert.NoError(t, err)
	assert.Equal(t, th.foundID, parsed.String())
}

func TestWithRequestID(t *testing.T) {
	assert.NotPanics(t, func() {
		ctx := withRequestID(context.Background(), "12345")
		id := RequestID(ctx)
		assert.Equal(t, "12345", id)
	})

	assert.Panics(t, func() {
		RequestID(context.Background())
	})
}
