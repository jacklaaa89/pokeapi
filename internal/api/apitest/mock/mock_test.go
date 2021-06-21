package mock

// even though the mock API code is only used in tests
// ensuring that this library works as expected using tests
// confirms that we dont have any false positives in tests which use
// this library based on any bugs with this.
//
// this confirms that the mock API works as expected when using it in tests.

import (
	"context"
	_errors "errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jacklaaa89/pokeapi/internal/api/format"
	"github.com/jacklaaa89/pokeapi/internal/api/format/json"
)

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

func TestNewMockAPI(t *testing.T) {

}

func TestExpectation_WithResult(t *testing.T) {
	ex := &expectation{}
	ex.WithResult(http.StatusOK, "12345")
	assert.Equal(t, http.StatusOK, ex.code)
	assert.Equal(t, "12345", ex.body)
}

func TestExpectation_WithStatusCode(t *testing.T) {
	ex := &expectation{}
	ex.WithStatusCode(http.StatusOK)
	assert.Equal(t, http.StatusOK, ex.code)
	assert.Nil(t, ex.body)
}

func TestMockAPI_AllExpectationsMet(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/status", nil)
	require.NoError(t, err)

	tt := []struct {
		Name     string
		Setup    func(path, method string, m API)
		Expected func(t *testing.T, m API)
	}{
		{
			Name:  "NoInitialExpectations",
			Setup: func(path, method string, m API) {},
			Expected: func(t *testing.T, m API) {
				assert.NoError(t, m.AllExpectationsMet())
			},
		},
		{
			Name: "AllExpectationsConsumed",
			Setup: func(path, method string, m API) {
				m.Expect(path, method).WithStatusCode(http.StatusOK)
			},
			Expected: func(t *testing.T, m API) {
				assert.NoError(t, m.AllExpectationsMet())
			},
		},
		{
			Name: "HasRemainingExpectationsInStack",
			Setup: func(path, method string, m API) {
				// attach two expectations for the same path, method.
				m.Expect(path, method).WithStatusCode(http.StatusOK)
				m.Expect(path, method).WithStatusCode(http.StatusOK)
			},
			Expected: func(t *testing.T, m API) {
				assert.Error(t, m.AllExpectationsMet())
			},
		},
		{
			Name: "HasRemainingExpectationPath",
			Setup: func(path, method string, m API) {
				m.Expect(path, method).WithStatusCode(http.StatusOK)
				m.Expect(path+"-test", http.MethodGet).WithStatusCode(http.StatusOK)
			},
			Expected: func(t *testing.T, m API) {
				assert.Error(t, m.AllExpectationsMet())
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			m := NewMockAPI(nil)
			tc.Setup(req.URL.Path, req.Method, m)
			w := httptest.NewRecorder()
			m.(*mockAPI).ServeHTTP(w, req) // consume a single expectation, if matched.
			tc.Expected(st, m)
		})
	}
}

func TestMockAPI_Close(t *testing.T) {
	m := NewMockAPI(nil)
	m.Start()
	assert.True(t, m.(*mockAPI).serving)

	m.Close()
	assert.False(t, m.(*mockAPI).serving)

	// ensure calling close again has no affect.
	m.Close()
	assert.False(t, m.(*mockAPI).serving)
}

func TestMockAPI_Expect(t *testing.T) {
	tt := []struct {
		Name        string
		Test        func(m API)
		ShouldPanic bool
		Expected    func(t *testing.T, m *mockAPI)
	}{
		{
			Name:        "Started",
			ShouldPanic: true,
			Test: func(m API) {
				m.Start()
				m.Expect("/path", http.MethodGet)
			},
			Expected: func(t *testing.T, m *mockAPI) {},
		},
		{
			Name:        "NotStarted",
			ShouldPanic: false,
			Test: func(m API) {
				m.Expect("/path", http.MethodGet)
			},
			Expected: func(t *testing.T, m *mockAPI) {
				assert.Len(t, m.expectations, 1)
				assert.Len(t, m.expectations["/path-GET"], 1)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			m := NewMockAPI(nil)
			if tc.ShouldPanic {
				assert.Panics(st, func() {
					tc.Test(m)
				})
				return
			}

			tc.Test(m)
			tc.Expected(t, m.(*mockAPI))
		})
	}
}

func TestMockAPI_ServeHTTP(t *testing.T) {
	tt := []struct {
		Name     string
		Encoder  format.Encoder
		Setup    func(path, method string, m API)
		Expected func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			Name:    "WithValidExpectation",
			Encoder: json.New(),
			Setup: func(path, method string, m API) {
				m.Expect(path, method).WithStatusCode(http.StatusOK)
			},
			Expected: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, "null\n", w.Body.String()) // we set no body in our expectation.
			},
		},
		{
			Name:    "WithNoExpectation",
			Encoder: json.New(),
			Setup:   func(path, method string, m API) {},
			Expected: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, w.Code)
			},
		},
		{
			Name:    "ErrorEncoding",
			Encoder: &errorOnEncode{},
			Setup:   func(path, method string, m API) {},
			Expected: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, w.Code)
			},
		},
		{
			Name:    "ErrorEncodeRead",
			Encoder: &errorOnEncodeRead{},
			Setup:   func(path, method string, m API) {},
			Expected: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, w.Code)
			},
		},
		{
			Name:    "ErrorInvalidStatusCode",
			Encoder: json.New(),
			Setup: func(path, method string, m API) {
				m.Expect(path, method).WithStatusCode(0)
			},
			Expected: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, w.Code)
			},
		},
		{
			Name:    "NilEncoder",
			Encoder: nil,
			Setup: func(path, method string, m API) {
				m.Expect(path, method).WithStatusCode(http.StatusOK)
			},
			Expected: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/status", nil)
	require.NoError(t, err)

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			m := NewMockAPI(tc.Encoder).(*mockAPI)
			tc.Setup(req.URL.Path, req.Method, m)
			w := httptest.NewRecorder()

			// we dont have to start the server here, as we
			// are calling the handler directly.
			m.ServeHTTP(w, req)
			tc.Expected(st, w)
		})
	}
}

func TestMockAPI_Start(t *testing.T) {
	m := NewMockAPI(nil)
	m.Start()
	assert.True(t, m.(*mockAPI).serving)

	// ensure calling start again has no affect.
	m.Start()
	assert.True(t, m.(*mockAPI).serving)
}

func TestMockAPI_URL(t *testing.T) {
	m := NewMockAPI(nil)
	assert.Panics(t, func() {
		m.URL()
	})

	m.Start()
	defer m.Close()
	assert.NotEmpty(t, m.URL())
	u, err := url.Parse(m.URL())
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1", u.Hostname())
}
