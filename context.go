package bolt

import (
	"encoding/json"
	"io"
	"net/url"
	"strconv"
)

// Context represents the request context
type Context struct {
	Request    *Request
	Response   ResponseWriter
	app        *App
	params     ParamMap
	query      QueryValues
	StatusCode StatusCode
}

// Param gets a URL parameter by key
func (c *Context) Param(key string) string {
	return c.params[key]
}

// Query gets a query parameter by key
func (c *Context) Query(key string) string {
	// The query map is now pre-filled in ServeHTTP, so no need for lazy init
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

// JSON sends a JSON response
func (c *Context) JSON(status int, v interface{}) error {
	c.StatusCode = StatusCode(status)
	c.Response.Header().Set("Content-Type", string(ContentTypeJSON))
	c.Response.WriteHeader(status)
	return json.NewEncoder(c.Response).Encode(v)
}

// String sends a plain text response
func (c *Context) String(status int, s string) error {
	c.StatusCode = StatusCode(status)
	c.Response.Header().Set("Content-Type", string(ContentTypeText))
	c.Response.WriteHeader(status)
	_, err := c.Response.Write([]byte(s))
	return err
}

// HTML sends an HTML response
func (c *Context) HTML(status int, html string) error {
	c.StatusCode = StatusCode(status)
	c.Response.Header().Set("Content-Type", string(ContentTypeHTML))
	c.Response.WriteHeader(status)
	_, err := c.Response.Write([]byte(html))
	return err
}

// Bytes sends a byte response with custom content type
func (c *Context) Bytes(status int, contentType ContentType, data []byte) error {
	c.StatusCode = StatusCode(status)
	c.Response.Header().Set("Content-Type", string(contentType))
	c.Response.WriteHeader(status)
	_, err := c.Response.Write(data)
	return err
}

// BindJSON binds request body to a struct by reading the body into a pooled buffer
// and then unmarshaling. This reduces allocations.
func (c *Context) BindJSON(v interface{}) error {
	if c.Request.Body == nil {
		return ErrBadRequest
	}

	// Use buffer pool if available for efficiency
	if c.app != nil && c.app.bufferPool != nil {
		buf := c.app.bufferPool.Acquire()
		defer c.app.bufferPool.Release(buf)

		lr := io.LimitReader(c.Request.Body, 10<<20) // 10MB limit

		_, err := io.Copy(buf, lr)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(buf.Bytes(), v); err != nil {
			return ErrBadRequest
		}
		return nil
	}

	// Fallback for when pooling is disabled
	lr := io.LimitReader(c.Request.Body, 10<<20)
	decoder := json.NewDecoder(lr)
	if err := decoder.Decode(v); err != nil {
		return ErrBadRequest
	}
	return nil
}

// SetHeader sets a response header
func (c *Context) SetHeader(key, value string) {
	c.Response.Header().Set(key, value)
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

// NoContent sends a 204 No Content response
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
	c.Response.Header().Set("Location", url)
	c.Response.WriteHeader(code)
	return nil
}

