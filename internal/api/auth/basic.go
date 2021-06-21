package auth

import (
	"net/http"
)

// basicAuth is the Credentials implementation
// where a username and password are set on each request
type basicAuth struct {
	username string // username is the username to set.
	password string // password is the password to set.
}

// set implements Credentials interface
// sets the stored username and password on the request using basic auth.
func (b *basicAuth) set(r *http.Request) { r.SetBasicAuth(b.username, b.password) }

// BasicAuth generates a Credentials set which sets the supplied username and
// password via basic authentication.
func BasicAuth(username, password string) Credentials {
	return &basicAuth{username, password}
}
