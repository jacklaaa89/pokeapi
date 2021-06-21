package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBearerToken(t *testing.T) {
	req := newRequest(t, BearerToken("token"))
	assert.Equal(t, "Bearer token", req.Header.Get("Authorization"))
}
