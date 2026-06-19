package middleware

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Run a background goroutine to remove old entries from the visitors map.
func init() {
	timeout := 3 * time.Minute
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go cleanupVisitors(ctx, timeout)
}

// Create a custom visitor struct which holds the rate limiter for each
// visitor and the last time that the visitor was seen.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Change the the map to hold values of the type visitor.
var visitors = make(map[string]*visitor)
var mu sync.Mutex

func getVisitor(ip string, requestsPerSecond rate.Limit, maxRequests int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {

		// Allow 2 requests per second, with a maximum of 5 requests in a burst
		limiter := rate.NewLimiter(requestsPerSecond, maxRequests)

		// Include the current time when creating a new visitor.
		visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}

	// Update the last seen time for the visitor.
	v.lastSeen = time.Now()
	return v.limiter
}

// Every minute check the map for visitors that haven't been seen for
// more than 3 minutes and delete the entries.
func cleanupVisitors(ctx context.Context, timeout time.Duration) {

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}

		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > timeout {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// Limit middleware
// requestsPerSecond sets how many requests are allowed per second.
// maxRequests sets the maximum burst size.
func Limit(requestsPerSecond rate.Limit, maxRequests int) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			// log.Println("5. Limit middleware fire!")

			ip, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				log.Println(err.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			limiter := getVisitor(ip, requestsPerSecond, maxRequests)
			if !limiter.Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}
