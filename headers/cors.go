package headers

import (
	"net/http"
	"strings"
)

type CORSConfig struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

// CustomHeaders middleware handler setup CORS and other headers
func CORSHeaders(config CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			// log.Println("3. CORS Headers middleware fire!")

			// CORS headers
			if len(config.AllowOrigins) > 0 {
				w.Header().Set(
					"Access-Control-Allow-Origin",
					strings.Join(config.AllowOrigins, ", "),
				)
			} else {
				// by default
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			if len(config.AllowMethods) > 0 {
				w.Header().Set(
					"Access-Control-Allow-Methods",
					strings.Join(config.AllowMethods, ", "),
				)
			} else {
				// by default
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			}

			if len(config.AllowHeaders) > 0 {
				w.Header().Set(
					"Access-Control-Allow-Headers",
					strings.Join(config.AllowHeaders, ", "),
				)
			} else {
				// by default
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")
			}

			// preflight requests immediately return 204 OK no content
			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return // to avoid "superfluous response.WriteHeader"
			}

			next.ServeHTTP(w, req)
		})
	}
}
