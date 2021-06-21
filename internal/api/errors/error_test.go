package errors

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

// newRequest generates a new HTTP request.
func newRequest(t *testing.T, method, path string, body io.ReadCloser) *http.Request {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := fmt.Sprintf("http://localhost:5555%s", path)
	req, err := http.NewRequestWithContext(
		context.Background(),
		method, u,
		body,
	)
	require.NoError(t, err)

	req.Header.Set(RequestIDHeader, "12345")
	if body == nil {
		return req
	}

	req.GetBody = func() (io.ReadCloser, error) {
		return body, nil
	}
	return req
}

// newRequestBody takes byte data and generates a io.ReadCloser
func newRequestBody(data []byte) io.ReadCloser { return io.NopCloser(bytes.NewBuffer(data)) }

// generateBodySample generates a byte body sample of size n
func generateBodySample(n int) []byte {
	b := make([]byte, n)
	for i := 0; i < len(b); i++ {
		b[i] = 'a'
	}
	return b
}

func TestFromResponse(t *testing.T) {
	tt := []struct {
		Name     string
		Req      *http.Request
		Res      *http.Response
		ResBody  io.ReadCloser
		Expected func(t *testing.T, err *Error)
	}{
		{
			Name: "Valid",
			Req: newRequest(
				t, http.MethodGet, "/path",
				newRequestBody(generateBodySample(100)),
			),
			Res:     &http.Response{StatusCode: http.StatusBadRequest},
			ResBody: newRequestBody(generateBodySample(100)),
			Expected: func(t *testing.T, err *Error) {
				// assert we did not need to clip the request or response
				// body data.
				assert.NotContains(t, err.Response, []byte("..."))
				assert.NotContains(t, err.Request, []byte("..."))
				assert.Equal(t, CodeInvalidRequest, err.Code)
				assert.Equal(t, http.StatusBadRequest, err.StatusCode)
				assert.Equal(t, "12345", err.RequestID)
				assert.Equal(t, http.MethodGet, err.Method)
				assert.Equal(t, "/path", err.Resource)
				assert.Empty(t, err.Source)
			},
		},
		{
			Name: "NoBody",
			Req: newRequest(
				t, http.MethodGet, "/path", nil,
			),
			Res:     &http.Response{StatusCode: http.StatusBadRequest},
			ResBody: newRequestBody(generateBodySample(100)),
			Expected: func(t *testing.T, err *Error) {
				// assert we did not need to clip the request or response
				// body data.
				assert.NotContains(t, err.Response, []byte("..."))
				assert.Nil(t, err.Request)
				assert.Equal(t, CodeInvalidRequest, err.Code)
				assert.Equal(t, http.StatusBadRequest, err.StatusCode)
				assert.Equal(t, "12345", err.RequestID)
				assert.Equal(t, http.MethodGet, err.Method)
				assert.Equal(t, "/path", err.Resource)
				assert.Empty(t, err.Source)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			tc.Expected(st, FromResponse(tc.Req, tc.Res, tc.ResBody))
		})
	}
}

// TestFromResponse_Codes test function specifically for
// testing that we get the correct error codes from different
// HTTP status codes.
func TestFromResponse_Codes(t *testing.T) {
	tt := []struct {
		Name       string
		StatusCode int
		Expected   Code
	}{
		{"StatusBadRequest", http.StatusBadRequest, CodeInvalidRequest},
		{"StatusForbidden", http.StatusForbidden, CodeForbidden},
		{"StatusUnauthorized", http.StatusUnauthorized, CodeUnauthorized},
		{"StatusNotFound", http.StatusNotFound, CodeNotFound},
		{"StatusInvalidOperation", http.StatusMethodNotAllowed, CodeInvalidOperation},
		{"StatusConflict", http.StatusConflict, CodeConflict},
		{"StatusPreconditionFailed", http.StatusPreconditionFailed, CodeInvalidContentType},
		{"StatusUnprocessableEntity", http.StatusUnprocessableEntity, CodeValidationError},
		{"StatusTooManyRequests", http.StatusTooManyRequests, CodeRateLimitExceeded},
		{"StatusInternalServerError", http.StatusInternalServerError, CodeServerError},
		{"StatusServiceUnavailable", http.StatusServiceUnavailable, CodeServerUnavailable},
		{"StatusGatewayTimeout", http.StatusGatewayTimeout, CodeServerUnavailable},
		{"StatusTeapot", http.StatusTeapot, CodeUnknownError},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			err := FromResponse(
				newRequest(st, http.MethodGet, "/path", nil),
				&http.Response{StatusCode: tc.StatusCode},
				nil,
			)
			assert.Equal(st, tc.Expected, err.Code)
		})
	}
}

// errorReader an io.Reader implementation which
// always returns an error on Read
type errorReader struct{}

// Read implements io.Reader interface.
func (*errorReader) Read([]byte) (n int, err error) {
	err = errors.New("read error")
	return
}

// Close implements Closer interface.
func (*errorReader) Close() error { return nil }

func TestRetrieveSample(t *testing.T) {
	tt := []struct {
		Name     string
		Input    io.ReadCloser
		Expected func(t *testing.T, b []byte)
	}{
		{
			Name:  "Valid",
			Input: newRequestBody(generateBodySample(100)),
			Expected: func(t *testing.T, b []byte) {
				assert.Len(t, b, 100)
			},
		},
		{
			Name:  "ValidWithOverflow",
			Input: newRequestBody(generateBodySample(700)),
			Expected: func(t *testing.T, b []byte) {
				assert.Len(t, b, bodySampleSize+3) // include the 3 dots
				assert.True(t, strings.HasSuffix(string(b), "..."))
			},
		},
		{
			Name:  "ErrorPerformingReadAll",
			Input: &errorReader{},
			Expected: func(t *testing.T, b []byte) {
				assert.Nil(t, b)
			},
		},
		{
			Name:  "NilReader",
			Input: nil,
			Expected: func(t *testing.T, b []byte) {
				assert.Nil(t, b)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			tc.Expected(st, RetrieveSample(tc.Input))
		})
	}
}

// expectedErrorJSON the expected error JSON structure.
const expectedErrorJSON = `{"code":"unknown_error","status_code":418,"method":"GET","resource":"/path","request_id":"12345","request":"YQ==","response":"YQ=="}`

func TestError_Error(t *testing.T) {
	err := &Error{
		Code:       CodeUnknownError,
		StatusCode: http.StatusTeapot,
		Method:     http.MethodGet,
		Resource:   "/path",
		RequestID:  "12345",
		Request:    generateBodySample(1),
		Response:   generateBodySample(1),
	}
	assert.Equal(t, expectedErrorJSON, err.Error())
}

func TestFromSource(t *testing.T) {
	tt := []struct {
		Name     string
		Input    error
		Expected func(t *testing.T, err *Error)
	}{
		{
			Name:  "WithSource",
			Input: errors.New("a request error"),
			Expected: func(t *testing.T, err *Error) {
				assert.Equal(t, CodeEncodingError, err.Code)
				assert.Equal(t, "/path", err.Resource)
				assert.Equal(t, http.MethodGet, err.Method)
				assert.Equal(t, "12345", err.RequestID)
				assert.Equal(t, "a request error", err.Source)
			},
		},
		{
			Name:  "WithoutSource",
			Input: nil,
			Expected: func(t *testing.T, err *Error) {
				assert.Equal(t, CodeEncodingError, err.Code)
				assert.Equal(t, "/path", err.Resource)
				assert.Equal(t, http.MethodGet, err.Method)
				assert.Equal(t, "12345", err.RequestID)
				assert.Empty(t, err.Source)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			tc.Expected(st, FromSource(CodeEncodingError, "/path", http.MethodGet, "12345", tc.Input))
		})
	}
}

func TestFromRequestAndSource(t *testing.T) {
	err := errors.New("a request error")
	apiErr := FromRequestAndSource(newRequest(t, http.MethodGet, "/path", nil), CodeRequestError, err)
	assert.Equal(t, CodeRequestError, apiErr.Code)
	assert.Equal(t, "/path", apiErr.Resource)
	assert.Equal(t, http.MethodGet, apiErr.Method)
	assert.Equal(t, "12345", apiErr.RequestID)
	assert.Equal(t, err.Error(), apiErr.Source)
}
