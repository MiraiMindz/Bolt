package benchmarks

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bolt"
)

// Fast (Unsugared) API benchmarks - using zero-allocation FastContext API

var (
	// Pre-allocated responses for zero-allocation benchmarks
	helloWorldBytes = []byte("Hello, World!")
)

func BenchmarkFastStaticRoute(b *testing.B) {
	app := bolt.New(bolt.WithDocs(false))
	app.Get("/hello", func(c *bolt.Context) error {
		return c.Text(200, helloWorldBytes)
	})

	req := httptest.NewRequest("GET", "/hello", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkFastDynamicRoute(b *testing.B) {
	app := bolt.New(bolt.WithDocs(false))
	app.Get("/user/:id", func(c *bolt.Context) error {
		id := c.Param("id")
		return c.JSONFields(200,
			bolt.String("id", id),
		)
	})

	req := httptest.NewRequest("GET", "/user/123", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkFastTypedJSON(b *testing.B) {
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	app := bolt.New(bolt.WithDocs(false))
	app.Post("/users", func(c *bolt.Context) error {
		var user User
		if err := c.BindJSON(&user); err != nil {
			return err
		}

		return c.JSONFields(201,
			bolt.String("name", user.Name),
			bolt.String("email", user.Email),
		)
	})

	body := `{"name":"John","email":"john@example.com"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkFastMiddleware(b *testing.B) {
	app := bolt.New(bolt.WithDocs(false))
	app.Use(func(next bolt.Handler) bolt.Handler {
		return func(c *bolt.Context) error {
			c.Response.Header().Set("X-Test-Header", "test")
			return next(c)
		}
	})
	app.Get("/hello", func(c *bolt.Context) error {
		return c.Text(200, helloWorldBytes)
	})

	req := httptest.NewRequest("GET", "/hello", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

// Extended benchmarks for Fast API
func BenchmarkFastComplexRouting(b *testing.B) {
	app := bolt.New(bolt.WithDocs(false))

	app.Get("/api/v1/users", func(c *bolt.Context) error {
		// For arrays, we still need to use regular JSON
		// but with strongly-typed fields for each object
		return c.JSON(200, []map[string]interface{}{
			{"id": 1, "name": "User1"},
			{"id": 2, "name": "User2"},
		})
	})

	app.Get("/api/v1/users/:id", func(c *bolt.Context) error {
		id := c.Param("id")
		return c.JSONFields(200,
			bolt.String("id", id),
			bolt.String("name", "User"+id),
		)
	})

	type NewUser struct {
		Name string `json:"name"`
	}
	app.Post("/api/v1/users", func(c *bolt.Context) error {
		var user NewUser
		if err := c.BindJSON(&user); err != nil {
			return err
		}

		return c.JSONFields(201,
			bolt.Int("id", 456),
			bolt.String("name", user.Name),
		)
	})

	requests := []*http.Request{
		httptest.NewRequest("GET", "/api/v1/users", nil),
		httptest.NewRequest("GET", "/api/v1/users/123", nil),
		httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(`{"name":"NewUser"}`)),
	}
	for _, req := range requests {
		if req.Method == "POST" {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := requests[i%len(requests)]
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkFastLargeJSON(b *testing.B) {
	type LargeObject struct {
		ID          int                    `json:"id"`
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Tags        []string               `json:"tags"`
		Metadata    map[string]interface{} `json:"metadata"`
		Items       []map[string]string    `json:"items"`
	}

	app := bolt.New(bolt.WithDocs(false))
	app.Post("/large", func(c *bolt.Context) error {
		var obj LargeObject
		if err := c.BindJSON(&obj); err != nil {
			return err
		}

		// For complex nested structures, fallback to regular JSON
		// but this still benefits from FastContext's other optimizations
		return c.JSON(200, obj)
	})

	largePayload := `{
		"id": 12345,
		"name": "Large Object Test",
		"description": "This is a large JSON object for testing serialization performance with multiple nested fields and arrays",
		"tags": ["performance", "testing", "json", "serialization", "benchmark"],
		"metadata": {
			"created_at": "2025-01-01T00:00:00Z",
			"updated_at": "2025-01-01T12:00:00Z",
			"version": "1.0.0",
			"author": "benchmark-test"
		},
		"items": [
			{"key1": "value1", "key2": "value2", "key3": "value3"},
			{"key1": "value4", "key2": "value5", "key3": "value6"},
			{"key1": "value7", "key2": "value8", "key3": "value9"}
		]
	}`

	req := httptest.NewRequest("POST", "/large", strings.NewReader(largePayload))
	req.Header.Set("Content-Type", "application/json")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkFastQueryParameters(b *testing.B) {
	app := bolt.New(bolt.WithDocs(false))
	app.Get("/search", func(c *bolt.Context) error {
		page := c.QueryInt("page", 1)
		limit := c.QueryInt("limit", 10)
		query := c.Query("q")

		return c.JSONFields(200,
			bolt.Int("page", page),
			bolt.Int("limit", limit),
			bolt.String("query", query),
			// For arrays, we need to use Any() which uses reflection
			bolt.Any("results", []string{"result1", "result2", "result3"}),
		)
	})

	req := httptest.NewRequest("GET", "/search?page=1&limit=10&q=golang", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkFastFileUpload(b *testing.B) {
	app := bolt.New(bolt.WithDocs(false))
	app.Post("/upload", func(c *bolt.Context) error {
		buf := make([]byte, 10000)
		n, _ := c.Request.Body.Read(buf)

		return c.JSONFields(200,
			bolt.Bool("uploaded", true),
			bolt.Int("size", n),
			bolt.String("filename", "test.txt"),
		)
	})

	fileContent := strings.Repeat("This is test file content for upload benchmark.\n", 100)
	req := httptest.NewRequest("POST", "/upload", strings.NewReader(fileContent))
	req.Header.Set("Content-Type", "multipart/form-data")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}
