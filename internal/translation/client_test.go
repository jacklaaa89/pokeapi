package translation

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jacklaaa89/pokeapi/internal/api/apitest/mock"
	"github.com/jacklaaa89/pokeapi/internal/api/format/json"
)

func TestNew(t *testing.T) {
	c := New("")
	assert.NotNil(t, c)
}

func TestNewWithEndpoint(t *testing.T) {
	c := NewWithEndpoint(defaultURL, "")
	assert.NotNil(t, c)
}

func TestClient_Translate(t *testing.T) {
	tt := []struct {
		Name     string
		Setup    func(m mock.API)
		Input    string
		Method   Method
		Expected func(t *testing.T, output string, err error)
	}{
		{
			Name:   "YodaTranslation",
			Input:  "Master Obiwan has lost a planet.",
			Method: Yoda,
			Setup: func(m mock.API) {
				m.Expect("/yoda.json", http.MethodGet).WithResult(http.StatusOK, &response{
					Success: responseSuccessData{
						Total: 1,
					},
					Contents: responseContents{
						Translated:  "Lost a planet, master obiwan has.",
						Text:        "Master Obiwan has lost a planet.",
						Translation: Yoda,
					},
				})
			},
			Expected: func(t *testing.T, output string, err error) {
				assert.Equal(t, "Lost a planet, master obiwan has.", output)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "ShakespeareTranslation",
			Input:  "You gave Mr. Tim a hearty meal, but unfortunately what he ate made him die.",
			Method: Shakespeare,
			Setup: func(m mock.API) {
				m.Expect("/shakespeare.json", http.MethodGet).WithResult(http.StatusOK, &response{
					Success: responseSuccessData{
						Total: 1,
					},
					Contents: responseContents{
						Translated:  "Thee did giveth mr. Tim a hearty meal, but unfortunately what he did doth englut did maketh him kicketh the bucket.",
						Text:        "You gave Mr. Tim a hearty meal, but unfortunately what he ate made him die.",
						Translation: Shakespeare,
					},
				})
			},
			Expected: func(t *testing.T, output string, err error) {
				assert.Equal(t, "Thee did giveth mr. Tim a hearty meal, but unfortunately what he did doth englut did maketh him kicketh the bucket.", output)
				assert.NoError(t, err)
			},
		},
		{
			Name:   "InvalidMethod",
			Input:  "You gave Mr. Tim a hearty meal, but unfortunately what he ate made him die.",
			Method: Method(-1),
			Setup:  func(m mock.API) {},
			Expected: func(t *testing.T, output string, err error) {
				assert.Empty(t, output)
				assert.Error(t, err)
			},
		},
		{
			Name:   "ErrorFromAPI",
			Input:  "You gave Mr. Tim a hearty meal, but unfortunately what he ate made him die.",
			Method: Shakespeare,
			Setup: func(m mock.API) {
				m.Expect("/shakespeare.json", http.MethodGet).WithStatusCode(http.StatusUnauthorized)
			},
			Expected: func(t *testing.T, output string, err error) {
				assert.Empty(t, output)
				assert.Error(t, err)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			m := mock.NewMockAPI(json.New())
			tc.Setup(m)

			m.Start()
			defer m.Close()
			c := NewWithEndpoint(m.URL(), "")
			out, err := c.Translate(context.Background(), tc.Input, tc.Method)
			tc.Expected(st, out, err)
		})
	}
}
