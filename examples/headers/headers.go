package main

import (
	"net/http"

	"github.com/iqhater/pkg/examples"
	"github.com/iqhater/pkg/headers"
	"github.com/iqhater/pkg/middleware"
)

// Headers example
func main() {

	mid := middleware.Middlewares(
		headers.CORSHeaders(headers.CORSConfig{
			AllowOrigins: []string{"http://localhost:" + examples.HTTP_PORT},
			AllowMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodOptions,
				http.MethodHead,
			},
			AllowHeaders: []string{
				"Accept",
				"Content-Type",
				"Authorization",
			},
		}),
		headers.SecureHeaders,
		headers.ContentTypeHeaders("application/json"),
		middleware.Log,
	)

	http.HandleFunc("GET /headers-test", middleware.Bind(mid, func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("Headers CORS from server: Access-Control-Allow-Origin: " + w.Header().Get("Access-Control-Allow-Origin")))
		w.Write([]byte("Headers CORS from server: Access-Control-Allow-Methods: " + w.Header().Get("Access-Control-Allow-Methods")))
		w.Write([]byte("Headers secure from server: Content-Security-Policy: " + w.Header().Get("Content-Security-Policy")))
		w.Write([]byte("Headers content-type from server: Content-Type: " + w.Header().Get("application/json")))
	}))

	http.ListenAndServe(":"+examples.HTTP_PORT, nil)
}
