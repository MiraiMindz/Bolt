package main

import (  
    "log"  
    "time"  
    "bolt"  
)

func Logger() bolt.Middleware {  
    return func(next bolt.Handler) bolt.Handler {  
        return func(c *bolt.Context) error {  
            start := time.Now()  
            err := next(c)  
            duration := time.Since(start)  
            log.Printf("[%s] %s %d (%v)", c.Request.Method, c.Request.URL.Path, c.StatusCode, duration)  
            return err  
        }  
    }  
}

func main() {  
    app := bolt.New(bolt.WithDevMode(true))  
      
    // Global middleware is still applied to the app instance.  
    app.Use(Logger())  
      
    app.Get("/", func(c *bolt.Context) error {  
        return c.JSON(200, map[string]string{"message": "Public"})  
    }).Doc(bolt.RouteDoc{Summary: "Public endpoint"})  
      
    log.Fatal(app.Listen(":3000"))  
}