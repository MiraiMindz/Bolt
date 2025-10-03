package bolt

import (
	"bytes"
	"sync"
)

const (
	// MaxParams is the threshold for resetting the params map.
	// If a request has more than this many params, we create a new map
	// instead of clearing the old one to avoid memory bloat.
	MaxParams = 8
	// DefaultParamsSize is the initial size for new/reset params maps.
	DefaultParamsSize = 4
)

// ContextPool manages Context object reuse
type ContextPool struct {
	pool sync.Pool
}

// NewContextPool creates a new context pool with optimized initial sizes.
func NewContextPool() *ContextPool {
	return &ContextPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Context{
					// Most routes have fewer than 4 params or query values.
					params: make(ParamMap, DefaultParamsSize),
					query:  make(QueryValues, DefaultParamsSize),
				}
			},
		},
	}
}

// Acquire gets a Context from the pool
func (p *ContextPool) Acquire() *Context {
	return p.pool.Get().(*Context)
}

// Release returns a Context to the pool with a more aggressive clearing strategy.
func (p *ContextPool) Release(c *Context) {
	// Reset basic fields
	c.Request = nil
	c.Response = nil
	c.app = nil
	c.StatusCode = 0
	c.params = nil // The router is responsible for pooling params

	// If the query map grew too large, create a new one to prevent memory bloat.
	// Otherwise, just clear the existing one.
	if len(c.query) > MaxParams {
		c.query = make(QueryValues, DefaultParamsSize)
	} else {
		for k := range c.query {
			delete(c.query, k)
		}
	}

	p.pool.Put(c)
}

// BufferPool manages byte buffer reuse
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool creates a new buffer pool that pools *bytes.Buffer objects
func NewBufferPool() *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

// Acquire gets a buffer from the pool
func (p *BufferPool) Acquire() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

// Release returns a buffer to the pool after resetting it
func (p *BufferPool) Release(buf *bytes.Buffer) {
	buf.Reset()
	p.pool.Put(buf)
}
