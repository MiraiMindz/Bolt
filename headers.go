package bolt

import (
	"net/http"
	"strings"
	"sync"
)

// HeaderPool manages reusable HTTP headers to eliminate allocations
type HeaderPool struct {
	pool sync.Pool
}

// NewHeaderPool creates a new header pool
func NewHeaderPool() *HeaderPool {
	return &HeaderPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make(http.Header, 8) // Pre-allocate common size
			},
		},
	}
}

// Acquire gets a clean header from the pool
func (hp *HeaderPool) Acquire() http.Header {
	return hp.pool.Get().(http.Header)
}

// Release returns a header to the pool after cleaning it
func (hp *HeaderPool) Release(h http.Header) {
	// Clear all entries but keep underlying map
	for k := range h {
		delete(h, k)
	}
	hp.pool.Put(h)
}

// ZeroCopyResponseWriter wraps http.ResponseWriter to minimize header allocations
type ZeroCopyResponseWriter struct {
	http.ResponseWriter
	headerPool   *HeaderPool
	customHeader http.Header
	statusCode   int
	written      bool
}

// NewZeroCopyResponseWriter creates an optimized response writer
func NewZeroCopyResponseWriter(w http.ResponseWriter, headerPool *HeaderPool) *ZeroCopyResponseWriter {
	return &ZeroCopyResponseWriter{
		ResponseWriter: w,
		headerPool:     headerPool,
		customHeader:   headerPool.Acquire(),
		statusCode:     200,
	}
}

// Header returns the header map that will be sent by WriteHeader
func (zw *ZeroCopyResponseWriter) Header() http.Header {
	return zw.customHeader
}

// WriteHeader sends an HTTP response header with the provided status code
func (zw *ZeroCopyResponseWriter) WriteHeader(statusCode int) {
	if zw.written {
		return
	}

	zw.statusCode = statusCode
	zw.written = true

	// Only copy headers that were actually set
	originalHeader := zw.ResponseWriter.Header()
	for k, v := range zw.customHeader {
		originalHeader[k] = v
	}

	zw.ResponseWriter.WriteHeader(statusCode)
}

// Write writes the data to the connection as part of an HTTP reply
func (zw *ZeroCopyResponseWriter) Write(data []byte) (int, error) {
	if !zw.written {
		zw.WriteHeader(200)
	}
	return zw.ResponseWriter.Write(data)
}

// Release returns the header to the pool
func (zw *ZeroCopyResponseWriter) Release() {
	if zw.customHeader != nil {
		zw.headerPool.Release(zw.customHeader)
		zw.customHeader = nil
	}
}

// Common header strings that we can intern to reduce allocations
var (
	// Pre-interned common headers
	commonHeaders = map[string]string{
		"content-type":     "Content-Type",
		"content-length":   "Content-Length",
		"authorization":    "Authorization",
		"accept":           "Accept",
		"accept-encoding":  "Accept-Encoding",
		"accept-language":  "Accept-Language",
		"cache-control":    "Cache-Control",
		"connection":       "Connection",
		"cookie":           "Cookie",
		"host":             "Host",
		"referer":          "Referer",
		"user-agent":       "User-Agent",
		"x-forwarded-for":  "X-Forwarded-For",
		"x-forwarded-proto": "X-Forwarded-Proto",
		"x-real-ip":        "X-Real-Ip",
	}

	// Pre-interned common values
	commonValues = map[string]string{
		"application/json":           "application/json",
		"application/json; charset=utf-8": "application/json; charset=utf-8",
		"text/plain":                 "text/plain",
		"text/plain; charset=utf-8":  "text/plain; charset=utf-8",
		"text/html":                  "text/html",
		"text/html; charset=utf-8":   "text/html; charset=utf-8",
		"gzip":                       "gzip",
		"deflate":                    "deflate",
		"close":                      "close",
		"keep-alive":                 "keep-alive",
		"no-cache":                   "no-cache",
		"max-age=0":                  "max-age=0",
	}
)

// InternHeaderName returns an interned version of a header name
func InternHeaderName(name string) string {
	lower := strings.ToLower(name)
	if interned, ok := commonHeaders[lower]; ok {
		return interned
	}

	// Use http.CanonicalHeaderKey for consistency
	return http.CanonicalHeaderKey(name)
}

// InternHeaderValue returns an interned version of a header value if it's common
func InternHeaderValue(value string) string {
	if interned, ok := commonValues[value]; ok {
		return interned
	}
	return value
}

// FastHeader represents an optimized header implementation
type FastHeader struct {
	data map[string][]string
	pool *sync.Pool
}

// NewFastHeader creates an optimized header
func NewFastHeader() *FastHeader {
	return &FastHeader{
		data: make(map[string][]string, 8),
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]string, 0, 4)
			},
		},
	}
}

// Set sets the header entries associated with key to the single element value
func (fh *FastHeader) Set(key, value string) {
	key = InternHeaderName(key)
	value = InternHeaderValue(value)

	// Release old slice back to pool if it exists
	if existing, ok := fh.data[key]; ok && len(existing) > 0 {
		for i := range existing {
			existing[i] = ""
		}
		existing = existing[:0]
		if cap(existing) <= 8 { // Only pool small slices
			fh.pool.Put(existing)
		}
	}

	// Get new slice from pool
	slice := fh.pool.Get().([]string)
	slice = append(slice, value)
	fh.data[key] = slice
}

// Get gets the first value associated with the given key
func (fh *FastHeader) Get(key string) string {
	key = InternHeaderName(key)
	if values, ok := fh.data[key]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

// Add adds the key, value pair to the header
func (fh *FastHeader) Add(key, value string) {
	key = InternHeaderName(key)
	value = InternHeaderValue(value)

	if existing, ok := fh.data[key]; ok {
		fh.data[key] = append(existing, value)
	} else {
		slice := fh.pool.Get().([]string)
		slice = append(slice, value)
		fh.data[key] = slice
	}
}

// Reset clears the header for reuse
func (fh *FastHeader) Reset() {
	for key, values := range fh.data {
		// Clear strings and return slice to pool
		for i := range values {
			values[i] = ""
		}
		values = values[:0]
		if cap(values) <= 8 {
			fh.pool.Put(values)
		}
		delete(fh.data, key)
	}
}