package headers

import (
	"net/http"

	"github.com/iqhater/pkg/middleware"
)

// CustomHeaders middleware handler setup content type headers
func ContentTypeHeaders(contentType string) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			// log.Println("10. ContentType Headers middleware fire!")

			if contentType == "" {
				// set default "application/json" header
				w.Header().Set("Content-Type", "application/json")
			} else {
				w.Header().Set("Content-Type", contentType)
			}
			next.ServeHTTP(w, req)
		})
	}
}
