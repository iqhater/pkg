package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/iqhater/pkg/examples"
	"github.com/iqhater/pkg/headers"
	"github.com/iqhater/pkg/middleware"
)

/*
Execution order of middlewares:
 1. Recover ->
 2. Request ID ->
 3. CORS ->
 4. Logger ->
 5. Rate Limiter/Auth ->
 6. Context Timeout ->
 7. Secure Headers ->
 8. Compress ->
 9. Cache ->
 10. ContentType Headers ->
 11. Your handler
*/
// Middlewares example
func main() {

	cache := middleware.NewCache("10s")

	mid := middleware.Middlewares(
		middleware.Recover,
		middleware.RequestID,
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
		middleware.Log,
		middleware.Limit(2, 5),
		middleware.ContextTimeout(2*time.Second),
		headers.SecureHeaders,
		middleware.Compress, // gzip, brotli
		cache.CacheResponse,
		headers.ContentTypeHeaders("application/json"),
	)

	http.HandleFunc("GET /middlewares-test", middleware.Bind(mid, func(w http.ResponseWriter, req *http.Request) {

		// your heavy task handler
		// delay with context cancellation
		// must be used for ContextTimeout middleware
		ctx := req.Context()

		timeout := 3 * time.Second
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-ctx.Done():

			// handle context cancellation
			http.Error(w, http.StatusText(http.StatusRequestTimeout)+"\n"+ctx.Err().Error()+"\n"+errors.New("Handler with heavy task timeout!").Error(), http.StatusRequestTimeout)
			return
		}

		w.Write([]byte("Server Response OK!"))
	}))

	http.ListenAndServe(":"+examples.HTTP_PORT, nil)
}
