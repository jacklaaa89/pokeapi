package auth

// BearerToken sets the supplied token as a Authorization Bearer token on each request.
func BearerToken(token string) Credentials { return FromHeader("Authorization", "Bearer "+token) }
