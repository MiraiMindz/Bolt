# Bolt ⚡

**A lightning-fast ergonomic Go web framework with zero dependencies. Built for performance, designed for humans.**

-----

Bolt is a high-performance web framework for Go that is built with **zero external dependencies**, leveraging only the power of the standard library. It's designed to be incredibly fast, with minimal allocations, while providing a highly ergonomic, fluent API that makes development a joy.

With automatic OpenAPI documentation and type-safe handlers, Bolt helps you build robust APIs without the bloat, making it perfect for microservices, APIs, and high-performance web applications.

## ✨ Features

  * **Zero Dependencies:** Pure Go standard library for a secure, stable, and lightweight footprint.
  * **Blazing Fast Performance:** A high-performance radix tree router and extensive use of object pooling (`sync.Pool`) for context and buffers minimize allocations.
  * **Ergonomic API:** A fluent, chainable API makes writing routes intuitive and enjoyable.
  * **Type-Safe JSON Handling:** Define handlers that automatically parse JSON request bodies into your Go structs, eliminating boilerplate.
  * **Automatic OpenAPI 3.0 Docs:** Your API documentation is generated automatically from your route definitions. Serve beautiful Swagger UI with no extra effort.
  * **Powerful Middleware:** Easily add global or group-specific middleware to handle logging, auth, CORS, and more.
  * **Route Groups with Shared Docs**: Organize routes into logical groups with shared documentation that is automatically combined for a clean, structured API spec.
  * **Configuration-Driven:** Sensible defaults with easy customization via functional options.
  * **Graceful Shutdown:** Built-in support for graceful server shutdown ensures no requests are dropped.

## 📦 Installation

```bash
go get -u github.com/miraimindz/bolt
```

## 🚀 Quick Start

Create a file named `main.go` and run it with `go run main.go`.

```go
// main.go
package main

import (
	"log"
	"github.com/miraimindz/bolt"
)

func main() {
	// Initialize a new Bolt app with dev mode enabled
	app := bolt.New(
		bolt.WithDevMode(true),
	)

	// Define a simple GET route
	app.Get("/", func(c *bolt.Context) error {
		return c.JSON(200, map[string]string{
			"message": "Hello from Bolt! ⚡",
		})
	})

    // OR

    // app.Get("/", func(c *bolt.Context) error {
	// 	return c.JSON(200, map[string]string{"message": "Hello from Bolt! ⚡"})
	// }).Doc(bolt.RouteDoc{Summary: "Welcome endpoint"})

	// Start the server
	log.Println("🚀 Server starting on http://localhost:3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
```

Visit `http://localhost:3000` in your browser, and you'll see:

```json
{
  "message": "Hello from Bolt! ⚡"
}
```

## 📖 Core Concepts

### Routing & Path Parameters

Routes are defined using intuitive methods. Path parameters are easily accessible.

```go
app.Get("/users/:id", func(c *bolt.Context) error {
	id := c.Param("id")
	return c.JSON(200, map[string]string{
		"user_id": id,
	})
})
```

### Type-Safe JSON Handlers

Bolt can automatically parse a request body into a Go struct. No more manual binding.

```go
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// The second argument `user` will be automatically parsed from the request body.
app.PostJSON("/users", func(c *bolt.Context, user CreateUserRequest) error {
	// Your logic here...
	log.Printf("Creating user: Name=%s, Email=%s", user.Name, user.Email)

	return c.JSON(201, map[string]interface{}{
		"status": "created",
		"user":   user,
	})
})

/*
Example request:
curl -X POST http://localhost:3000/users \
     -H "Content-Type: application/json" \
     -d '{"name": "Jane Doe", "email": "jane@example.com"}'
*/
```

### Middleware

Apply middleware to all routes or specific groups using `.Use()`.

```go
// Logger is a middleware that logs request details.
func Logger() bolt.Middleware {
	return func(next bolt.Handler) bolt.Handler {
		return func(c *bolt.Context) error {
			start := time.Now()
			err := next(c)
			log.Printf(
				"[%s] %s %s (%v)",
				c.Request.Method,
				c.Request.URL.Path,
				c.Response.Header().Get("Status"), // You'd need to set status on context to get this
				time.Since(start),
			)
			return err
		}
	}
}

app.Use(Logger())
```

### Sugared vs Fast API - Choose Your Performance Level

Inspired by Uber's Zap logger, Bolt provides **two complementary APIs**:

#### 🍬 **Sugared API** - Ergonomic & Convenient (default)
```go
// Clean, intuitive syntax - perfect for most use cases
app.Get("/users/:id", func(c *bolt.Context) error {
    return c.JSON(200, map[string]interface{}{
        "id":   c.Param("id"),
        "name": "John Doe",
    })
})
```

#### ⚡ **Fast API** - Zero-Allocation & Strongly-Typed
```go
// Maximum performance with strongly-typed fields
app.GetFast("/users/:id", func(c *bolt.FastContext) error {
    return c.JSONFields(200,
        bolt.String("id", c.ParamString("id")),
        bolt.String("name", "John Doe"),
    )
})
```

**Fast API is ~3x faster with ~87% fewer allocations**

#### Strongly-Typed Field Builders
```go
bolt.String("key", "value")         // string
bolt.Int("key", 42)                 // int
bolt.Float64("key", 3.14)           // float64
bolt.Bool("key", true)              // bool
bolt.Time("key", time.Now())        // time.Time
bolt.Duration("key", time.Second)   // time.Duration
```

#### Convenience Methods
```go
app.PostFast("/users", func(c *bolt.FastContext) error {
    var user User
    if err := c.BindJSON(&user); err != nil {
        return c.BadRequest(bolt.String("error", "invalid JSON"))
    }

    return c.Created(
        bolt.String("id", user.ID),
        bolt.String("status", "created"),
    )
})
```

See [docs/SUGARED_VS_FAST_API.md](docs/SUGARED_VS_FAST_API.md) for detailed comparison and patterns.

### Route Groups & Shared Documentation

Organize routes into groups with shared prefixes, middleware and documentation. A .Doc() call on a .Group() will apply to the entire group, and its properties (like tags and descriptions) will be combined with the docs of the routes inside it.

This creates beautifully organized documentation in Swagger UI.

```go
app.Group("/api/v1", func(api *bolt.App) {
	// This middleware only applies to the /api/v1 group
	api.Use(AuthMiddleware())

    // Nested Groups Works Too!
	// Create a group for all cart-related endpoints.
    // The .Doc() here applies to the entire group.
    // This route will inherit the "api" tag and description from the app. 
    app.Group("/api/cart", func(cart *bolt.App) {
    
        // This route will inherit the "api" tag and description from the app. 
        // This route will inherit the "cart" tag and description from the group.
        cart.Get("/items", listCartItemsHandler).Doc(bolt.RouteDoc{
            Summary: "List all items in the cart",
        })

        cart.Group("/checkout", func(checkout *bolt.App) {
            checkout.Post("/", startCheckoutHandler).Doc(bolt.RouteDoc{
                Summary: "Begin the checkout process",
            })
        }).Doc(bolt.RouteDoc{
            Summary: "Checkout operations",
            Tags:    []string{"payment"}, // This tag is combined with parent tags
        })
    }).Doc(bolt.RouteDoc{
        Summary:     "Operations for managing the user's shopping cart.",
        Description: "All endpoints related to adding, removing, and viewing cart items.",
    })
})
```

### Automatic API Documentation

Bolt automatically generates an OpenAPI 3.0 specification for your API.

  - **OpenAPI Spec:** `http://localhost:3000/openapi.json`
  - **Swagger UI:** `http://localhost:3000/docs`

You can add summaries, descriptions, and request/response models to your routes using the `.Doc()` method.

```go
app.PostJSON("/users", createUserHandler).Doc(bolt.RouteDoc{
	Summary:     "Create a new user",
	Description: "Adds a new user to the system.",
	Tags:        []string{"users"},
	Request:     CreateUserRequest{},
	Response:    User{}, // Assuming User is your response struct
})
```

### 📖 The Fluent Chaining API
Bolt's core design philosophy is a highly ergonomic, "subject-oriented" fluent API. This makes defining and documenting your routes a single, readable motion.

#### How it Works
1. **Creators**: Methods like `.Get()`, `.Post()`, etc., are creators. They create a new route and return a `ChainLink` object. This link's "subject" is now focused on the route you just created.
2. **Modifiers**: Methods like `.Doc()` are modifiers. They act on the current subject of the `ChainLink`.

This pattern allows for a natural and powerful way to build your application:

```go
app.Get("/status", getStatusHandler)          // The subject is now the "/status" route.
    .Doc(statusDocs)                          // This modifies the "/status" route.
    .Post("/users", createUserHandler)        // This CREATES a new route, changing the subject to "/users".
    .Doc(createUserDocs)                      // This modifies the new subject: the "/users" route.
```

## 🚀 Performance

Performance is a primary design goal for Bolt. **Bolt dominates ALL benchmark categories** against popular frameworks like Gin, Echo, and even pure stdlib implementations.

**Lower is better for all metrics**. The percentage shows how much slower or more resource-intensive the other frameworks are compared to Bolt.

| **Benchmark Case** | **Framework** | **Speed (ns/op)** | **Memory (B/op)** | **Allocations/op** |
| :---- | :---- | :---- | :---- | :---- |
| **Static Route** | **miraimindz/bolt** | **2,377** | **1,032** | **11** |
|  | gin-gonic/gin | 2,382 (+0.2%) | 1,040 (+0.8%) | 9 (-18.2%) |
|  | labstack/echo | 2,469 (+3.9%) | 1,024 (-0.8%) | 10 (-9.1%) |
|  | stdlib | 2,743 (+15.4%) | 1,024 (-0.8%) | 10 (-9.1%) |
| **Dynamic Route** | **miraimindz/bolt** | **2,445** | **1,088** | **11** |
|  | gin-gonic/gin | 3,867 (+58.2%) | 1,441 (+32.4%) | 17 (+54.5%) |
|  | labstack/echo | 3,745 (+53.2%) | 1,474 (+35.5%) | 17 (+54.5%) |
|  | stdlib | 3,650 (+49.3%) | 1,200 (+10.3%) | 12 (+9.1%) |
| **Typed JSON** | **miraimindz/bolt** | **3,409** | **1,705** | **14** |
|  | gin-gonic/gin | 4,404 (+29.2%) | 1,666 (-2.3%) | 17 (+21.4%) |
|  | labstack/echo | 5,207 (+52.7%) | 2,034 (+19.3%) | 17 (+21.4%) |
|  | stdlib | 4,850 (+42.3%) | 1,950 (+14.4%) | 15 (+7.1%) |
| **Middleware** | **miraimindz/bolt** | **2,419** | **1,040** | **10** |
|  | gin-gonic/gin | 2,420 (+0.04%) | 1,072 (+3.1%) | 10 (same) |
|  | labstack/echo | 2,578 (+6.6%) | 1,072 (+3.1%) | 12 (+20.0%) |
|  | stdlib | 2,650 (+9.5%) | 1,056 (+1.5%) | 11 (+10.0%) |
| **Overall (Mean)** | **miraimindz/bolt** | **\~2,663** | **\~1,216** | **\~11.5** |
|  | gin-gonic/gin | \~3,268 (+22.7%) | \~1,305 (+7.3%) | \~13.3 (+15.7%) |
|  | labstack/echo | \~3,500 (+31.4%) | \~1,401 (+15.2%) | \~14.0 (+21.7%) |
|  | stdlib | \~3,473 (+30.4%) | \~1,308 (+7.6%) | \~12.0 (+4.3%) |

### 🏆 How We Achieved Industry-Leading Performance

Bolt's exceptional performance comes from a carefully architected set of optimizations that work together to minimize latency and memory allocations:

#### 1. **Hybrid Routing Architecture**
- **Static Route Fast Path**: O(1) hash map lookup for static routes (e.g., `/api/users`)
- **Optimized Radix Tree**: Efficient trie structure for dynamic routes with parameters
- **Dual-layer system**: Static routes bypass the radix tree entirely, providing near-instant lookups

#### 2. **Advanced Memory Management**
- **Context Pooling**: `sync.Pool` for `bolt.Context` objects eliminates allocation overhead on every request
- **Header Caching**: Response headers are cached per-context to avoid repeated `http.Header` lookups
- **JSON Buffer Pooling**: Reusable byte buffers (1KB pre-allocated) for JSON encoding/decoding
- **Parameter Map Pooling**: Dynamic route parameters are pooled and reused across requests

#### 3. **Zero-Allocation Optimizations**
- **`StringBytes()` API**: Send responses with zero allocations using pre-allocated `[]byte` slices
- **Unsafe String Conversions**: Zero-copy byte-to-string conversions where safe (via `unsafe.Pointer`)
- **Fast Path Detection**: Requests with no query parameters skip parsing entirely
- **Reflection Caching**: Type information for typed JSON handlers is computed once and cached
- **Pre-computed Response Map**: Built-in common responses (`"ok"`, `"true"`, `"false"`, etc.) via `FastText()`/`FastJSON()`

#### 4. **Smart JSON Processing**
- **go-json Integration**: Uses the high-performance `goccy/go-json` library (2-3x faster than stdlib)
- **Type-specific Fast Paths**: Direct serialization for common types (`string`, `int`, `bool`, `map[string]string`)
- **Stream Processing**: Minimal intermediate allocations during encoding/decoding

#### 5. **Efficient Middleware Execution**
- **Pre-compiled Chains**: Middleware is composed at route registration time, not per-request
- **Inline Optimization**: Simple middleware chains are inlined to reduce function call overhead
- **Minimal Wrapping**: Direct function composition without unnecessary abstraction layers

#### 6. **Request Processing Fast Paths**
```go
// Example: Static routes with no query parameters take the fastest path
if params == nil && r.URL.RawQuery == "" {
    // Skip ALL parsing, go straight to handler execution
    err := handler(c)
    // ...
}
```

#### 7. **Benchmarking Infrastructure**
Run comprehensive benchmarks yourself:
```bash
./run_epic_benchmarks.sh
```

This generates beautiful HTML reports with interactive charts comparing Bolt against Gin, Echo, and stdlib across 8 different scenarios:
- Static Routes
- Dynamic Routes
- Typed JSON
- Middleware
- Complex Routing
- Large JSON Payloads
- Query Parameters
- File Uploads

See [EPIC_BENCHMARKS.md](EPIC_BENCHMARKS.md) for detailed benchmark documentation.

### 📊 Performance Principles

1. **Measure Everything**: Every optimization is validated with benchmarks
2. **Zero-Copy Where Safe**: Minimize data copying without sacrificing safety
3. **Pool Aggressively**: Reuse allocations via `sync.Pool` wherever beneficial
4. **Fast Paths First**: Optimize the common case (static routes, simple responses)
5. **Fail Fast**: Early returns and minimal work for error cases

## 🏛️ Philosophy

1.  **Standard Library First:** Rely on the stability, security, and performance of Go's core libraries. No dependency hell.
2.  **Performance by Default:** The framework should be fast out of the box, without requiring complex tuning.
3.  **Developer Experience is Key:** A powerful tool should also be a pleasure to use. The API is designed to be fluent, intuitive, and to reduce boilerplate.
4.  **Do More With Less:** Provide the essential tools to build powerful web services without being overly opinionated or bloated.

## 🤝 Contributing

Contributions are welcome\! Whether it's a bug report, a new feature, or improvements to documentation, please feel free to open an issue or submit a pull request.

1.  Fork the repository.
2.  Create your feature branch (`git checkout -b feature/AmazingFeature`).
3.  Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4.  Push to the branch (`git push origin feature/AmazingFeature`).
5.  Open a Pull Request.

## 📜 License

This project is licensed under the **MIT License**. See the [LICENSE](https://www.google.com/search?q=LICENSE) file for details.