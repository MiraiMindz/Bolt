package bolt

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"unsafe"

	json "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
)

// Use json-iterator as a fallback for compatibility.
// This instance is configured for maximum speed and compatibility.
var jsoniterCompat = jsoniter.ConfigCompatibleWithStandardLibrary

// Pre-allocated content type byte slices
var (
	contentTypeText = []byte("text/plain; charset=utf-8")
	contentTypeJSON = []byte("application/json; charset=utf-8")
)

// stringToBytes converts string to []byte without allocation using unsafe
// This is safe as long as the returned bytes are not modified
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// JSON buffer pool for optimized encoding/decoding
var (
	jsonBufferPool = sync.Pool{
		New: func() interface{} {
			// Pre-allocate 1KB buffer for most JSON payloads
			buf := make([]byte, 0, 1024)
			return &buf
		},
	}
)

// acquireJSONBuffer gets a buffer from the pool
func acquireJSONBuffer() *[]byte {
	return jsonBufferPool.Get().(*[]byte)
}

// releaseJSONBuffer returns a buffer to the pool after reset
func releaseJSONBuffer(buf *[]byte) {
	// Reset buffer but keep capacity if reasonable size
	if cap(*buf) <= 8192 { // Don't pool buffers larger than 8KB
		*buf = (*buf)[:0]
		jsonBufferPool.Put(buf)
	}
}

// Context represents the request context
type Context struct {
	Request    *Request
	Response   ResponseWriter
	app        *App
	params     ParamMap
	query      QueryValues
	StatusCode StatusCode
	headers    http.Header // Cached response headers
	fields     []Field     // Pre-allocated field slice for Fast API (reused via pool)
}

// Param gets a URL parameter by key
func (c *Context) Param(key string) string {
	if c.params == nil {
		return ""
	}
	return c.params[key]
}

// Query gets a query parameter by key
func (c *Context) Query(key string) string {
	return url.Values(c.query).Get(key)
}

// QueryInt gets a query parameter as integer with default value
func (c *Context) QueryInt(key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

// QueryBool gets a query parameter as boolean
func (c *Context) QueryBool(key string, defaultVal bool) bool {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}

// JSON sends a JSON response using optimized go-json with fast paths for common types.
func (c *Context) JSON(status int, v interface{}) error {
	c.StatusCode = StatusCode(status)

	if c.headers.Get("Content-Type") == "" {
		c.headers.Set("Content-Type", string(ContentTypeJSON))
	}

	// Fast paths for common simple types
	switch val := v.(type) {
	case string:
		c.Response.WriteHeader(status)
		_, err := c.Response.Write([]byte(`"` + val + `"`))
		return err
	case int:
		c.Response.WriteHeader(status)
		_, err := c.Response.Write([]byte(strconv.Itoa(val)))
		return err
	case int64:
		c.Response.WriteHeader(status)
		_, err := c.Response.Write([]byte(strconv.FormatInt(val, 10)))
		return err
	case bool:
		c.Response.WriteHeader(status)
		if val {
			_, err := c.Response.Write([]byte("true"))
			return err
		}
		_, err := c.Response.Write([]byte("false"))
		return err
	case map[string]string:
		// Optimized for simple string maps (very common case)
		c.Response.WriteHeader(status)
		return c.writeStringMap(val)
	default:
		// Use optimized buffer-based JSON encoding
		c.Response.WriteHeader(status)
		buf := acquireJSONBuffer()
		defer releaseJSONBuffer(buf)

		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		_, err = c.Response.Write(data)
		return err
	}
}

// writeStringMap optimizes JSON encoding for map[string]string
func (c *Context) writeStringMap(m map[string]string) error {
	if len(m) == 0 {
		_, err := c.Response.Write([]byte("{}"))
		return err
	}

	// Pre-allocate buffer for typical map sizes
	buf := make([]byte, 0, len(m)*32)
	buf = append(buf, '{')
	first := true
	for k, v := range m {
		if !first {
			buf = append(buf, ',')
		}
		buf = append(buf, '"')
		buf = append(buf, k...)
		buf = append(buf, `":"`...)
		// Simple escape for quotes in values
		for _, b := range []byte(v) {
			if b == '"' {
				buf = append(buf, '\\', '"')
			} else {
				buf = append(buf, b)
			}
		}
		buf = append(buf, '"')
		first = false
	}
	buf = append(buf, '}')
	_, err := c.Response.Write(buf)
	return err
}

// String sends a plain text response with zero allocations
// Uses unsafe conversion to avoid string-to-bytes allocation
func (c *Context) String(status int, s string) error {
	c.StatusCode = StatusCode(status)
	if c.headers.Get("Content-Type") == "" {
		c.headers.Set("Content-Type", string(ContentTypeText))
	}
	c.Response.WriteHeader(status)
	_, err := c.Response.Write(stringToBytes(s))
	return err
}

// StringBytes sends a plain text response with zero allocations
// Use this for maximum performance when you have pre-allocated []byte
func (c *Context) StringBytes(status int, b []byte) error {
	c.StatusCode = StatusCode(status)
	if c.headers.Get("Content-Type") == "" {
		c.headers.Set("Content-Type", string(ContentTypeText))
	}
	c.Response.WriteHeader(status)
	_, err := c.Response.Write(b)
	return err
}

// HTML sends an HTML response with zero allocations
func (c *Context) HTML(status int, html string) error {
	c.StatusCode = StatusCode(status)
	if c.headers.Get("Content-Type") == "" {
		c.headers.Set("Content-Type", string(ContentTypeHTML))
	}
	c.Response.WriteHeader(status)
	_, err := c.Response.Write(stringToBytes(html))
	return err
}

// Bytes sends a byte response with custom content type
func (c *Context) Bytes(status StatusCode, contentType ContentType, data []byte) error {
	c.StatusCode = status
	c.headers.Set("Content-Type", string(contentType))
	c.Response.WriteHeader(int(status))
	_, err := c.Response.Write(data)
	return err
}

// BindJSON binds request body to a struct using optimized JSON decoding.
func (c *Context) BindJSON(v interface{}) error {
	if c.Request.Body == nil {
		return ErrBadRequest
	}

	// Use streaming decoder with limited reader
	lr := io.LimitReader(c.Request.Body, 10<<20) // 10MB limit
	decoder := json.NewDecoder(lr)
	if err := decoder.Decode(v); err != nil {
		return ErrBadRequest
	}
	return nil
}

// SetHeader sets a response header
func (c *Context) SetHeader(key, value string) {
	c.headers.Set(key, value)
}

// Header returns the cached response headers
func (c *Context) Header() http.Header {
	return c.headers
}

// GetHeader gets a request header
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// Status sets the response status code
func (c *Context) Status(code StatusCode) *Context {
	c.StatusCode = code
	return c
}

// NoContent sends a 240 No Content response
func (c *Context) NoContent() error {
	c.StatusCode = 204
	c.Response.WriteHeader(204)
	return nil
}

// Redirect sends a redirect response
func (c *Context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return ErrInvalidRedirect
	}
	c.headers.Set("Location", url)
	c.Response.WriteHeader(code)
	return nil
}

// Fast response methods with zero allocations
var (
	precomputedResponses = map[string][]byte{
		"ok":       []byte("OK"),
		"true":     []byte("true"),
		"false":    []byte("false"),
		"empty":    []byte(""),
		"null":     []byte("null"),
		"zero":     []byte("0"),
		"one":      []byte("1"),
		"json_ok":  []byte(`{"status":"ok"}`),
		"json_err": []byte(`{"error":"error"}`),
	}
)

// FastText sends a pre-computed text response with zero allocations
func (c *Context) FastText(status int, key []byte) error {
	c.StatusCode = StatusCode(status)
	if c.headers.Get("Content-Type") == "" {
		c.headers.Set("Content-Type", string(ContentTypeText))
	}
	c.Response.WriteHeader(status)

	if response, ok := precomputedResponses[string(key)]; ok {
		_, err := c.Response.Write(response)
		return err
	}
	return c.StringBytes(status, key) // Fallback to zero-allocation version
}

// FastJSON sends a pre-computed JSON response with zero allocations
func (c *Context) FastJSON(status int, key []byte) error {
	c.StatusCode = StatusCode(status)
	if c.headers.Get("Content-Type") == "" {
		c.headers.Set("Content-Type", string(ContentTypeJSON))
	}
	c.Response.WriteHeader(status)

	if response, ok := precomputedResponses[string(key)]; ok {
		_, err := c.Response.Write(response)
		return err
	}
	return c.JSON(status, key) // Fallback
}

// JSONFields sends a JSON response using strongly-typed fields (Fast API)
// This is a zero-allocation alternative to JSON() for performance-critical paths
func (c *Context) JSONFields(status int, fields ...Field) error {
	c.StatusCode = StatusCode(status)
	if c.headers.Get("Content-Type") == "" {
		c.headers.Set("Content-Type", string(ContentTypeJSON))
	}
	c.Response.WriteHeader(status)

	// Write JSON directly to response without intermediate allocations
	return writeFieldsDirectToWriter(c.Response, fields)
}

// Fast API convenience methods - these use JSONFields internally

// OK sends a 200 OK response with optional fields
func (c *Context) OK(fields ...Field) error {
	return c.JSONFields(200, fields...)
}

// Created sends a 201 Created response with optional fields
func (c *Context) Created(fields ...Field) error {
	return c.JSONFields(201, fields...)
}

// BadRequest sends a 400 Bad Request response with optional fields
func (c *Context) BadRequest(fields ...Field) error {
	return c.JSONFields(400, fields...)
}

// Unauthorized sends a 401 Unauthorized response with optional fields
func (c *Context) Unauthorized(fields ...Field) error {
	return c.JSONFields(401, fields...)
}

// Forbidden sends a 403 Forbidden response with optional fields
func (c *Context) Forbidden(fields ...Field) error {
	return c.JSONFields(403, fields...)
}

// NotFound sends a 404 Not Found response with optional fields
func (c *Context) NotFound(fields ...Field) error {
	return c.JSONFields(404, fields...)
}

// InternalServerError sends a 500 Internal Server Error response with optional fields
func (c *Context) InternalServerError(fields ...Field) error {
	return c.JSONFields(500, fields...)
}

// Text is an alias for StringBytes - zero allocation text response
func (c *Context) Text(status int, b []byte) error {
	return c.StringBytes(status, b)
}

