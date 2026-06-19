package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	"bytes"
)

// helper function
func parse(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}

// helper function
func assertContentEquals(t *testing.T, content, expected []byte) {
	if !bytes.Equal(content, expected) {
		t.Errorf("content should '%s', but was '%s'", expected, content)
	}
}

func TestGetEmpty(t *testing.T) {

	// Arrange
	storage := NewCache("5s")

	// Act
	content, contentType := storage.Get("MY_KEY")

	// Assert
	assertContentEquals(t, content, []byte(""))
	assertContentEquals(t, []byte(contentType), []byte(""))
}

func TestGetValue(t *testing.T) {

	// Arrange
	storage := NewCache("5s")
	storage.Set("MY_KEY", []byte("123456"), "application/json", parse("10ms"))

	// Act
	content, contentType := storage.Get("MY_KEY")

	// Assert
	assertContentEquals(t, content, []byte("123456"))
	assertContentEquals(t, []byte(contentType), []byte("application/json"))
}

func TestGetExpiredValue(t *testing.T) {

	// Arrange
	storage := NewCache("5s")
	storage.Set("MY_KEY", []byte("123456"), "", parse("10ms"))
	time.Sleep(parse("30ms"))

	// Act
	content, _ := storage.Get("MY_KEY")

	// Assert
	assertContentEquals(t, content, []byte(""))
}

func TestGetValueAfterSet(t *testing.T) {

	// Arrange
	storage := NewCache("5s")
	storage.Set("MY_KEY", []byte("123456"), "application/json", parse("10ms"))

	// Act
	content, contentType := storage.Get("MY_KEY")

	// Assert
	assertContentEquals(t, content, []byte("123456"))
	assertContentEquals(t, []byte(contentType), []byte("application/json"))
}

func TestNewCacheInvalidDuration(t *testing.T) {

	// Arrange
	duration := "invalid"

	// Act
	cache := NewCache(duration)

	// Assert
	if cache.duration <= 0 {
		t.Errorf("NewCache should setup default valid duration %s\n", duration)
	}
}

func TestCacheResponse(t *testing.T) {
	tests := []struct {
		name                  string
		cacheKey              string
		duration              string
		prepopulate           func(*Cache)
		sleepAfterPrepopulate time.Duration
		nextHandler           http.Handler
		wantStatus            int
		wantBody              string
		wantCacheHeader       string
		wantContentType       string
		wantCached            bool
		wantCachedBody        string
		wantCachedContentType string
		wantHandlerCalled     bool
	}{
		{
			name:     "cache miss caches successful response",
			cacheKey: "/success",
			duration: "1m",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("fresh content"))
			}),
			wantStatus:            http.StatusOK,
			wantBody:              "fresh content",
			wantCacheHeader:       "MISS",
			wantContentType:       "application/json",
			wantCached:            true,
			wantCachedBody:        "fresh content",
			wantCachedContentType: "application/json",
			wantHandlerCalled:     true,
		},
		{
			name:     "cache hit returns cached response",
			cacheKey: "/hit",
			duration: "1m",
			prepopulate: func(cache *Cache) {
				cache.Set("/hit", []byte("cached content"), "application/json", cache.duration)
			},
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("fresh content"))
			}),
			wantStatus:            http.StatusOK,
			wantBody:              "cached content",
			wantCacheHeader:       "HIT",
			wantContentType:       "application/json",
			wantCached:            true,
			wantCachedBody:        "cached content",
			wantCachedContentType: "application/json",
			wantHandlerCalled:     false,
		},
		{
			name:     "non ok status is not cached",
			cacheKey: "/created",
			duration: "1m",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("created content"))
			}),
			wantStatus:        http.StatusCreated,
			wantBody:          "created content",
			wantCacheHeader:   "MISS",
			wantContentType:   "text/plain",
			wantCached:        false,
			wantHandlerCalled: true,
		},
		{
			name:     "expired cache entry is refreshed",
			cacheKey: "/expired",
			duration: "1m",
			prepopulate: func(cache *Cache) {
				cache.Set("/expired", []byte("expired content"), "application/json", 10*time.Millisecond)
			},
			sleepAfterPrepopulate: 30 * time.Millisecond,
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("fresh content"))
			}),
			wantStatus:            http.StatusOK,
			wantBody:              "fresh content",
			wantCacheHeader:       "MISS",
			wantContentType:       "application/json",
			wantCached:            true,
			wantCachedBody:        "fresh content",
			wantCachedContentType: "application/json",
			wantHandlerCalled:     true,
		},
		{
			name:     "invalid duration uses default cache duration",
			cacheKey: "/invalid",
			duration: "invalid",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("content"))
			}),
			wantStatus:            http.StatusOK,
			wantBody:              "content",
			wantCacheHeader:       "MISS",
			wantContentType:       "application/json",
			wantCached:            true,
			wantCachedBody:        "content",
			wantCachedContentType: "application/json",
			wantHandlerCalled:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Arrange
			cache := NewCache(tt.duration)
			called := false

			if tt.prepopulate != nil {
				tt.prepopulate(cache)
			}

			if tt.sleepAfterPrepopulate > 0 {
				time.Sleep(tt.sleepAfterPrepopulate)
			}

			handler := cache.CacheResponse(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				tt.nextHandler.ServeHTTP(w, r)
			}))

			req := httptest.NewRequest(http.MethodGet, tt.cacheKey, nil)
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if body := rec.Body.String(); body != tt.wantBody {
				t.Errorf("expected body %q, got %q", tt.wantBody, body)
			}

			if got := rec.Header().Get("X-Cache"); got != tt.wantCacheHeader {
				t.Errorf("expected X-Cache header %q, got %q", tt.wantCacheHeader, got)
			}

			if got := rec.Header().Get("Cache-Control"); got == "" {
				t.Error("expected Cache-Control header to be set")
			}

			if got := rec.Header().Get("Content-Type"); got != tt.wantContentType {
				t.Errorf("expected Content-Type header %q, got %q", tt.wantContentType, got)
			}

			if called != tt.wantHandlerCalled {
				t.Errorf("expected handler called %v, got %v", tt.wantHandlerCalled, called)
			}

			if tt.wantCached {
				content, contentType := cache.Get(tt.cacheKey)

				if !bytes.Equal(content, []byte(tt.wantCachedBody)) {
					t.Errorf("expected cached body %q, got %q", tt.wantCachedBody, string(content))
				}

				if contentType != tt.wantCachedContentType {
					t.Errorf("expected cached content type %q, got %q", tt.wantCachedContentType, contentType)
				}
			} else {
				content, _ := cache.Get(tt.cacheKey)

				if content != nil {
					t.Errorf("expected content not to be cached, got %q", string(content))
				}
			}
		})
	}
}

func TestCacheResponseHit(t *testing.T) {

	// Arrange
	cache := NewCache("1m")
	uri := "/test-endpoint"
	content := []byte("cached content")
	contentType := "application/json"

	cache.Set(uri, content, contentType, time.Minute)

	// create handler
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fresh content"))
	})

	req := httptest.NewRequest("GET", uri, nil)
	w := httptest.NewRecorder()

	// Act
	cache.CacheResponse(next).ServeHTTP(w, req)

	// Assert
	if called {
		t.Error("Next handler was called, but it should have been skipped on cache HIT")
	}

	if w.Header().Get("X-Cache") != "HIT" {
		t.Errorf("Expected X-Cache header 'HIT', got %s", w.Header().Get("X-Cache"))
	}

	if w.Body.String() != string(content) {
		t.Errorf("Expected body %s, got %s", string(content), w.Body.String())
	}
}

// errorResponseWriter is a custom ResponseWriter that simulates a write error after a certain number of writes.
type errorResponseWriter struct {
	*httptest.ResponseRecorder
	writeError bool
	failAfter  int
}

func (w *errorResponseWriter) Write(b []byte) (int, error) {
	if w.failAfter <= 0 {
		w.writeError = true
		return 0, errors.New("simulated write error")
	}
	w.failAfter--
	return w.ResponseRecorder.Write(b)
}

func TestCacheResponseWriteError(t *testing.T) {

	// Arrange
	cache := NewCache("10s")
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test content"))
	})
	handler := cache.CacheResponse(nextHandler)

	req := httptest.NewRequest("GET", "http://example.com/error-test", nil)
	rec := &errorResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
		failAfter:        0,
	}

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if !rec.writeError {
		t.Error("expected write error, but none occurred")
	}
}
