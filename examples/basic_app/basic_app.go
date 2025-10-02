package main

import (  
    "log"  
    "bolt"  
    "strconv"
)

type User struct {  
    ID    int    `json:"id"`  
    Name  string `json:"name"`  
    Email string `json:"email"`  
}

type CreateUserRequest struct {  
    Name  string `json:"name"`  
    Email string `json:"email"`  
}

func main() {
    app := bolt.New(  
        bolt.WithAPIInfo("User API", "User management API", "1.0.0"),  
        bolt.WithDevMode(true),  
    )  
      
    // The new fluent API in action.  
    app.Get("/", func(c *bolt.Context) error {  
        return c.JSON(200, map[string]string{  
            "message": "Hello World",  
        })  
    }).Doc(bolt.RouteDoc{  
        Summary:     "Welcome endpoint",  
        Description: "Returns a welcome message",  
    }).Get("/users/:id", func(c *bolt.Context) error {  
        id, _ := strconv.Atoi(c.Param("id"))
        user := User{ ID: id, Name: "John Doe", Email: "john@example.com" }  
        return c.JSON(200, user)  
    }).Doc(bolt.RouteDoc{  
        Summary:  "Get user by ID",  
        Response: User{},  
    }).PostJSON("/users", func(c *bolt.Context, req CreateUserRequest) error {  
        user := User{ ID: 2, Name: req.Name, Email: req.Email }  
        return c.JSON(201, user)  
    }).Doc(bolt.RouteDoc{  
        Summary:  "Create new user",  
        Request:  CreateUserRequest{},  
        Response: User{},  
    })  
      
    log.Fatal(app.Listen(":3000"))  
}