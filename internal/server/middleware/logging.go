package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gorilla/mux"

	"github.com/jacklaaa89/pokeapi/internal/api/log"
)

// loggerContextKey the context key to use for the logger.
type loggerContextKey struct{}

// WithLogger generates a middleware which performs two actions:
// - assigns a logger instance in the context assigned to the request
// - perform access log logging.
//
// in order to get the deemed status code we need to record the response
// and then flush that into the waiting writer.
func WithLogger(l log.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// perform the next handler, recording the response.
			r := httptest.NewRecorder()
			t := time.Now()
			next.ServeHTTP(r, req.WithContext(withLogger(req.Context(), l)))
			l.Debugf("%s: %s - %d - %s", req.Method, req.URL.Path, r.Code, time.Since(t).String())

			// duplicate the response into the expecting writer.
			w.WriteHeader(r.Code)
			for k := range r.Header() {
				w.Header().Set(k, r.Header().Get(k))
			}
			w.Write(r.Body.Bytes())
		})
	}
}

// withLogger assigns a logger into a context.
func withLogger(ctx context.Context, l log.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, l)
}

// Logger helper function to retrieve the logger assigned to a context.
func Logger(ctx context.Context) log.Logger {
	return ctx.Value(loggerContextKey{}).(log.Logger)
}
