package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	h := Handler()
	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet,
		"/status", nil,
	)
	require.NoError(t, err)

	h.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
}
