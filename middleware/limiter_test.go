package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// helper function to reset the visitors map before each test.
func cleanupVisitorsOnce(timeout time.Duration) {

	mu.Lock()
	defer mu.Unlock()

	for ip, v := range visitors {
		if time.Since(v.lastSeen) > timeout {
			delete(visitors, ip)
		}
	}
}

func TestGetVisitor(t *testing.T) {

	// Reset the global visitors map for a clean test environment.
	visitors = make(map[string]*visitor)

	ip := "192.168.0.1"
	requestsPerSecond := rate.Limit(2)
	maxRequests := 5

	limiter1 := getVisitor(ip, requestsPerSecond, maxRequests)
	if limiter1 == nil {
		t.Fatal("getVisitor returned nil limiter")
	}

	// Save the lastSeen time after the first call.
	mu.Lock()
	firstSeen := visitors[ip].lastSeen
	mu.Unlock()

	// Wait a bit to ensure that the lastSeen time would be different if it were updated.
	time.Sleep(10 * time.Millisecond)
	limiter2 := getVisitor(ip, requestsPerSecond, maxRequests)
	if limiter2 != limiter1 {
		t.Fatal("getVisitor returned a different limiter for the same IP")
	}

	mu.Lock()
	secondSeen := visitors[ip].lastSeen
	mu.Unlock()

	if !secondSeen.After(firstSeen) {
		t.Fatal("getVisitor did not update lastSeen time on subsequent call")
	}
}

func TestMultipleIPs(t *testing.T) {

	// Reset the global visitors map for a clean test environment.
	visitors = make(map[string]*visitor)

	ip1 := "192.168.0.1"
	ip2 := "10.0.0.2"
	requestsPerSecond := rate.Limit(2)
	maxRequests := 5

	limiter1 := getVisitor(ip1, requestsPerSecond, maxRequests)
	limiter2 := getVisitor(ip2, requestsPerSecond, maxRequests)

	if limiter1 == nil || limiter2 == nil {
		t.Fatal("getVisitor returned nil limiter for one of the IPs")
	}

	if limiter1 == limiter2 {
		t.Fatal("getVisitor returned the same limiter for different IPs")
	}

	// Check that both IPs are stored in the visitors map.
	if len(visitors) != 2 {
		t.Fatalf("Wanted to find 2 entries in visitors map, found %d", len(visitors))
	}
}

func TestLimitMiddleware(t *testing.T) {

	// Reset the global visitors map for a clean test environment.
	visitors = make(map[string]*visitor)

	// Test handler that simply returns 200 OK.
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	requestsPerSecond := rate.Limit(2)
	maxRequests := 5

	limitedHandler := Limit(requestsPerSecond, maxRequests)(testHandler)

	// Request with valid RemoteAddr.
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = net.JoinHostPort("127.0.0.1", "12345")
	rr := httptest.NewRecorder()

	// execute the request and check for 200 OK.
	limitedHandler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Want: status %d, get: %d", http.StatusOK, rr.Code)
	}

	// Check that the limiter is working by sending multiple requests in quick succession.
	// Since the limiter allows 5 requests per second, the 6th request should be blocked.
	var lastCode int
	requests := 7

	for range requests {
		limitedHandler.ServeHTTP(rr, req)
		lastCode = rr.Code
		rr = httptest.NewRecorder()
	}

	if lastCode != http.StatusTooManyRequests {
		t.Fatalf("Want: status %d over limit, get: %d", http.StatusTooManyRequests, lastCode)
	}
}

func TestLimitMiddlewareInvalidRemoteAddr(t *testing.T) {

	// Test branch where RemoteAddr is not a valid IP:port format.
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	requestsPerSecond := rate.Limit(2)
	maxRequests := 5

	limitedHandler := Limit(requestsPerSecond, maxRequests)(testHandler)

	// Create a request with an invalid RemoteAddr.
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "invalid_remote_addr"
	rr := httptest.NewRecorder()

	limitedHandler.ServeHTTP(rr, req)

	// Wait 500 error
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("Want: status %d for invalid RemoteAddr, get: %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestCleanupVisitors(t *testing.T) {

	// reset the global visitors map.
	visitors = make(map[string]*visitor)

	ip := "10.0.0.1"

	// Add a visitor with lastSeen time set to 2 minutes ago.
	mu.Lock()
	visitors[ip] = &visitor{
		limiter:  rate.NewLimiter(rate.Limit(1), 1),
		lastSeen: time.Now().Add(-2 * time.Minute),
	}
	mu.Unlock()

	// call the cleanup function with a timeout of 150 milliseconds.
	cleanupVisitorsOnce(150 * time.Millisecond)

	mu.Lock()
	_, exists := visitors[ip]
	mu.Unlock()

	if exists {
		t.Fatalf("Want visitor with IP %s was deleted", ip)
	}
}
