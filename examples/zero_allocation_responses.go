package main

import (
	"log"

	"bolt"
)

// Pre-compute common responses at package level for zero allocations
var (
	helloWorldBytes = []byte("Hello, World!")
	okResponse      = []byte("OK")
	notFoundMsg     = []byte("Not Found")
	welcomeHTML     = []byte("<h1>Welcome!</h1>")
	apiPrefix       = []byte("API v1.0 - ")
)

func main() {
	app := bolt.New(bolt.WithDevMode(true))

	// OPTION 1: Normal API - Simple and ergonomic (small allocation cost)
	app.Get("/hello", func(c *bolt.Context) error {
		return c.String(200, "Hello, World!")
	})

	// OPTION 2: Zero-Allocation API - Maximum performance
	// Use StringBytes() with pre-allocated byte slices
	app.Get("/hello-fast", func(c *bolt.Context) error {
		return c.StringBytes(200, helloWorldBytes)
	})

	// OPTION 3: Dynamic zero-allocation responses
	// Pre-compute at package level, use in handlers
	app.Get("/status", func(c *bolt.Context) error {
		return c.StringBytes(200, okResponse)
	})

	// OPTION 4: Pre-computed responses in a map for flexibility
	responses := map[string][]byte{
		"success": []byte("Operation successful"),
		"error":   []byte("Operation failed"),
		"pending": []byte("Operation pending"),
	}

	app.Get("/result/:status", func(c *bolt.Context) error {
		status := c.Param("status")
		if msg, ok := responses[status]; ok {
			return c.StringBytes(200, msg)
		}
		return c.StringBytes(404, notFoundMsg)
	})

	// OPTION 5: Combining dynamic content with pre-allocated prefixes
	// This still allocates but shows the pattern
	app.Get("/api/info", func(c *bolt.Context) error {
		// For truly zero-allocation, you'd need to use a buffer pool
		// But this demonstrates the concept
		info := append(apiPrefix, []byte("All systems operational")...)
		return c.StringBytes(200, info)
	})

	// OPTION 6: Using FastText for common precomputed responses
	// FastText checks a global map of precomputed responses first
	app.Get("/ok", func(c *bolt.Context) error {
		return c.FastText(200, []byte("ok")) // Uses precomputed "OK" response
	})

	// OPTION 7: Using FastJSON for precomputed JSON responses
	app.Get("/status-json", func(c *bolt.Context) error {
		return c.FastJSON(200, []byte("json_ok")) // Uses precomputed {"status":"ok"}
	})

	// OPTION 8: Best of both worlds - check request and choose path
	app.Get("/adaptive", func(c *bolt.Context) error {
		// For hot paths or high-traffic endpoints, use zero-allocation
		if c.Query("fast") == "true" {
			return c.StringBytes(200, helloWorldBytes)
		}
		// For regular requests, use normal ergonomic API
		return c.String(200, "Hello, World!")
	})

	log.Println("🚀 Zero-allocation example server on http://localhost:3000")
	log.Println("\nEndpoints:")
	log.Println("  GET /hello          - Normal API (ergonomic)")
	log.Println("  GET /hello-fast     - Zero-allocation API (maximum performance)")
	log.Println("  GET /status         - Pre-computed response")
	log.Println("  GET /result/:status - Map-based pre-computed responses")
	log.Println("  GET /api/info       - Combined prefix + dynamic content")
	log.Println("  GET /ok             - FastText with global precomputed map")
	log.Println("  GET /status-json    - FastJSON with global precomputed map")
	log.Println("  GET /adaptive?fast=true - Adaptive response based on query")

	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
