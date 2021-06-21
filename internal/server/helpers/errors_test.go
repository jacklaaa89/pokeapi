package helpers

import (
	_errors "errors"
	_fmt "fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jacklaaa89/pokeapi/internal/api/errors"
	"github.com/jacklaaa89/pokeapi/internal/api/format"
	"github.com/jacklaaa89/pokeapi/internal/api/format/json"
	"github.com/jacklaaa89/pokeapi/internal/api/log/fmt"
	"github.com/jacklaaa89/pokeapi/internal/server/middleware"
)

type errorOnEncode struct{}

func (errorOnEncode) ContentType() string                     { return "content-type" }
func (errorOnEncode) Accept() string                          { return "content-type" }
func (errorOnEncode) Decode(_ io.Reader, _ interface{}) error { return nil }
func (errorOnEncode) EncodeTo(w io.Writer, i interface{}) error {
	return _errors.New("could not encode")
}
func (errorOnEncode) Encode(interface{}) (io.Reader, error) {
	return nil, _errors.New("could not encode")
}

func TestInvalidRequest(t *testing.T) {
	err := _fmt.Errorf("a test error")
	irErr := InvalidRequest(err)
	assert.Implements(t, (*compoundError)(nil), irErr)
	assert.Equal(t, errors.CodeInvalidRequest, irErr.(compoundError).Code())
	assert.Equal(t, http.StatusBadRequest, irErr.(compoundError).StatusCode())
}

func TestRespondError(t *testing.T) {
	tt := []struct {
		Name     string
		Encoder  format.Encoder
		Error    error
		Expected func(t *testing.T, code int, err *errorResponse)
	}{
		{
			Name:  "StdErr",
			Error: _fmt.Errorf("a typical error"),
			Expected: func(t *testing.T, code int, err *errorResponse) {
				assert.Equal(t, http.StatusInternalServerError, code)
				assert.Equal(t, errors.CodeServerError, err.Code)
				assert.Equal(t, "a typical error", err.Error)
			},
		},
		{
			Name:    "CouldNotEncode",
			Encoder: &errorOnEncode{},
			Error:   _fmt.Errorf("a typical error"),
			Expected: func(t *testing.T, code int, err *errorResponse) {
				assert.Nil(t, err)
			},
		},
		{
			Name:  "NilErr",
			Error: nil,
			Expected: func(t *testing.T, code int, err *errorResponse) {
				assert.Nil(t, err)
				assert.Equal(t, http.StatusOK, code) // the default
			},
		},
		{
			Name:  "CompoundErr",
			Error: InvalidRequest(_fmt.Errorf("a typical error")),
			Expected: func(t *testing.T, code int, err *errorResponse) {
				assert.Equal(t, http.StatusBadRequest, code)
				assert.Equal(t, errors.CodeInvalidRequest, err.Code)
				assert.Equal(t, "a typical error", err.Error)
			},
		},
		{
			Name: "APIErr",
			Error: &errors.Error{
				Code:       errors.CodeEncodingError,
				StatusCode: http.StatusInternalServerError,
				Method:     http.MethodGet,
				Resource:   "/path",
				RequestID:  "12345",
				Source:     "could not decode response",
			},
			Expected: func(t *testing.T, code int, err *errorResponse) {
				assert.Equal(t, http.StatusInternalServerError, code)
				assert.Equal(t, errors.CodeEncodingError, err.Code)
				assert.Equal(t, string(errors.CodeEncodingError), err.Error)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			r := httptest.NewRecorder()

			enc := tc.Encoder
			if enc == nil {
				enc = json.New()
			}

			// get the server to respond using the default encoder.
			WithEncoder(enc)

			// produce a handler func to wrap the test execution in.
			// this is required as the RespondError handler requires the
			// request id is placed in the context from middleware.WithRequestID
			var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				RespondError(req.Context(), w, tc.Error)
			})

			req, err := http.NewRequest(http.MethodGet, "/get", nil)
			require.NoError(st, err)

			h = withMiddleware(h, middleware.WithRequestID(), middleware.WithLogger(fmt.New(fmt.LevelNone)))
			h.ServeHTTP(r, req)

			var res = new(response)
			err = enc.Decode(r.Body, res)
			if err != io.EOF {
				require.NoError(st, err)
			}
			tc.Expected(st, r.Code, res.Error)
		})
	}
}
