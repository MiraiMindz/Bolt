package main

import (  
    "log"  
    "bolt"  
)

type User struct {  
    Name string `json:"name"`  
}

func main() {  
    app := bolt.New(bolt.WithDevMode(true))  
      
    // API v1 group  
    app.Group("/api/v1", func(api *bolt.App) {  
          
        api.Get("/users", func(c *bolt.Context) error {  
            return c.JSON(200, []map[string]string{{"id": "1", "name": "User 1"}})  
        }).Doc(bolt.RouteDoc{Summary: "Get all users"})  
          
        api.PostJSON("/users", func(c *bolt.Context, user User) error {  
            return c.JSON(201, user)  
        }).Doc(bolt.RouteDoc{Summary: "Create a user", Request: User{}})  
    })  
      
    log.Fatal(app.Listen(":3000"))  
}