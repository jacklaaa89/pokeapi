package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromHeader(t *testing.T) {
	tt := []struct {
		Name       string
		Key, Value string
		Expected   func(t *testing.T, c Credentials, req *http.Request)
	}{
		{
			Name:  "Valid",
			Key:   "header",
			Value: "value",
			Expected: func(t *testing.T, c Credentials, req *http.Request) {
				assert.NotNil(t, c)
				assert.NotEmpty(t, req.Header.Get("header"))
				assert.Equal(t, "value", req.Header.Get("header"))
			},
		},
		{
			Name:  "NoHeaderKeyValue",
			Key:   "",
			Value: "value",
			Expected: func(t *testing.T, c Credentials, _ *http.Request) {
				assert.Nil(t, c)
			},
		},
		{
			Name:  "NoHeaderValue",
			Key:   "header",
			Value: "",
			Expected: func(t *testing.T, c Credentials, req *http.Request) {
				assert.Nil(t, c)
				assert.Empty(t, req.Header.Get("header"))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			c := FromHeader(tc.Key, tc.Value)
			req := newRequest(st, c)
			tc.Expected(st, c, req)
		})
	}
}
