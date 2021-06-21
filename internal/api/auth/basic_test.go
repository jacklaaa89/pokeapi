package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newRequest generates a new request using the supplied credentials.
func newRequest(t *testing.T, c Credentials) (req *http.Request) {
	var err error
	req, err = NewRequest(context.Background(), http.MethodGet, "/path", c)
	require.NoError(t, err)
	return
}

func TestBasicAuth(t *testing.T) {
	req := newRequest(t, BasicAuth("username", "password"))
	u, p, ok := req.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, "username", u)
	assert.Equal(t, "password", p)
}
