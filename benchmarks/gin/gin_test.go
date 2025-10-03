package benchmarks

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func BenchmarkStaticRoute(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})

	req := httptest.NewRequest("GET", "/hello", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkDynamicRoute(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.GET("/user/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, map[string]string{"id": id})
	})

	req := httptest.NewRequest("GET", "/user/123", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkTypedJSON(b *testing.B) {
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.POST("/users", func(c *gin.Context) {
		var user User
		c.BindJSON(&user)
		c.JSON(201, user)
	})

	body := `{"name":"John","email":"john@example.com"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkMiddleware(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Header("X-Test-Header", "test")
		c.Next()
	})
	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})

	req := httptest.NewRequest("GET", "/hello", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
