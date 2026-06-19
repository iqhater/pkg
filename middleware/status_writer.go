package middleware

import "net/http"

// StatusHTTP custom struct for show response status code in log middleware
type StatusHTTP struct {
	http.ResponseWriter
	StatusCode int
	written    bool // flag to check first call
}

// WriteHeader method wrties status code in http header
func (sr *StatusHTTP) WriteHeader(statusCode int) {

	// check if status code already set
	if !sr.written {
		sr.StatusCode = statusCode
		sr.ResponseWriter.WriteHeader(statusCode)
		sr.written = true
	}
}

// NewStatusHTTP function init new StatusHTTP
func NewStatusHTTP(w http.ResponseWriter) *StatusHTTP {
	return &StatusHTTP{w, http.StatusOK, false}
}
