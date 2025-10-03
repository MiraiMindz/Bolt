package main

import (
	"log"
	"time"

	"bolt"
)

// Pre-allocated responses for FastContext (zero-allocation)
var (
	helloBytes = []byte("Hello, World!")
	okBytes    = []byte("OK")
)

func main() {
	app := bolt.New(bolt.WithDevMode(true))

	// ============================================================================
	// SUGARED API - Ergonomic, easy to use, small allocations
	// ============================================================================

	// Simple GET route - normal API
	app.Get("/sugared/hello", func(c *bolt.Context) error {
		return c.String(200, "Hello from Sugared API!")
	})

	// JSON response - normal API (uses reflection)
	app.Get("/sugared/user/:id", func(c *bolt.Context) error {
		id := c.Param("id")

		return c.JSON(200, map[string]interface{}{
			"id":      id,
			"name":    "John Doe",
			"email":   "john@example.com",
			"created": time.Now(),
		})
	})

	// Complex JSON - normal API
	app.Get("/sugared/dashboard", func(c *bolt.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status": "online",
			"metrics": map[string]interface{}{
				"requests": 12345,
				"latency":  "45ms",
				"uptime":   "99.9%",
			},
			"timestamp": time.Now(),
		})
	})

	// ============================================================================
	// FAST API - Zero-allocation, strongly-typed, maximum performance
	// ============================================================================

	// Simple GET route - Fast API (zero allocations)
	app.GetFast("/fast/hello", func(c *bolt.FastContext) error {
		return c.Text(200, helloBytes)
	})

	// JSON response - Fast API (no reflection, strongly typed)
	app.GetFast("/fast/user/:id", func(c *bolt.FastContext) error {
		id := c.ParamString("id")

		return c.JSONFields(200,
			bolt.String("id", id),
			bolt.String("name", "John Doe"),
			bolt.String("email", "john@example.com"),
			bolt.Time("created", time.Now()),
		)
	})

	// Complex JSON - Fast API (zero-allocation field building)
	app.GetFast("/fast/dashboard", func(c *bolt.FastContext) error {
		return c.JSONFields(200,
			bolt.String("status", "online"),
			bolt.Int("requests", 12345),
			bolt.String("latency", "45ms"),
			bolt.Float64("uptime", 99.9),
			bolt.Time("timestamp", time.Now()),
		)
	})

	// ============================================================================
	// COMPARISON: Sugared vs Fast for the same endpoint
	// ============================================================================

	// Sugared: Dynamic params endpoint
	app.Post("/sugared/calculate", func(c *bolt.Context) error {
		var input struct {
			A int `json:"a"`
			B int `json:"b"`
		}

		if err := c.BindJSON(&input); err != nil {
			return c.JSON(400, map[string]string{
				"error": "invalid input",
			})
		}

		result := input.A + input.B

		return c.JSON(200, map[string]interface{}{
			"operation": "add",
			"a":         input.A,
			"b":         input.B,
			"result":    result,
			"timestamp": time.Now(),
		})
	})

	// Fast: Same endpoint with zero allocations
	app.PostFast("/fast/calculate", func(c *bolt.FastContext) error {
		var input struct {
			A int `json:"a"`
			B int `json:"b"`
		}

		if err := c.BindJSON(&input); err != nil {
			return c.BadRequest(bolt.String("error", "invalid input"))
		}

		result := input.A + input.B

		return c.JSONFields(200,
			bolt.String("operation", "add"),
			bolt.Int("a", input.A),
			bolt.Int("b", input.B),
			bolt.Int("result", result),
			bolt.Time("timestamp", time.Now()),
		)
	})

	// ============================================================================
	// FAST API - Convenience methods
	// ============================================================================

	app.GetFast("/fast/ok", func(c *bolt.FastContext) error {
		return c.OK(bolt.String("message", "all systems operational"))
	})

	app.PostFast("/fast/create", func(c *bolt.FastContext) error {
		return c.Created(
			bolt.String("id", "12345"),
			bolt.String("status", "created"),
		)
	})

	app.GetFast("/fast/notfound", func(c *bolt.FastContext) error {
		return c.NotFound(bolt.String("resource", "user not found"))
	})

	app.GetFast("/fast/error", func(c *bolt.FastContext) error {
		return c.InternalServerError(
			bolt.String("error", "database connection failed"),
			bolt.String("code", "DB_001"),
		)
	})

	// ============================================================================
	// HYBRID - Mix and match based on needs
	// ============================================================================

	// Regular route can upgrade to Fast when needed
	app.Get("/hybrid/adaptive", func(c *bolt.Context) error {
		// Check if we need high performance
		if c.GetHeader("X-Fast-Mode") == "true" {
			// Upgrade to FastContext
			return c.Fast().JSONFields(200,
				bolt.String("mode", "fast"),
				bolt.Bool("optimized", true),
				bolt.Time("timestamp", time.Now()),
			)
		}

		// Use normal sugared API
		return c.JSON(200, map[string]interface{}{
			"mode":      "normal",
			"optimized": false,
			"timestamp": time.Now(),
		})
	})

	// ============================================================================
	// START SERVER
	// ============================================================================

	log.Println("🚀 Sugared vs Fast API Demo")
	log.Println("============================")
	log.Println()
	log.Println("Sugared API (Ergonomic, Easy):")
	log.Println("  GET  /sugared/hello")
	log.Println("  GET  /sugared/user/:id")
	log.Println("  GET  /sugared/dashboard")
	log.Println("  POST /sugared/calculate")
	log.Println()
	log.Println("Fast API (Zero-Allocation, Strongly-Typed):")
	log.Println("  GET  /fast/hello")
	log.Println("  GET  /fast/user/:id")
	log.Println("  GET  /fast/dashboard")
	log.Println("  POST /fast/calculate")
	log.Println("  GET  /fast/ok")
	log.Println("  POST /fast/create")
	log.Println("  GET  /fast/notfound")
	log.Println("  GET  /fast/error")
	log.Println()
	log.Println("Hybrid API (Adaptive):")
	log.Println("  GET  /hybrid/adaptive")
	log.Println("       Add header 'X-Fast-Mode: true' for zero-allocation mode")
	log.Println()
	log.Println("Server running on http://localhost:3000")

	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
