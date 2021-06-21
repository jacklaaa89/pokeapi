package translation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMethod_MarshalJSON(t *testing.T) {
	tt := []struct {
		Name     string
		Input    Method
		Expected func(t *testing.T, out string)
	}{
		{
			Name:  "Yoda",
			Input: Yoda,
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, `"yoda"`, out)
			},
		},
		{
			Name:  "Shakespeare",
			Input: Shakespeare,
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, `"shakespeare"`, out)
			},
		},
		{
			Name:  "InvalidMethod",
			Input: Method(-1),
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, `""`, out)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			b, err := tc.Input.MarshalJSON()
			assert.NoError(t, err)
			tc.Expected(st, string(b))
		})
	}
}

func TestMethod_Name(t *testing.T) {
	tt := []struct {
		Name     string
		Input    Method
		Expected func(t *testing.T, out string)
	}{
		{
			Name:  "Yoda",
			Input: Yoda,
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, `yoda`, out)
			},
		},
		{
			Name:  "Shakespeare",
			Input: Shakespeare,
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, `shakespeare`, out)
			},
		},
		{
			Name:  "InvalidMethod",
			Input: Method(-1),
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, "", out)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			tc.Expected(st, tc.Input.Name())
		})
	}
}

func TestMethod_String(t *testing.T) {
	tt := []struct {
		Name     string
		Input    Method
		Expected func(t *testing.T, out string)
	}{
		{
			Name:  "Yoda",
			Input: Yoda,
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, `yoda`, out)
			},
		},
		{
			Name:  "Shakespeare",
			Input: Shakespeare,
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, `shakespeare`, out)
			},
		},
		{
			Name:  "InvalidMethod",
			Input: Method(-1),
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, "", out)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			tc.Expected(st, tc.Input.String())
		})
	}
}

func TestMethod_UnmarshalJSON(t *testing.T) {
	tt := []struct {
		Name     string
		Input    string
		Expected func(t *testing.T, out Method, err error)
	}{
		{
			Name:  "Yoda",
			Input: `"yoda"`,
			Expected: func(t *testing.T, out Method, err error) {
				assert.Equal(t, Yoda, out)
				assert.NoError(t, err)
			},
		},
		{
			Name:  "Shakespeare",
			Input: `"shakespeare"`,
			Expected: func(t *testing.T, out Method, err error) {
				assert.Equal(t, Shakespeare, out)
				assert.NoError(t, err)
			},
		},
		{
			Name:  "UnQuoted",
			Input: "shakespeare",
			Expected: func(t *testing.T, out Method, err error) {
				assert.Equal(t, Method(0), out)
				assert.Error(t, err)
			},
		},
		{
			Name:  "JSONObject",
			Input: "{}",
			Expected: func(t *testing.T, out Method, err error) {
				assert.Equal(t, Method(0), out)
				assert.Error(t, err)
			},
		},
		{
			Name:  "InvalidMethod",
			Input: `"invalid"`,
			Expected: func(t *testing.T, out Method, err error) {
				assert.Equal(t, Method(0), out)
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			var method Method
			err := method.UnmarshalJSON([]byte(tc.Input))
			tc.Expected(st, method, err)
		})
	}
}
