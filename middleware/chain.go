package middleware

import (
	"net/http"
)

type Middleware func(next http.Handler) http.Handler

// Middlewares chains multiple middleware functions together.
// in order of execution
func Middlewares(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Bind applies the middleware to the route handler.
// This function takes a Middleware and an http.HandlerFunc (the route handler) and returns a new http.HandlerFunc that incorporates the middleware logic.
func Bind(mid Middleware, route http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		mid(http.HandlerFunc(route)).ServeHTTP(w, req)
	}
}
