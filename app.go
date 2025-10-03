package bolt

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
)

// App is the main application. It can also represent a sub-application within a group.
type App struct {
	router       *Router
	config       Config
	routes       []RouteInfo
	middleware   []Middleware
	errorHandler ErrorHandler
	contextPool  *ContextPool
	bufferPool   *BufferPool
	server       *Server
	prefix       string
	parentGroup  *RouteGroup // The group this App instance belongs to
}

// New creates a new top-level App
func New(opts ...Option) *App {
	config := DefaultConfig()

	for _, opt := range opts {
		opt(&config)
	}

	app := &App{
		router:       NewRouter(),
		config:       config,
		routes:       make([]RouteInfo, 0, config.PreallocateRoutes),
		middleware:   make([]Middleware, 0, 8),
		errorHandler: DefaultErrorHandler,
		parentGroup:  nil, // A new app has no parent
	}

	if config.EnablePooling {
		app.contextPool = NewContextPool()
		app.bufferPool = NewBufferPool() // Updated to use the new BufferPool
	}

	return app
}

// Routes returns a copy of all registered routes.
func (a *App) Routes() []RouteInfo {
	return a.routes
}

// Use adds middleware to the application.
func (a *App) Use(middleware ...Middleware) *App {
	a.middleware = append(a.middleware, middleware...)
	return a
}

// Get registers a GET route.
func (a *App) Get(path string, handler Handler) *ChainLink {
	return a.addRoute(MethodGet, path, handler)
}

// Post registers a POST route.
func (a *App) Post(path string, handler Handler) *ChainLink {
	return a.addRoute(MethodPost, path, handler)
}

// Put registers a PUT route.
func (a *App) Put(path string, handler Handler) *ChainLink {
	return a.addRoute(MethodPut, path, handler)
}

// Delete registers a DELETE route.
func (a *App) Delete(path string, handler Handler) *ChainLink {
	return a.addRoute(MethodDelete, path, handler)
}

// Patch registers a PATCH route.
func (a *App) Patch(path string, handler Handler) *ChainLink {
	return a.addRoute(MethodPatch, path, handler)
}

// Head registers a HEAD route.
func (a *App) Head(path string, handler Handler) *ChainLink {
	return a.addRoute(MethodHead, path, handler)
}

// Options registers an OPTIONS route.
func (a *App) Options(path string, handler Handler) *ChainLink {
	return a.addRoute(MethodOptions, path, handler)
}

// PostJSON registers a POST route with automatic JSON parsing.
func (a *App) PostJSON(path string, handler interface{}) *ChainLink {
	wrappedHandler := wrapTypedHandler(handler)
	return a.addRoute(MethodPost, path, wrappedHandler)
}

// PutJSON registers a PUT route with automatic JSON parsing.
func (a *App) PutJSON(path string, handler interface{}) *ChainLink {
	wrappedHandler := wrapTypedHandler(handler)
	return a.addRoute(MethodPut, path, wrappedHandler)
}

// PatchJSON registers a PATCH route with automatic JSON parsing.
func (a *App) PatchJSON(path string, handler interface{}) *ChainLink {
	wrappedHandler := wrapTypedHandler(handler)
	return a.addRoute(MethodPatch, path, wrappedHandler)
}

// addRoute adds a route to the router and associates it with the current group context.
// It also pre-compiles the middleware chain for the route.
func (a *App) addRoute(method HTTPMethod, path string, handler Handler) *ChainLink {
	fullPath := a.prefix + path

	// Pre-compile the middleware chain for this route
	finalHandler := handler
	for i := len(a.middleware) - 1; i >= 0; i-- {
		finalHandler = a.middleware[i](finalHandler)
	}

	a.router.AddRoute(method, fullPath, finalHandler)

	routeInfo := &RouteInfo{
		Method:  method,
		Path:    fullPath,
		Handler: handler, // Store original handler for documentation/inspection
		Group:   a.parentGroup,
	}
	a.routes = append(a.routes, *routeInfo)

	return &ChainLink{app: a, subject: routeInfo}
}

// Group creates a route group. It returns a ChainLink so that methods like .Doc()
// can be called on the group itself.
func (a *App) Group(prefix string, fn GroupFunc) *ChainLink {
	group := &RouteGroup{
		Prefix: a.prefix + prefix,
	}

	// Create a sub-app that will be passed to the user's function.
	// This sub-app carries the context of the group.
	subApp := &App{
		router:       a.router,
		config:       a.config,
		routes:       a.routes,     // Use the same slice to collect all routes
		middleware:   a.middleware, // Inherit middleware
		errorHandler: a.errorHandler,
		contextPool:  a.contextPool,
		bufferPool:   a.bufferPool,
		prefix:       group.Prefix, // New prefix for routes defined inside
		parentGroup:  group,        // All routes in here will belong to this group
	}

	// Execute the user's function, defining all the routes within the group.
	fn(subApp)

	// Update the parent's route slice, ensuring it has all the new routes.
	a.routes = subApp.routes

	// Return a chain link focused on the group object itself.
	return &ChainLink{app: a, subject: group}
}

// --- ChainLink Methods ---

// Doc can be called on a route or a group.
func (cl *ChainLink) Doc(doc RouteDoc) *ChainLink {
	switch v := cl.subject.(type) {
	case *RouteInfo:
		// When called after a route, update the last added route's doc.
		if len(cl.app.routes) > 0 {
			cl.app.routes[len(cl.app.routes)-1].Doc = doc
		}
	case *RouteGroup:
		// When called after a group, update the group's doc.
		v.Doc = doc
	}
	return cl
}

// Get, Post, etc., on a ChainLink delegate to the app, creating a new route.
func (cl *ChainLink) Get(path string, handler Handler) *ChainLink { return cl.app.Get(path, handler) }
func (cl *ChainLink) Post(path string, handler Handler) *ChainLink { return cl.app.Post(path, handler) }
func (cl *ChainLink) Put(path string, handler Handler) *ChainLink { return cl.app.Put(path, handler) }
func (cl *ChainLink) Delete(path string, handler Handler) *ChainLink {
	return cl.app.Delete(path, handler)
}
func (cl *ChainLink) Patch(path string, handler Handler) *ChainLink { return cl.app.Patch(path, handler) }
func (cl *ChainLink) Head(path string, handler Handler) *ChainLink { return cl.app.Head(path, handler) }
func (cl *ChainLink) Options(path string, handler Handler) *ChainLink {
	return cl.app.Options(path, handler)
}
func (cl *ChainLink) PostJSON(path string, handler interface{}) *ChainLink {
	return cl.app.PostJSON(path, handler)
}
func (cl *ChainLink) PutJSON(path string, handler interface{}) *ChainLink {
	return cl.app.PutJSON(path, handler)
}
func (cl *ChainLink) PatchJSON(path string, handler interface{}) *ChainLink {
	return cl.app.PatchJSON(path, handler)
}
func (cl *ChainLink) Group(prefix string, fn GroupFunc) *ChainLink { return cl.app.Group(prefix, fn) }

// wrapTypedHandler remains the same
func wrapTypedHandler(handler interface{}) Handler {
	return func(c *Context) error {
		handlerValue := reflect.ValueOf(handler)
		handlerType := handlerValue.Type()
		if handlerType.NumIn() != 2 {
			panic("typed handler must accept exactly 2 parameters: (*Context, T)")
		}
		bodyType := handlerType.In(1)
		bodyValue := reflect.New(bodyType)
		bodyPtr := bodyValue.Interface()
		if err := c.BindJSON(bodyPtr); err != nil {
			return err
		}
		results := handlerValue.Call([]reflect.Value{
			reflect.ValueOf(c),
			bodyValue.Elem(),
		})
		if len(results) > 0 && !results[0].IsNil() {
			return results[0].Interface().(error)
		}
		return nil
	}
}

// ServeHTTP is the main entry point for handling requests.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var c *Context
	if a.contextPool != nil {
		c = a.contextPool.Acquire()
		defer a.contextPool.Release(c)
	} else {
		c = &Context{
			params: make(ParamMap, 8),
			query:  make(QueryValues, 8),
		}
	}
	c.Request = r
	c.Response = w
	c.app = a

	// Eagerly parse query parameters and populate the pooled map.
	// r.URL.Query() will allocate, but we reuse our `c.query` map.
	parsedQuery := r.URL.Query()
	for k, v := range parsedQuery {
		c.query[k] = v
	}

	handler, params := a.router.GetValue(HTTPMethod(r.Method), r.URL.Path)
	if handler == nil {
		a.errorHandler(c, ErrNotFound)
		return
	}
	c.params = params

	// The middleware chain is pre-compiled, so we can call the handler directly.
	if err := handler(c); err != nil {
		a.errorHandler(c, err)
	}
}

// DefaultErrorHandler remains the same
func DefaultErrorHandler(c *Context, err error) {
	if err == nil {
		return
	}
	code := http.StatusInternalServerError
	message := "Internal Server Error"
	switch err {
	case ErrNotFound:
		code = http.StatusNotFound
		message = "Not Found"
	case ErrBadRequest:
		code = http.StatusBadRequest
		message = "Bad Request"
	}
	// We check if the response has been written to already
	if c.Response.Header().Get("Content-Type") == "" {
		c.JSON(code, map[string]string{"error": message})
	}
}

// SetErrorHandler sets a custom error handler for the application.
func (a *App) SetErrorHandler(handler ErrorHandler) *App {
	a.errorHandler = handler
	return a
}

// Listen, setupDocs, handleShutdown, and Shutdown remain the same
func (a *App) setupDocs() {
	if !a.config.GenerateDocs || !a.config.DocsConfig.Enabled {
		return
	}
	specPath := a.config.DocsConfig.SpecPath
	a.Get(specPath, func(c *Context) error {
		spec := a.GenerateDocs()
		return c.JSON(http.StatusOK, spec)
	})
	uiPath := a.config.DocsConfig.UIPath
	a.Get(uiPath, ServeSwaggerUI(specPath))
	if a.config.DevMode && a.server != nil && a.server.Addr != "" {
		log.Printf("📚 API Documentation available at: http://localhost%s%s", a.server.Addr, uiPath)
	}
}

func (a *App) Listen(addr string) error {
	a.server = &http.Server{
		Addr:         addr,
		Handler:      a,
		ReadTimeout:  a.config.ReadTimeout,
		WriteTimeout: a.config.WriteTimeout,
		IdleTimeout:  a.config.IdleTimeout,
	}
	go a.handleShutdown()
	if a.config.DevMode {
		log.Printf("🚀 Server starting on http://localhost%s", addr)
	}
	a.setupDocs() // Call after server is configured to get Addr
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return a.server.Serve(listener)
}

func (a *App) handleShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("✅ Server stopped")
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.server == nil {
		return nil
	}
	return a.server.Shutdown(ctx)
}

