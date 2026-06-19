package middleware

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Storage interface for cache middleware implementation
type Storage interface {

	// Get retrieves the cached content for the given key.
	Get(key string) ([]byte, string)

	// Set stores the content in the cache with the specified key and duration.
	Set(key string, content []byte, contentType string, duration time.Duration)
}

// static interface implementation check for convinience
var _ Storage = (*Cache)(nil)

// Item is a cached reference
type Item struct {
	Content     []byte
	ContentType string
	Expiration  int64
}

// Cache struct for caching strings in memory
type Cache struct {
	items    map[string]Item
	mu       sync.RWMutex
	duration time.Duration
}

// Expired returns true if the item has expired.
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// Get a cached content by key
func (c *Cache) Get(key string) ([]byte, string) {

	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, ""
	}

	if item.Expired() {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, ""
	}
	return item.Content, item.ContentType
}

// Set a cached content by key
func (c *Cache) Set(key string, content []byte, contentType string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item{
		Content:     content,
		ContentType: contentType,
		Expiration:  time.Now().Add(duration).UnixNano(),
	}
}

// NewCache creates a new in memory Cache
func NewCache(cacheDuration string) *Cache {

	duration, err := time.ParseDuration(cacheDuration)
	if err != nil || duration <= 0 {
		log.Println("Failed to parse cache duration. Will be set default cache duration.", err)

		// set default cache duration to 5 minutes if parsing fails
		duration = time.Minute * 5
	}
	return &Cache{
		items:    make(map[string]Item),
		mu:       sync.RWMutex{},
		duration: duration,
	}
}

// cachingResponseWriter is a wrapper around http.ResponseWriter that allows us to cache the response body.
type cachingResponseWriter struct {
	*StatusHTTP
	body *bytes.Buffer
}

func (cw *cachingResponseWriter) Write(b []byte) (int, error) {

	// save into buffer
	cw.body.Write(b)
	return cw.ResponseWriter.Write(b)
}

// CacheResponse middleware
func (c *Cache) CacheResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// log.Println("9. CacheResponse middleware fire!")

		w.Header().Set("Cache-Control", "private, max-age="+fmt.Sprintf("%.0f", c.duration.Seconds()))
		// w.Header().Set("Cache-Control", "no-store")

		content, contentType := c.Get(req.RequestURI)
		if content != nil {
			w.Header().Set("X-Cache", "HIT")

			// restore content type
			if contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}

			_, err := w.Write(content)
			if err != nil {
				log.Println("Cache write error! ", err)
			}
			return
		}
		w.Header().Set("X-Cache", "MISS")

		// init wrapper for caching response body
		cw := &cachingResponseWriter{
			StatusHTTP: NewStatusHTTP(w),
			body:       new(bytes.Buffer),
		}

		next.ServeHTTP(cw, req)

		// cache if status code is OK
		if cw.StatusCode == http.StatusOK {
			contentType := cw.Header().Get("Content-Type")
			c.Set(req.RequestURI, cw.body.Bytes(), contentType, c.duration)
		}
	})
}
