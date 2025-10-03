package benchmarks

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func BenchmarkStaticRoute(b *testing.B) {
	e := echo.New()
	e.GET("/hello", func(c echo.Context) error {
		return c.String(200, "Hello, World!")
	})

	req := httptest.NewRequest("GET", "/hello", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
	}
}

func BenchmarkDynamicRoute(b *testing.B) {
	e := echo.New()
	e.GET("/user/:id", func(c echo.Context) error {
		id := c.Param("id")
		return c.JSON(200, map[string]string{"id": id})
	})

	req := httptest.NewRequest("GET", "/user/123", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
	}
}

func BenchmarkTypedJSON(b *testing.B) {
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	e := echo.New()
	e.POST("/users", func(c echo.Context) error {
		var user User
		c.Bind(&user)
		return c.JSON(201, user)
	})

	body := `{"name":"John","email":"john@example.com"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
	}
}

func BenchmarkMiddleware(b *testing.B) {
	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-Test-Header", "test")
			return next(c)
		}
	})
	e.GET("/hello", func(c echo.Context) error {
		return c.String(200, "Hello, World!")
	})

	req := httptest.NewRequest("GET", "/hello", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
	}
}
