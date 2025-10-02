package benchmarks_bolt_only

import (  
    "net/http/httptest"  
    "strings"  
    "testing"  
    "bolt"  
)

func BenchmarkStaticRoute(b *testing.B) {  
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

func BenchmarkDynamicRoute(b *testing.B) {  
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

func BenchmarkTypedJSON(b *testing.B) {  
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
