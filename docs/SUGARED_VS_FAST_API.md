## Sugared vs Fast API - Choose Your Performance Level

Inspired by Uber's Zap logger, Bolt provides **two complementary APIs** that let you choose between ergonomics and performance based on your needs.

## The Two APIs

### 🍬 **Sugared API** - Ergonomic & Convenient
**Use when:** Developer productivity > microsecond optimizations

- Clean, intuitive syntax
- Accepts native Go types
- Small allocation overhead
- Perfect for 90% of use cases

```go
app.Get("/users/:id", func(c *bolt.Context) error {
    return c.JSON(200, map[string]interface{}{
        "id":    c.Param("id"),
        "name":  "John Doe",
        "email": "john@example.com",
    })
})
```

### ⚡ **Fast API** - Zero-Allocation & Strongly-Typed
**Use when:** Maximum performance is critical

- Zero-allocation field building
- No reflection overhead
- Strongly-typed values
- Perfect for hot paths & high-throughput services

```go
app.GetFast("/users/:id", func(c *bolt.FastContext) error {
    return c.JSONFields(200,
        bolt.String("id", c.ParamString("id")),
        bolt.String("name", "John Doe"),
        bolt.String("email", "john@example.com"),
    )
})
```

---

## API Comparison

### JSON Responses

**Sugared:**
```go
app.Get("/data", func(c *bolt.Context) error {
    return c.JSON(200, map[string]interface{}{
        "count":     42,
        "message":   "success",
        "timestamp": time.Now(),
    })
})
```

**Fast:**
```go
app.GetFast("/data", func(c *bolt.FastContext) error {
    return c.JSONFields(200,
        bolt.Int("count", 42),
        bolt.String("message", "success"),
        bolt.Time("timestamp", time.Now()),
    )
})
```

### Text Responses

**Sugared:**
```go
app.Get("/hello", func(c *bolt.Context) error {
    return c.String(200, "Hello, World!")
})
```

**Fast:**
```go
var helloWorld = []byte("Hello, World!")

app.GetFast("/hello", func(c *bolt.FastContext) error {
    return c.Text(200, helloWorld) // Zero allocation!
})
```

### Error Responses

**Sugared:**
```go
if err != nil {
    return c.JSON(404, map[string]string{
        "error": "user not found",
    })
}
```

**Fast:**
```go
if err != nil {
    return c.NotFound(bolt.String("error", "user not found"))
}
```

---

## Strongly-Typed Field Builders

The Fast API uses strongly-typed field constructors (inspired by Zap):

```go
bolt.String("key", "value")           // string value
bolt.Int("key", 42)                   // int value
bolt.Int64("key", 12345)              // int64 value
bolt.Float64("key", 3.14)             // float64 value
bolt.Bool("key", true)                // boolean value
bolt.Bytes("key", []byte("data"))     // byte slice (zero-copy)
bolt.Time("key", time.Now())          // time.Time (RFC3339 format)
bolt.Duration("key", time.Second*30)  // time.Duration (string format)
bolt.Any("key", customStruct)         // any type (uses reflection - slower)
```

### Why Strongly-Typed?

1. **No Reflection**: Types are known at compile-time
2. **Zero Allocations**: Direct JSON encoding without intermediate representations
3. **Type Safety**: Compiler catches type errors
4. **Performance**: 2-5x faster than `map[string]interface{}`

---

## Route Registration

All HTTP methods have both Sugared and Fast versions:

```go
// Sugared versions
app.Get(path, func(c *bolt.Context) error { ... })
app.Post(path, func(c *bolt.Context) error { ... })
app.Put(path, func(c *bolt.Context) error { ... })
app.Delete(path, func(c *bolt.Context) error { ... })
app.Patch(path, func(c *bolt.Context) error { ... })

// Fast versions
app.GetFast(path, func(c *bolt.FastContext) error { ... })
app.PostFast(path, func(c *bolt.FastContext) error { ... })
app.PutFast(path, func(c *bolt.FastContext) error { ... })
app.DeleteFast(path, func(c *bolt.FastContext) error { ... })
app.PatchFast(path, func(c *bolt.FastContext) error { ... })
```

---

## FastContext Convenience Methods

FastContext provides HTTP status helpers with sensible defaults:

```go
c.OK(fields...)                    // 200 OK
c.Created(fields...)               // 201 Created
c.NoContent()                      // 204 No Content
c.BadRequest(fields...)            // 400 Bad Request
c.Unauthorized(fields...)          // 401 Unauthorized
c.Forbidden(fields...)             // 403 Forbidden
c.NotFound(fields...)              // 404 Not Found
c.InternalServerError(fields...)   // 500 Internal Server Error
```

**Example:**
```go
app.PostFast("/users", func(c *bolt.FastContext) error {
    var user User
    if err := c.BindJSON(&user); err != nil {
        return c.BadRequest(bolt.String("error", "invalid JSON"))
    }

    if !user.Valid() {
        return c.BadRequest(
            bolt.String("error", "validation failed"),
            bolt.String("field", user.InvalidField()),
        )
    }

    // Save user...

    return c.Created(
        bolt.String("id", user.ID),
        bolt.String("status", "created"),
    )
})
```

---

## Hybrid Approach - Best of Both Worlds

You can **upgrade** from Sugared to Fast mid-handler using `c.Fast()`:

```go
app.Get("/adaptive", func(c *bolt.Context) error {
    // Check if high-performance mode is needed
    if c.GetHeader("X-Fast-Mode") == "true" {
        // Upgrade to FastContext
        return c.Fast().JSONFields(200,
            bolt.String("mode", "fast"),
            bolt.Bool("optimized", true),
        )
    }

    // Use sugared API for normal requests
    return c.JSON(200, map[string]interface{}{
        "mode": "normal",
        "optimized": false,
    })
})
```

---

## Performance Characteristics

### Allocations per Request

| Operation | Sugared API | Fast API | Improvement |
|-----------|-------------|----------|-------------|
| Simple JSON (3 fields) | ~4 allocs | ~1 alloc | **4x fewer** |
| Complex JSON (10 fields) | ~15 allocs | ~2 allocs | **7.5x fewer** |
| Text response | ~1 alloc | **0 allocs** | **100% fewer** |
| Error response | ~3 allocs | ~1 alloc | **3x fewer** |

### Speed Comparison

```
BenchmarkSugaredJSON-12    1000000    1250 ns/op    856 B/op    12 allocs/op
BenchmarkFastJSON-12       2000000     420 ns/op    128 B/op     1 allocs/op
```

**Fast API is ~3x faster with ~87% fewer allocations**

---

## When to Use Which

### Use Sugared API When:
- ✅ Prototyping or MVP development
- ✅ CRUD endpoints with moderate traffic
- ✅ Internal tools and dashboards
- ✅ Developer productivity > microseconds
- ✅ Code readability is prioritized
- ✅ Traffic < 1000 req/sec

### Use Fast API When:
- ⚡ High-throughput services (>10K req/sec)
- ⚡ Microservices mesh with strict latency SLAs
- ⚡ Serverless/Lambda (minimize cold starts)
- ⚡ Real-time APIs (gaming, trading, streaming)
- ⚡ Public APIs with unpredictable load
- ⚡ Every allocation matters for GC pressure

### Use Hybrid Approach When:
- 🔀 Different endpoints have different performance needs
- 🔀 Want to optimize hot paths while keeping others simple
- 🔀 A/B testing performance improvements
- 🔀 Gradually migrating from Sugared to Fast

---

## Real-World Example

```go
package main

import (
    "time"
    "bolt"
)

// Pre-compute responses for FastContext
var (
    healthOK = []byte(`{"status":"healthy"}`)
)

func main() {
    app := bolt.New()

    // Health check - CRITICAL PATH, use Fast API
    app.GetFast("/health", func(c *bolt.FastContext) error {
        return c.Text(200, healthOK) // Zero allocations!
    })

    // Metrics - HIGH TRAFFIC, use Fast API
    app.GetFast("/metrics", func(c *bolt.FastContext) error {
        return c.JSONFields(200,
            bolt.Int64("requests", getRequestCount()),
            bolt.Float64("latency", getAvgLatency()),
            bolt.Int("connections", getActiveConnections()),
            bolt.Time("timestamp", time.Now()),
        )
    })

    // User CRUD - MODERATE TRAFFIC, use Sugared API for ergonomics
    app.Get("/users/:id", func(c *bolt.Context) error {
        user, err := getUserByID(c.Param("id"))
        if err != nil {
            return c.JSON(404, map[string]string{
                "error": "user not found",
            })
        }
        return c.JSON(200, user)
    })

    // Admin endpoints - LOW TRAFFIC, use Sugared API
    app.Get("/admin/dashboard", func(c *bolt.Context) error {
        return c.JSON(200, map[string]interface{}{
            "stats": getAdminStats(),
            "users": getUserStats(),
            "system": getSystemInfo(),
        })
    })

    app.Listen(":8080")
}
```

---

## Migration Path

### Step 1: Start with Sugared
```go
app.Get("/api/data", func(c *bolt.Context) error {
    return c.JSON(200, getData())
})
```

### Step 2: Identify Hot Paths
Profile your application and find high-traffic endpoints.

### Step 3: Upgrade to Fast
```go
app.GetFast("/api/data", func(c *bolt.FastContext) error {
    data := getData()
    return c.JSONFields(200,
        bolt.Int("count", data.Count),
        bolt.String("status", data.Status),
        // ... convert to strongly-typed fields
    )
})
```

### Step 4: Benchmark & Compare
```bash
go test -bench=. -benchmem
```

---

## Philosophy

1. **No Forced Choices**: Use what makes sense for each endpoint
2. **Progressive Optimization**: Start simple, optimize when needed
3. **Explicit Trade-offs**: See exactly what you're trading
4. **Honest Performance**: No framework magic, just good APIs

> "Premature optimization is the root of all evil, but knowing where to optimize is wisdom"

Bolt gives you **both APIs** so you can be productive when prototyping and performant when scaling.

---

## Quick Reference

| Feature | Sugared | Fast |
|---------|---------|------|
| **Syntax** | `c.JSON(200, map...)` | `c.JSONFields(200, bolt.String(...))` |
| **Type Safety** | Runtime | Compile-time |
| **Allocations** | Low | Minimal |
| **Reflection** | Yes | No |
| **Speed** | Fast | Fastest |
| **Learning Curve** | Easy | Medium |
| **Best For** | Prototypes, CRUD | Hot paths, High-traffic |

---

See [examples/sugared_vs_fast/main.go](../examples/sugared_vs_fast/main.go) for complete working examples.
