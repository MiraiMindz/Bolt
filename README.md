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

Performance is a primary design goal for Bolt.

| Framework         | Requests/sec | Allocations/op |
| ----------------- | ------------ | -------------- |
| **Bolt (Static)** | \~185,000     | 0 allocs/op    |
| **Bolt (Params)** | \~140,000     | 2 allocs/op    |
| `net/http`        | \~75,000      | 0 allocs/op    |

*\*Benchmarks are illustrative. Please run your own tests for accurate numbers.*

This is achieved through:

  * **Radix Tree Router:** For fast O(log n) route lookups.
  * **Object Pooling:** `sync.Pool` is used for `bolt.Context` and I/O buffers to dramatically reduce garbage collection pressure.
  * **Zero-Copy Operations:** Careful use of interfaces and streaming to avoid unnecessary data copies.

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