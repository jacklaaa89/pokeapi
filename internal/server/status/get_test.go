package status

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	r := httptest.NewRecorder()
	Get(r, nil) // the request does not get used in this handler.
	assert.Equal(t, http.StatusOK, r.Code)
}
