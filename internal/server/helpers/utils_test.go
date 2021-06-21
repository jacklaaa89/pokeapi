package helpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jacklaaa89/pokeapi/internal/api/log/fmt"
	"github.com/jacklaaa89/pokeapi/internal/server/middleware"

	"github.com/gorilla/mux"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jacklaaa89/pokeapi/internal/api/format"
	"github.com/jacklaaa89/pokeapi/internal/api/format/xml"
)

// withMiddleware wraps the supplied http.handler with all of the middlewares
func withMiddleware(h http.Handler, m ...mux.MiddlewareFunc) http.Handler {
	for _, md := range m {
		h = md.Middleware(h)
	}
	return h
}

func TestRespondOK(t *testing.T) {
	r := httptest.NewRecorder()

	// produce a handler func to wrap the test execution in.
	// this is required as the RespondError handler requires the
	// request id is placed in the context from middleware.WithRequestID
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		RespondOK(req.Context(), w, "12345")
	})

	req, err := http.NewRequest(http.MethodGet, "/get", nil)
	require.NoError(t, err)

	h = withMiddleware(h, middleware.WithRequestID(), middleware.WithLogger(fmt.New(fmt.LevelNone)))
	h.ServeHTTP(r, req)

	res := &response{Data: ""}
	json.NewDecoder(r.Body).Decode(res)
	assert.Equal(t, "12345", res.Data)
}

func TestWithEncoder(t *testing.T) {
	orig := formatter

	tt := []struct {
		Name     string
		Encoder  format.Encoder
		Expected func(t *testing.T)
	}{
		{
			Name:    "NilEncoder",
			Encoder: nil,
			Expected: func(t *testing.T) {
				// assert nothing was changed
				assert.NotNil(t, formatter)
				assert.Implements(t, (*format.Encoder)(nil), formatter)
				assert.Equal(t, orig, formatter)
			},
		},
		{
			Name:    "NewEncoder",
			Encoder: xml.New(),
			Expected: func(t *testing.T) {
				// assert we changed to the XML formatter.
				assert.NotNil(t, formatter)
				assert.Implements(t, (*format.Encoder)(nil), formatter)
				assert.NotEqual(t, orig, formatter)
				assert.Equal(t, "application/xml", formatter.ContentType())
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			WithEncoder(tc.Encoder)
			tc.Expected(st)
		})
	}
}
