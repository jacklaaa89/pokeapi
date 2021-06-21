package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromQueryString(t *testing.T) {
	tt := []struct {
		Name       string
		Key, Value string
		Expected   func(t *testing.T, c Credentials, req *http.Request)
	}{
		{
			Name:  "Valid",
			Key:   "api_key",
			Value: "value",
			Expected: func(t *testing.T, c Credentials, req *http.Request) {
				assert.NotNil(t, c)
				v := req.URL.Query()
				assert.Equal(t, "value", v.Get("api_key"))
			},
		},
		{
			Name:  "ValidWithQueryEncoding",
			Key:   "api_key",
			Value: "url _encoding/values#+",
			Expected: func(t *testing.T, c Credentials, req *http.Request) {
				assert.NotNil(t, c)
				v := req.URL.Query()
				assert.Contains(t, req.URL.RawQuery, "url+_encoding%2Fvalues%23%2B")
				assert.Equal(t, "url _encoding/values#+", v.Get("api_key"))
			},
		},
		{
			Name:  "NoQueryKeyValue",
			Key:   "",
			Value: "value",
			Expected: func(t *testing.T, c Credentials, _ *http.Request) {
				assert.Nil(t, c)
			},
		},
		{
			Name:  "NoQueryValue",
			Key:   "api_key",
			Value: "",
			Expected: func(t *testing.T, c Credentials, req *http.Request) {
				assert.Nil(t, c)
				v := req.URL.Query()
				assert.Equal(t, "", v.Get("api_key"))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			c := FromQueryString(tc.Key, tc.Value)
			req := newRequest(st, c)
			tc.Expected(st, c, req)
		})
	}
}
