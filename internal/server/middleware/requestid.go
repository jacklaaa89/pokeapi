package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/jacklaaa89/pokeapi/internal/api/errors"
)

// requestIDContextKey the context key to use for the request id.
type requestIDContextKey struct{}

// WithRequestID middleware function which generates a new request id
// assigns is to the request context and sets it as a response header.
func WithRequestID() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			id := uuid.New().String()
			w.Header().Set(errors.RequestIDHeader, id)
			next.ServeHTTP(w, req.WithContext(withRequestID(req.Context(), id)))
		})
	}
}

// withRequestID generates a context with the request id assigned as a value.
func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, id)
}

// RequestID attempts to retrieve the request id from the supplied context.
func RequestID(ctx context.Context) string {
	return ctx.Value(requestIDContextKey{}).(string)
}
