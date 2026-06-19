package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Recover middleware recovers panic and debug stacktrace
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// log.Println("1. Recover middleware fire!")

		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recover: %v\n%s", err, debug.Stack())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
