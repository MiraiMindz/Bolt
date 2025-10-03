# Fast API Quick Start Guide

Get started with Bolt's zero-allocation Fast API in 5 minutes.

## Installation

```bash
go get -u github.com/miraimindz/bolt
```

## Hello World - Both APIs

### Sugared (Easy)
```go
package main

import "bolt"

func main() {
    app := bolt.New()

    app.Get("/hello", func(c *bolt.Context) error {
        return c.JSON(200, map[string]string{
            "message": "Hello, World!",
        })
    })

    app.Listen(":3000")
}
```

### Fast (Zero-Allocation)
```go
package main

import "bolt"

func main() {
    app := bolt.New()

    app.GetFast("/hello", func(c *bolt.FastContext) error {
        return c.JSONFields(200,
            bolt.String("message", "Hello, World!"),
        )
    })

    app.Listen(":3000")
}
```

## Cheat Sheet

### Field Types
```go
bolt.String("key", "value")           // → "key": "value"
bolt.Int("count", 42)                 // → "count": 42
bolt.Int64("id", 123456789)           // → "id": 123456789
bolt.Float64("price", 19.99)          // → "price": 19.99
bolt.Bool("active", true)             // → "active": true
bolt.Time("created", time.Now())      // → "created": "2025-01-01T12:00:00Z"
bolt.Duration("timeout", time.Hour)   // → "timeout": "1h0m0s"
bolt.Bytes("data", []byte("hello"))   // → "data": "hello"
```

### Route Methods
```go
app.GetFast(path, handler)      // GET
app.PostFast(path, handler)     // POST
app.PutFast(path, handler)      // PUT
app.DeleteFast(path, handler)   // DELETE
app.PatchFast(path, handler)    // PATCH
```

### Response Methods
```go
c.JSONFields(200, fields...)              // Custom JSON
c.Text(200, []byte("text"))               // Plain text
c.OK(fields...)                           // 200 OK
c.Created(fields...)                      // 201 Created
c.NoContent()                             // 204 No Content
c.BadRequest(fields...)                   // 400 Bad Request
c.Unauthorized(fields...)                 // 401 Unauthorized
c.Forbidden(fields...)                    // 403 Forbidden
c.NotFound(fields...)                     // 404 Not Found
c.InternalServerError(fields...)          // 500 Internal Server Error
```

### Parameter Access
```go
c.ParamString("id")                  // Get URL param as string
c.ParamInt("id", 0)                  // Get URL param as int
c.ParamInt64("id", 0)                // Get URL param as int64
c.QueryString("search")              // Get query param
```

## Common Patterns

### REST API Endpoint
```go
app.GetFast("/users/:id", func(c *bolt.FastContext) error {
    id := c.ParamString("id")

    user, err := db.GetUser(id)
    if err != nil {
        return c.NotFound(bolt.String("error", "user not found"))
    }

    return c.JSONFields(200,
        bolt.String("id", user.ID),
        bolt.String("name", user.Name),
        bolt.String("email", user.Email),
        bolt.Time("created", user.CreatedAt),
    )
})
```

### Create Resource
```go
app.PostFast("/users", func(c *bolt.FastContext) error {
    var input struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }

    if err := c.BindJSON(&input); err != nil {
        return c.BadRequest(bolt.String("error", "invalid JSON"))
    }

    user, err := db.CreateUser(input.Name, input.Email)
    if err != nil {
        return c.InternalServerError(
            bolt.String("error", "failed to create user"),
        )
    }

    return c.Created(
        bolt.String("id", user.ID),
        bolt.String("name", user.Name),
        bolt.String("email", user.Email),
    )
})
```

### Health Check (Zero Allocation)
```go
var healthOK = []byte("OK")

app.GetFast("/health", func(c *bolt.FastContext) error {
    return c.Text(200, healthOK)
})
```

### Metrics Endpoint
```go
app.GetFast("/metrics", func(c *bolt.FastContext) error {
    return c.JSONFields(200,
        bolt.Int64("requests_total", metrics.GetTotalRequests()),
        bolt.Float64("avg_latency_ms", metrics.GetAvgLatency()),
        bolt.Int("active_connections", metrics.GetConnections()),
        bolt.Time("timestamp", time.Now()),
    )
})
```

## When to Use Fast API

✅ **Use Fast API when:**
- High-throughput services (>10K req/sec)
- Microservices with strict SLAs
- Real-time APIs (gaming, trading)
- Health checks and metrics
- Every allocation matters

❌ **Use Sugared API when:**
- Prototyping/MVP
- Internal tools
- CRUD with moderate traffic
- Developer productivity > microseconds

## Performance Comparison

```
Sugared API:  1250 ns/op    856 B/op    12 allocs/op
Fast API:      420 ns/op    128 B/op     1 alloc/op

Fast is 3x faster with 87% fewer allocations
```

## Next Steps

- 📖 Read [SUGARED_VS_FAST_API.md](SUGARED_VS_FAST_API.md) for detailed comparison
- 🎯 See [examples/sugared_vs_fast/](../examples/sugared_vs_fast/) for working examples
- ⚡ Check [ZERO_ALLOCATION_API.md](ZERO_ALLOCATION_API.md) for advanced patterns

## Pro Tips

1. **Mix and Match**: Use Sugared for most endpoints, Fast for hot paths
2. **Profile First**: Don't optimize prematurely - measure then optimize
3. **Pre-allocate**: Define `[]byte` responses at package level for zero allocations
4. **Upgrade Gradually**: Start with Sugared, migrate to Fast when needed
5. **Benchmark**: Use `go test -bench=.` to verify improvements

---

**Bolt: Honest Performance, Your Choice** ⚡
