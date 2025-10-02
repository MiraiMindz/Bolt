package bolt

import "sync"

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
      
    for k := range c.params {  
        delete(c.params, k)  
    }  
      
    c.query = nil  
      
    p.pool.Put(c)  
}

// BufferPool manages byte buffer reuse  
type BufferPool struct {  
    pool sync.Pool  
}

// NewBufferPool creates a new buffer pool  
func NewBufferPool(size int) *BufferPool {  
    return &BufferPool{  
        pool: sync.Pool{  
            New: func() interface{} {  
                buf := make([]byte, 0, size)  
                return &buf  
            },  
        },  
    }  
}

// Acquire gets a buffer from the pool  
func (p *BufferPool) Acquire() *[]byte {  
    return p.pool.Get().(*[]byte)  
}

// Release returns a buffer to the pool  
func (p *BufferPool) Release(buf *[]byte) {  
    *buf = (*buf)[:0]  
    p.pool.Put(buf)  
}