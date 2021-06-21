package api

import (
	"context"
	_errors "errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/jacklaaa89/pokeapi/internal/api/errors"
	"github.com/jacklaaa89/pokeapi/internal/api/format/json"
	"github.com/jacklaaa89/pokeapi/internal/api/opts"
)

// validRequestData re-use the same structure for the request body.
type validRequestData = dummyResponseBody

// errorOnEncode a format.Encoder instance which always returns
// an error when encode is called.
type errorOnEncode struct{}

func (errorOnEncode) ContentType() string                 { return "" }
func (errorOnEncode) Accept() string                      { return "" }
func (errorOnEncode) Decode(io.Reader, interface{}) error { return nil }
func (errorOnEncode) Encode(interface{}) (io.Reader, error) {
	return nil, _errors.New("could not encode request data")
}
func (errorOnEncode) EncodeTo(io.Writer, interface{}) error {
	return _errors.New("could not encode request data")
}

// errorOnEncodeRead a format.Encoder instance which always returns
// an error when Read is called on the reader returned from Encode
type errorOnEncodeRead struct{}

func (errorOnEncodeRead) ContentType() string                      { return "" }
func (errorOnEncodeRead) Accept() string                           { return "" }
func (errorOnEncodeRead) Decode(io.Reader, interface{}) error      { return nil }
func (e *errorOnEncodeRead) Encode(interface{}) (io.Reader, error) { return e, nil }
func (errorOnEncodeRead) EncodeTo(io.Writer, interface{}) error    { return nil }
func (errorOnEncodeRead) Read([]byte) (n int, err error) {
	return 0, _errors.New("could not read encoded data")
}

func TestNew(t *testing.T) {
	const endpoint = "http://localhost:3333"
	c := New(endpoint)
	assert.Implements(t, (*Client)(nil), c)
	assert.Equal(t, endpoint, c.(*client).endpoint)
}

// assertRequestHeaders asserts the user-agent and request id headers are valid.
func assertRequestHeaders(t *testing.T, req *http.Request) {
	assert.Equal(t, "user-agent", req.Header.Get("User-Agent"))
	id := req.Header.Get(errors.RequestIDHeader)
	assert.NotEmpty(t, id)
	_, err := uuid.Parse(id)
	assert.NoError(t, err)
}

// TestClient_Call tests all of the scenarios which Client.Call should
// cover us for when using this library with a particular API.
func TestClient_Call(t *testing.T) {
	// closed will automatically be closed, causing any process
	// which uses it to error.
	closed, cancel := context.WithCancel(context.Background())
	cancel()

	tt := []struct {
		Name      string
		Opts      []opts.APIOption
		Method    string
		Path      string
		Context   context.Context
		Body      interface{}
		Receiver  interface{}
		OnRequest func(t *testing.T, attempts int, req *http.Request)
		OnCall    func(t *testing.T, rcv interface{}, err error)
	}{
		{
			Name:     "GET",
			Context:  context.Background(),
			Method:   http.MethodGet,
			Opts:     []opts.APIOption{opts.WithUserAgent("user-agent")},
			Path:     "/200", // http.StatusOK
			Receiver: new(dummyResponseBody),
			OnCall: func(t *testing.T, rcv interface{}, err error) {
				assert.NoError(t, err)

				// check we have correctly decoded the dummy response from the server.
				assert.IsType(t, (*dummyResponseBody)(nil), rcv)
				assert.NotEmpty(t, rcv.(*dummyResponseBody).Data)
			},
			OnRequest: func(t *testing.T, attempts int, req *http.Request) {
				assertRequestHeaders(t, req)
				assert.Empty(t, req.Header.Get("Content-Type"))
				assert.Equal(t, 1, attempts) // only one attempt should be made.
			},
		},
		{
			Name:     "PathWithNoStartingSlash",
			Context:  context.Background(),
			Method:   http.MethodGet,
			Opts:     []opts.APIOption{opts.WithUserAgent("user-agent")},
			Path:     "200", // http.StatusOK, should be rewritten to /200
			Receiver: new(dummyResponseBody),
			OnCall: func(t *testing.T, rcv interface{}, err error) {
				assert.NoError(t, err)

				// check we have correctly decoded the dummy response from the server.
				assert.IsType(t, (*dummyResponseBody)(nil), rcv)
				assert.NotEmpty(t, rcv.(*dummyResponseBody).Data)
			},
			OnRequest: func(t *testing.T, attempts int, req *http.Request) {
				assertRequestHeaders(t, req)
				assert.Empty(t, req.Header.Get("Content-Type"))
				assert.Equal(t, 1, attempts) // only one attempt should be made.
			},
		},
		{
			Name:    "POST",
			Context: context.Background(),
			Method:  http.MethodPost,
			Opts:    []opts.APIOption{opts.WithUserAgent("user-agent"), opts.WithEncoder(json.New())},
			Path:    "/200", // http.StatusOK
			OnCall: func(t *testing.T, rcv interface{}, err error) {
				assert.NoError(t, err)
			},
			OnRequest: func(t *testing.T, attempts int, req *http.Request) {
				assertRequestHeaders(t, req)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				assert.Equal(t, 1, attempts) // only one attempt should be made.
			},
		},
		{
			Name:     "EncodingError",
			Context:  context.Background(),
			Method:   http.MethodGet,
			Opts:     []opts.APIOption{opts.WithUserAgent("user-agent"), opts.WithEncoder(json.New())},
			Path:     "/200",          // http.StatusOK
			Receiver: make(chan bool), // cannot unmarshal into a channel, this will cause an encoding error.
			OnCall: func(t *testing.T, rcv interface{}, err error) {
				assert.Error(t, err)
				assert.IsType(t, (*errors.Error)(nil), err)
				assert.Equal(t, errors.CodeEncodingError, err.(*errors.Error).Code)
			},
			OnRequest: func(t *testing.T, attempts int, req *http.Request) {
				assertRequestHeaders(t, req)
				assert.Equal(t, 1, attempts) // only one attempt should be made.
			},
		},
		{
			Name:    "ErrorShouldTimeout",
			Context: context.Background(),
			Method:  http.MethodGet,
			Opts: []opts.APIOption{
				opts.WithUserAgent("user-agent"),
				opts.WithEncoder(json.New()),
				opts.WithTimeout(time.Nanosecond),
			},
			Path: "/200", // http.StatusOK
			OnCall: func(t *testing.T, rcv interface{}, err error) {
				assert.Error(t, err)
				assert.IsType(t, (*errors.Error)(nil), err)
				assert.Equal(t, errors.CodeHTTPClientError, err.(*errors.Error).Code)
				assert.Contains(t, err.Error(), "context deadline exceeded")
			},
			OnRequest: func(t *testing.T, attempts int, req *http.Request) {},
		},
		{
			Name:    "EncodingRequestBodyError",
			Context: context.Background(),
			Method:  http.MethodPost,
			Opts:    []opts.APIOption{opts.WithUserAgent("user-agent"), opts.WithEncoder(&errorOnEncode{})},
			Path:    "/200", // http.StatusOK
			OnCall: func(t *testing.T, rcv interface{}, err error) {
				assert.Error(t, err)
				assert.IsType(t, (*errors.Error)(nil), err)
				assert.Equal(t, errors.CodeEncodingError, err.(*errors.Error).Code)
			},
			OnRequest: func(t *testing.T, attempts int, req *http.Request) {
				assertRequestHeaders(t, req)
				assert.Equal(t, 1, attempts) // only one attempt should be made.
			},
		},
		{
			Name:    "HitMaxRetryAttempts",
			Context: context.Background(),
			Method:  http.MethodGet,
			Path:    "/500",                                          // http.StatusInternalServerError
			Opts:    []opts.APIOption{opts.WithMaxNetworkRetries(2)}, // should retry two times.
			OnCall: func(t *testing.T, rcv interface{}, err error) {
				assert.Error(t, err)
			},
			OnRequest: func(t *testing.T, attempts int, req *http.Request) {
				assert.LessOrEqual(t, attempts, 3) // 2 retries + the initial attempt.
			},
		},
		{
			Name:    "NoRetryOnRateLimitExceeded",
			Context: context.Background(),
			Method:  http.MethodGet,
			Path:    "/429",                                          // http.StatusTooManyRequests
			Opts:    []opts.APIOption{opts.WithMaxNetworkRetries(2)}, // should retry two times.
			OnCall: func(t *testing.T, rcv interface{}, err error) {
				assert.Error(t, err)
			},
			OnRequest: func(t *testing.T, attempts int, req *http.Request) {
				assert.LessOrEqual(t, attempts, 1) // should only try one when rate limit hit.
			},
		},
		{
			Name:    "RequestDataHasToBeStruct",
			Context: context.Background(),
			Method:  http.MethodGet,
			Path:    "/200", // http.StatusOk
			Body:    "i am not a struct",
			OnCall: func(t *testing.T, rcv interface{}, err error) {
				assert.Error(t, err)
			},
			OnRequest: func(t *testing.T, attempts int, req *http.Request) {
				assert.LessOrEqual(t, attempts, 1) // should only try one when rate limit hit.
			},
		},
		{
			Name:    "InvalidRequestContext",
			Context: nil, // cannot have a nil request context.
			Method:  http.MethodGet,
			Path:    "/200", // http.StatusOk
			OnCall: func(t *testing.T, rcv interface{}, err error) {
				assert.Error(t, err)
			},
			OnRequest: func(t *testing.T, attempts int, req *http.Request) {
				assert.LessOrEqual(t, attempts, 1) // should only try one when rate limit hit.
			},
		},
		{
			Name:    "ClosedContext",
			Context: closed,
			Method:  http.MethodGet,
			Path:    "/200", // http.StatusOk
			OnCall: func(t *testing.T, rcv interface{}, err error) {
				assert.Error(t, err)
			},
			OnRequest: func(t *testing.T, attempts int, req *http.Request) {
				assert.LessOrEqual(t, attempts, 1) // should only try one when rate limit hit.
			},
		},
		{
			Name:    "ErrorSettingBodyOnRequest",
			Context: closed,
			Method:  http.MethodPost,
			Body:    validRequestData{Data: "12345"},
			// the encoder returns a valid io.Reader on Encode, but that reader will error on the read.
			Opts: []opts.APIOption{opts.WithEncoder(&errorOnEncodeRead{})},
			Path: "/200", // http.StatusOk
			OnCall: func(t *testing.T, rcv interface{}, err error) {
				assert.Error(t, err)
			},
			OnRequest: func(t *testing.T, attempts int, req *http.Request) {
				assert.LessOrEqual(t, attempts, 1) // should only try one when rate limit hit.
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			endpoint, closer := newEchoServer(st, tc.OnRequest)
			defer closer()

			c := New(endpoint, tc.Opts...)
			err := c.Call(tc.Context, tc.Method, tc.Path, tc.Body, tc.Receiver)
			tc.OnCall(t, tc.Receiver, err)
		})
	}
}
