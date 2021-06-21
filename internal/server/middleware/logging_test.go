package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jacklaaa89/pokeapi/internal/api/log/fmt"
)

var logFormatRegex = regexp.MustCompile(`\[([A-Z]+)] ([\w/]+ [\w:]+) ([A-Z]+): (/[\w/\-_+.]+) - (\d+) - ([\d.]+.+)`)

func testLogHandler(t *testing.T) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		assert.NotPanics(t, func() {
			l := Logger(req.Context())
			assert.NotNil(t, l)
		})
		w.Header().Set("X-Assertion", "true") // set a custom header we can test is copied correctly.
	}
}

func TestWithLogger(t *testing.T) {
	b := &bytes.Buffer{}
	l := fmt.NewWithOutputs(fmt.LevelDebug, b, b)
	m := WithLogger(l)
	h := m.Middleware(http.HandlerFunc(testLogHandler(t)))

	r := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet,
		"/status", nil,
	)
	require.NoError(t, err)

	// run the test handler, also performs assertions on the
	// generated context assigned to the request.
	h.ServeHTTP(r, req)

	assert.Equal(t, "true", r.Header().Get("X-Assertion"))
	matches := logFormatRegex.FindStringSubmatch(b.String())
	count := len(matches)
	assert.NotZero(t, count)
	if count == 0 {
		return
	}

	// perform all assertions on the generated log output.

	assert.Equal(t, "DEBUG", matches[1])
	pt, err := time.Parse("2006/01/02 15:04:05", matches[2])
	assert.NoError(t, err)
	assert.False(t, pt.IsZero())
	assert.Equal(t, http.MethodGet, matches[3])
	assert.Equal(t, "/status", matches[4])
	code, err := strconv.Atoi(matches[5])
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	d, err := time.ParseDuration(matches[6])
	assert.NoError(t, err)
	assert.NotZero(t, d)

}

func TestLogger(t *testing.T) {
	l := fmt.New(fmt.LevelNone)
	assert.NotPanics(t, func() {
		ctx := withLogger(context.Background(), l)
		cl := Logger(ctx)
		assert.NotNil(t, cl)
		assert.IsType(t, l, cl)
	})

	assert.Panics(t, func() {
		Logger(context.Background())
	})
}
