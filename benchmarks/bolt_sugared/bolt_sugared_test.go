package benchmarks

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bolt"
)

// Sugared API benchmarks - using the ergonomic Context API

func BenchmarkSugaredStaticRoute(b *testing.B) {
	app := bolt.New(bolt.WithDocs(false))
	app.Get("/hello", func(c *bolt.Context) error {
		return c.String(200, "Hello, World!")
	})

	req := httptest.NewRequest("GET", "/hello", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkSugaredDynamicRoute(b *testing.B) {
	app := bolt.New(bolt.WithDocs(false))
	app.Get("/user/:id", func(c *bolt.Context) error {
		id := c.Param("id")
		return c.JSON(200, map[string]string{"id": id})
	})

	req := httptest.NewRequest("GET", "/user/123", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkSugaredTypedJSON(b *testing.B) {
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	app := bolt.New(bolt.WithDocs(false))
	app.PostJSON("/users", func(c *bolt.Context, user User) error {
		return c.JSON(201, user)
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

func BenchmarkSugaredMiddleware(b *testing.B) {
	app := bolt.New(bolt.WithDocs(false))
	app.Use(func(next bolt.Handler) bolt.Handler {
		return func(c *bolt.Context) error {
			c.Response.Header().Set("X-Test-Header", "test")
			return next(c)
		}
	})
	app.Get("/hello", func(c *bolt.Context) error {
		return c.String(200, "Hello, World!")
	})

	req := httptest.NewRequest("GET", "/hello", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

// Extended benchmarks for Sugared API
func BenchmarkSugaredComplexRouting(b *testing.B) {
	app := bolt.New(bolt.WithDocs(false))

	app.Get("/api/v1/users", func(c *bolt.Context) error {
		return c.JSON(200, []map[string]interface{}{
			{"id": 1, "name": "User1"},
			{"id": 2, "name": "User2"},
		})
	})

	app.Get("/api/v1/users/:id", func(c *bolt.Context) error {
		id := c.Param("id")
		return c.JSON(200, map[string]interface{}{"id": id, "name": "User" + id})
	})

	type NewUser struct {
		Name string `json:"name"`
	}
	app.PostJSON("/api/v1/users", func(c *bolt.Context, user NewUser) error {
		return c.JSON(201, map[string]interface{}{"id": 456, "name": user.Name})
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

func BenchmarkSugaredLargeJSON(b *testing.B) {
	type LargeObject struct {
		ID          int                    `json:"id"`
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Tags        []string               `json:"tags"`
		Metadata    map[string]interface{} `json:"metadata"`
		Items       []map[string]string    `json:"items"`
	}

	app := bolt.New(bolt.WithDocs(false))
	app.PostJSON("/large", func(c *bolt.Context, obj LargeObject) error {
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

func BenchmarkSugaredQueryParameters(b *testing.B) {
	app := bolt.New(bolt.WithDocs(false))
	app.Get("/search", func(c *bolt.Context) error {
		page := c.QueryInt("page", 1)
		limit := c.QueryInt("limit", 10)
		query := c.Query("q")

		return c.JSON(200, map[string]interface{}{
			"page":    page,
			"limit":   limit,
			"query":   query,
			"results": []string{"result1", "result2", "result3"},
		})
	})

	req := httptest.NewRequest("GET", "/search?page=1&limit=10&q=golang", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkSugaredFileUpload(b *testing.B) {
	app := bolt.New(bolt.WithDocs(false))
	app.Post("/upload", func(c *bolt.Context) error {
		buf := make([]byte, 10000)
		n, _ := c.Request.Body.Read(buf)

		return c.JSON(200, map[string]interface{}{
			"uploaded": true,
			"size":     n,
			"filename": "test.txt",
		})
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
