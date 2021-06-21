package auth

import "net/http"

// fromHeader allows us to set Credentials
// which are assigned to the requests URL query string.
type fromQuery struct {
	key, value string
}

// set implements Credentials interface.
// sets the key value as query parameter key and the value as the value
// using the Encode function on url.Values also performs query escaping.
func (f *fromQuery) set(r *http.Request) {
	v := r.URL.Query()
	v.Set(f.key, f.value)
	r.URL.RawQuery = v.Encode()
}

// FromQueryString generates credentials where the key is assigned to the query string.
func FromQueryString(key, value string) Credentials {
	if key == "" || value == "" {
		return nil
	}
	return &fromQuery{key, value}
}
