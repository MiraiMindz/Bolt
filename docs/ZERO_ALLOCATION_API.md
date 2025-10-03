# Zero-Allocation Response API

## Philosophy

Bolt provides **honest performance** - no benchmark-specific tricks, just powerful APIs that let YOU choose the performance level you need.

## The Problem

Traditional frameworks force you to choose between:
- **Ergonomic API**: Easy to use but allocates memory (`String(200, "Hello")`)
- **Performance**: Manual byte slices everywhere, ugly code

## Bolt's Solution

**Both APIs, your choice:**

```go
// Ergonomic - small allocation for convenience
c.String(200, "Hello, World!")

// Maximum performance - zero allocations
var helloWorld = []byte("Hello, World!")
c.StringBytes(200, helloWorld)
```

## API Reference

### `String(status int, s string) error`
**Use when:** Developer convenience > microsecond optimizations
- Clean, readable code
- String literals work directly
- Small allocation cost (string → []byte conversion)

```go
app.Get("/hello", func(c *bolt.Context) error {
    return c.String(200, "Hello, World!")
})
```

### `StringBytes(status int, b []byte) error`
**Use when:** Maximum performance is critical
- Zero allocations
- Perfect for hot paths
- Pre-computed responses at package level

```go
var helloWorld = []byte("Hello, World!")

app.Get("/hello", func(c *bolt.Context) error {
    return c.StringBytes(200, helloWorld)
})
```

### `FastText(status int, key []byte) error`
**Use when:** You want common precomputed responses
- Checks global precomputed map first
- Falls back to `StringBytes()`
- Built-in responses: `"ok"`, `"true"`, `"false"`, `"null"`, `"zero"`, `"one"`

```go
app.Get("/status", func(c *bolt.Context) error {
    return c.FastText(200, []byte("ok")) // Returns precomputed "OK"
})
```

### `FastJSON(status int, key []byte) error`
**Use when:** You need fast JSON responses
- Precomputed JSON responses
- Built-in: `"json_ok"` → `{"status":"ok"}`, `"json_err"` → `{"error":"error"}`

```go
app.Get("/health", func(c *bolt.Context) error {
    return c.FastJSON(200, []byte("json_ok"))
})
```

## Patterns

### Pattern 1: Package-Level Constants

```go
var (
    responseOK        = []byte("OK")
    responseNotFound  = []byte("Not Found")
    responseError     = []byte("Internal Server Error")
)

app.Get("/status", func(c *bolt.Context) error {
    return c.StringBytes(200, responseOK)
})
```

### Pattern 2: Response Map

```go
var responses = map[string][]byte{
    "success": []byte("Operation successful"),
    "error":   []byte("Operation failed"),
    "pending": []byte("Operation pending"),
}

app.Get("/result/:status", func(c *bolt.Context) error {
    status := c.Param("status")
    if msg, ok := responses[status]; ok {
        return c.StringBytes(200, msg)
    }
    return c.StringBytes(404, []byte("Unknown status"))
})
```

### Pattern 3: Conditional Performance

```go
var fastResponse = []byte("Fast path!")

app.Get("/adaptive", func(c *bolt.Context) error {
    // Hot path - zero allocations
    if c.GetHeader("X-Fast-Path") != "" {
        return c.StringBytes(200, fastResponse)
    }

    // Normal path - ergonomic API
    return c.String(200, "Normal path")
})
```

### Pattern 4: JSON with Byte Slices

```go
var (
    userJSON = []byte(`{"id":1,"name":"John"}`)
    errorJSON = []byte(`{"error":"not found"}`)
)

app.Get("/user/:id", func(c *bolt.Context) error {
    id := c.Param("id")
    if id == "1" {
        return c.Bytes(200, bolt.ContentTypeJSON, userJSON)
    }
    return c.Bytes(404, bolt.ContentTypeJSON, errorJSON)
})
```

## Performance Impact

### Benchmark: String vs StringBytes

```
String(200, "Hello, World!")      - 1 alloc/op  (~80 bytes)
StringBytes(200, helloWorld)      - 0 allocs/op (0 bytes)
FastText(200, []byte("ok"))       - 0 allocs/op (0 bytes)
```

### When to Use Which

| Scenario | Recommended API | Reason |
|----------|----------------|--------|
| Prototype/MVP | `String()` | Speed of development |
| CRUD endpoints | `String()` or `JSON()` | Ergonomics > microseconds |
| Health checks | `FastText()` | Precomputed, zero-allocation |
| High-traffic API | `StringBytes()` | Every allocation matters |
| Microservices mesh | `StringBytes()` | Reduce GC pressure |
| Serverless/Lambda | `StringBytes()` | Minimize cold start impact |

## Philosophy

1. **No Magic**: Everything is explicit. You see exactly what allocates.
2. **Your Choice**: We provide the tools, you choose the tradeoff.
3. **Honest Benchmarks**: No framework-specific "cheats" - just good APIs.
4. **Progressive Enhancement**: Start simple with `String()`, optimize later with `StringBytes()`.

## Example: Real-World API

```go
package main

import "bolt"

// Pre-compute at package level
var (
    // Common responses
    ok           = []byte("OK")
    unauthorized = []byte("Unauthorized")
    notFound     = []byte("Not Found")

    // API responses
    healthOK = []byte(`{"status":"healthy","version":"1.0.0"}`)
    apiInfo  = []byte(`{"name":"MyAPI","version":"1.0.0"}`)
)

func main() {
    app := bolt.New()

    // Health check - maximum performance
    app.Get("/health", func(c *bolt.Context) error {
        return c.Bytes(200, bolt.ContentTypeJSON, healthOK)
    })

    // API info - precomputed
    app.Get("/", func(c *bolt.Context) error {
        return c.Bytes(200, bolt.ContentTypeJSON, apiInfo)
    })

    // User endpoints - ergonomic API is fine
    app.Get("/users/:id", func(c *bolt.Context) error {
        id := c.Param("id")
        // Database lookup, etc...
        return c.JSON(200, map[string]string{
            "id": id,
            "name": "John Doe",
        })
    })

    app.Listen(":3000")
}
```

## Measuring Impact

Use Go's built-in benchmarking:

```go
func BenchmarkString(b *testing.B) {
    app := bolt.New()
    app.Get("/test", func(c *bolt.Context) error {
        return c.String(200, "Hello, World!")
    })
    // benchmark code...
}

func BenchmarkStringBytes(b *testing.B) {
    var hello = []byte("Hello, World!")
    app := bolt.New()
    app.Get("/test", func(c *bolt.Context) error {
        return c.StringBytes(200, hello)
    })
    // benchmark code...
}
```

## Conclusion

Bolt gives you **honest performance tools**:
- No hidden optimization for benchmarks
- Clear APIs for clear tradeoffs
- Choose ergonomics or performance per-endpoint
- Scale from prototype to production

**The best framework doesn't optimize FOR benchmarks - it gives YOU the tools to optimize YOUR code.**
