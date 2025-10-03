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

// ContextPools manages different types of context objects for optimization
type ContextPools struct {
	staticPool   sync.Pool // For static routes (minimal context)
	dynamicPool  sync.Pool // For dynamic routes (full context)
	queryPool    sync.Pool // For routes that use query params
}

// ContextPool manages Context object reuse (backwards compatibility)
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
	c.headers = nil

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

// SmartBufferPool manages different sized buffers for optimal performance
type SmartBufferPool struct {
	smallPool  sync.Pool // < 1KB
	mediumPool sync.Pool // 1KB - 8KB
	largePool  sync.Pool // > 8KB
}

// BufferPool manages byte buffer reuse (backwards compatibility)
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

// NewSmartBufferPool creates a new smart buffer pool with size-aware pooling
func NewSmartBufferPool() *SmartBufferPool {
	return &SmartBufferPool{
		smallPool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, 512) // 512 bytes
				return bytes.NewBuffer(buf)
			},
		},
		mediumPool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, 4096) // 4KB
				return bytes.NewBuffer(buf)
			},
		},
		largePool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, 16384) // 16KB
				return bytes.NewBuffer(buf)
			},
		},
	}
}

// Acquire gets a buffer sized appropriately for the expected data
func (p *SmartBufferPool) Acquire(expectedSize int) *bytes.Buffer {
	switch {
	case expectedSize < 1024:
		return p.smallPool.Get().(*bytes.Buffer)
	case expectedSize < 8192:
		return p.mediumPool.Get().(*bytes.Buffer)
	default:
		return p.largePool.Get().(*bytes.Buffer)
	}
}

// Release returns a buffer to the appropriate pool
func (p *SmartBufferPool) Release(buf *bytes.Buffer, size int) {
	buf.Reset()
	switch {
	case size < 1024:
		p.smallPool.Put(buf)
	case size < 8192:
		p.mediumPool.Put(buf)
	default:
		p.largePool.Put(buf)
	}
}

// NewContextPools creates optimized context pools for different route types
func NewContextPools() *ContextPools {
	return &ContextPools{
		staticPool: sync.Pool{
			New: func() interface{} {
				// Minimal context for static routes
				return &Context{}
			},
		},
		dynamicPool: sync.Pool{
			New: func() interface{} {
				// Full context for dynamic routes
				return &Context{
					params: make(ParamMap, DefaultParamsSize),
					query:  make(QueryValues, DefaultParamsSize),
				}
			},
		},
		queryPool: sync.Pool{
			New: func() interface{} {
				// Context optimized for query-heavy routes
				return &Context{
					params: make(ParamMap, 2), // Minimal params
					query:  make(QueryValues, 16), // Larger query map
				}
			},
		},
	}
}

// AcquireStatic gets a minimal context for static routes
func (p *ContextPools) AcquireStatic() *Context {
	return p.staticPool.Get().(*Context)
}

// AcquireDynamic gets a full context for dynamic routes
func (p *ContextPools) AcquireDynamic() *Context {
	return p.dynamicPool.Get().(*Context)
}

// AcquireQuery gets a context optimized for query-heavy routes
func (p *ContextPools) AcquireQuery() *Context {
	return p.queryPool.Get().(*Context)
}

// Release returns a context to the appropriate pool
func (p *ContextPools) Release(c *Context, poolType string) {
	// Reset context
	c.Request = nil
	c.Response = nil
	c.app = nil
	c.StatusCode = 0
	c.params = nil

	switch poolType {
	case "static":
		p.staticPool.Put(c)
	case "dynamic":
		// Clear maps if they exist
		if c.query != nil {
			if len(c.query) > MaxParams {
				c.query = make(QueryValues, DefaultParamsSize)
			} else {
				for k := range c.query {
					delete(c.query, k)
				}
			}
		}
		p.dynamicPool.Put(c)
	case "query":
		// Clear query map
		if c.query != nil {
			if len(c.query) > 32 { // Higher threshold for query pool
				c.query = make(QueryValues, 16)
			} else {
				for k := range c.query {
					delete(c.query, k)
				}
			}
		}
		p.queryPool.Put(c)
	}
}
