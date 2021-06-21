package auth

import "net/http"

// fromHeader allows us to set Credentials
// for a custom header value.
type fromHeader struct {
	key, value string
}

// set implements Credentials interface.
// sets the key value as the header key and the value as the header value.
func (f *fromHeader) set(r *http.Request) {
	r.Header.Set(f.key, f.value)
}

// FromHeader generates credentials which allow us to set the header key and value
// on the request.
func FromHeader(key, value string) Credentials {
	if key == "" || value == "" {
		return nil
	}

	return &fromHeader{key, value}
}
