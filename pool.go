package bolt

import (
	"bytes"
	"sync"
)

// ContextPool manages Context object reuse
type ContextPool struct {
	pool sync.Pool
}

// NewContextPool creates a new context pool
func NewContextPool() *ContextPool {
	return &ContextPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Context{
					params: make(ParamMap, 8),
					query:  make(QueryValues, 8), // Pre-allocate query map
				}
			},
		},
	}
}

// Acquire gets a Context from the pool
func (p *ContextPool) Acquire() *Context {
	return p.pool.Get().(*Context)
}

// Release returns a Context to the pool
func (p *ContextPool) Release(c *Context) {
	c.Request = nil
	c.Response = nil
	c.app = nil
	c.StatusCode = 0

	// Clear params map for reuse
	for k := range c.params {
		delete(c.params, k)
	}

	// Clear query map for reuse
	for k := range c.query {
		delete(c.query, k)
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

