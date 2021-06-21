package translation

import (
	"encoding/json"
	"strings"
)

const (
	// Shakespeare performs translations using the shakespeare transformation method.
	// see: https://funtranslations.com/api/shakespeare
	Shakespeare Method = iota + 1
	// Yoda performs translations using the yoda transformation method.
	// see: https://funtranslations.com/api/yoda
	Yoda
)

// methodMappings maps a translation method to their string representations.
var methodMappings = map[Method]string{
	Shakespeare: "shakespeare",
	Yoda:        "yoda",
}

// Method represents the translation method
type Method int

// Name returns the name of the translation method as a string
func (m Method) Name() string { return methodMappings[m] }

// String implements the fmt.Stringer interface
// returns the name.
func (m Method) String() string { return m.Name() }

// UnmarshalJSON implements the json.Unmarshaler interface
// takes byte input and converts it to a Method.
func (m *Method) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	for k, v := range methodMappings {
		if strings.EqualFold(s, v) {
			*m = k
			break
		}
	}

	return
}

// MarshalJSON implements the json.Marshaler interface
// converts a method into its encoded JSON form.
func (m Method) MarshalJSON() ([]byte, error) { return json.Marshal(m.Name()) }

// request the request to perform a translation.
type request struct {
	Text string `url:"text" json:"-"` // Text to translate
}

// responseContents the response data for the translation.
type responseContents struct {
	Translated  string `json:"translated"`  // Translated is the translated output.
	Text        string `json:"text"`        // Text is the original input text
	Translation Method `json:"translation"` // Translation is the translation method.
}

// responseSuccessData the data from the result which
// informs us how many total successful translations occurred.
type responseSuccessData struct {
	Total int64 `json:"total"`
}

// response this is the JSON structure which in which
// the translation API returns.
type response struct {
	Success  responseSuccessData `json:"success"`
	Contents responseContents    `json:"contents"`
}
