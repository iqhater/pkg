package middleware

import (
	"context"
	"net/http"
	"time"
)

// ContextTimeout is a middleware that sets a timeout on the request context.
// It will be useful for long tasks such as IO or long queries at database.
// Strictly use with context cancellation in your handlers! Otherwise the middleware will be useless.
func ContextTimeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			// log.Println("6. Timeout middleware fire!")

			ctx, cancel := context.WithTimeout(req.Context(), duration)
			defer cancel()

			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}
