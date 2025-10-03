package bolt

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"
)

// Pre-serialized JSON error responses to avoid allocations in the error handler.
var (
	errNotFoundResponse      = []byte(`{"error":"Not Found"}`)
	errBadRequestResponse    = []byte(`{"error":"Bad Request"}`)
	errInternalServerResponse = []byte(`{"error":"Internal Server Error"}`)
)

// TypedHandlerInfo caches reflection information for typed handlers
type TypedHandlerInfo struct {
	HandlerValue reflect.Value
	BodyType     reflect.Type
}

// Global cache for typed handler reflection info to avoid repeated reflection
var (
	typedHandlerCache = make(map[uintptr]*TypedHandlerInfo)
	typedHandlerMutex sync.RWMutex
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
	pathBuilder  *pathBuilder       // For efficient path building
	server       *Server
	prefix       string
	parentGroup  *RouteGroup   // The group this App instance belongs to
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
		pathBuilder:  newPathBuilder(),
		parentGroup:  nil, // A new app has no parent
	}

	if config.EnablePooling {
		app.contextPool = NewContextPool()
		app.bufferPool = NewBufferPool()
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

// compileMiddleware pre-builds the middleware chain for a handler with fast path for single middleware.
func compileMiddleware(middleware []Middleware, finalHandler Handler) Handler {
	if len(middleware) == 0 {
		return finalHandler
	}

	// Fast path for single middleware (common case)
	if len(middleware) == 1 {
		return middleware[0](finalHandler)
	}

	// Chain building for multiple middleware
	h := finalHandler
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

// addRoute adds a route to the router with simplified middleware compilation.
func (a *App) addRoute(method HTTPMethod, path string, handler Handler) *ChainLink {
	// Use the path builder to avoid string concatenation allocations
	fullPath := a.pathBuilder.build(a.prefix, path)

	// Apply middleware compilation directly
	finalHandler := compileMiddleware(a.middleware, handler)
	a.router.AddRoute(method, fullPath, finalHandler)

	routeInfo := &RouteInfo{
		Method:  method,
		Path:    fullPath,
		Handler: handler, // Store original handler for documentation
		Group:   a.parentGroup,
	}
	a.routes = append(a.routes, *routeInfo)

	return &ChainLink{app: a, subject: routeInfo}
}

// Group creates a route group.
func (a *App) Group(prefix string, fn GroupFunc) *ChainLink {
	group := &RouteGroup{
		Prefix: a.pathBuilder.build(a.prefix, prefix),
	}

	subApp := &App{
		router:       a.router,
		config:       a.config,
		routes:       a.routes,
		middleware:   a.middleware,
		errorHandler: a.errorHandler,
		contextPool:  a.contextPool,
		bufferPool:   a.bufferPool,
		pathBuilder:  a.pathBuilder,
		prefix:       group.Prefix,
		parentGroup:  group,
	}

	fn(subApp)

	a.routes = subApp.routes

	return &ChainLink{app: a, subject: group}
}

// --- ChainLink Methods ---

// Doc can be called on a route or a group.
func (cl *ChainLink) Doc(doc RouteDoc) *ChainLink {
	switch v := cl.subject.(type) {
	case *RouteInfo:
		if len(cl.app.routes) > 0 {
			cl.app.routes[len(cl.app.routes)-1].Doc = doc
		}
	case *RouteGroup:
		v.Doc = doc
	}
	return cl
}

// Delegate methods for fluent API
func (cl *ChainLink) Get(path string, handler Handler) *ChainLink         { return cl.app.Get(path, handler) }
func (cl *ChainLink) Post(path string, handler Handler) *ChainLink        { return cl.app.Post(path, handler) }
func (cl *ChainLink) Put(path string, handler Handler) *ChainLink         { return cl.app.Put(path, handler) }
func (cl *ChainLink) Delete(path string, handler Handler) *ChainLink      { return cl.app.Delete(path, handler) }
func (cl *ChainLink) Patch(path string, handler Handler) *ChainLink       { return cl.app.Patch(path, handler) }
func (cl *ChainLink) Head(path string, handler Handler) *ChainLink        { return cl.app.Head(path, handler) }
func (cl *ChainLink) Options(path string, handler Handler) *ChainLink     { return cl.app.Options(path, handler) }
func (cl *ChainLink) PostJSON(path string, handler interface{}) *ChainLink { return cl.app.PostJSON(path, handler) }
func (cl *ChainLink) PutJSON(path string, handler interface{}) *ChainLink  { return cl.app.PutJSON(path, handler) }
func (cl *ChainLink) PatchJSON(path string, handler interface{}) *ChainLink{ return cl.app.PatchJSON(path, handler) }
func (cl *ChainLink) Group(prefix string, fn GroupFunc) *ChainLink         { return cl.app.Group(prefix, fn) }

// wrapTypedHandler uses cached reflection to handle strongly-typed JSON handlers efficiently.
func wrapTypedHandler(handler interface{}) Handler {
	// Pre-compute reflection info and cache it
	handlerPtr := reflect.ValueOf(handler).Pointer()

	typedHandlerMutex.RLock()
	info, exists := typedHandlerCache[handlerPtr]
	typedHandlerMutex.RUnlock()

	if !exists {
		// First time - compute and cache reflection info
		handlerValue := reflect.ValueOf(handler)
		handlerType := handlerValue.Type()
		if handlerType.NumIn() != 2 {
			panic("typed handler must accept exactly 2 parameters: (*Context, T)")
		}

		info = &TypedHandlerInfo{
			HandlerValue: handlerValue,
			BodyType:     handlerType.In(1),
		}

		typedHandlerMutex.Lock()
		typedHandlerCache[handlerPtr] = info
		typedHandlerMutex.Unlock()
	}

	return func(c *Context) error {
		// Fast path - use cached reflection info
		bodyValue := reflect.New(info.BodyType)
		bodyPtr := bodyValue.Interface()

		if err := c.BindJSON(bodyPtr); err != nil {
			return err
		}

		results := info.HandlerValue.Call([]reflect.Value{
			reflect.ValueOf(c),
			bodyValue.Elem(),
		})

		if len(results) > 0 && !results[0].IsNil() {
			return results[0].Interface().(error)
		}
		return nil
	}
}

// cachedHeaderWriter wraps http.ResponseWriter to return cached headers
type cachedHeaderWriter struct {
	http.ResponseWriter
	cachedHeaders http.Header
}

// Header returns the cached headers instead of calling the underlying writer's Header()
func (w *cachedHeaderWriter) Header() http.Header {
	return w.cachedHeaders
}

// ServeHTTP is the main entry point for handling requests - optimized for speed.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, params := a.router.GetValue(HTTPMethod(r.Method), r.URL.Path)
	if handler == nil {
		headers := w.Header()
		wrappedWriter := &cachedHeaderWriter{ResponseWriter: w, cachedHeaders: headers}
		c := &Context{Request: r, Response: wrappedWriter, app: a, headers: headers}
		a.errorHandler(c, ErrNotFound)
		return
	}

	var c *Context
	if a.config.EnablePooling && a.contextPool != nil {
		c = a.contextPool.Acquire()
		defer a.contextPool.Release(c)
	} else {
		c = &Context{}
	}

	// Cache headers once to avoid repeated Header() calls in middleware/handlers
	headers := w.Header()
	wrappedWriter := &cachedHeaderWriter{ResponseWriter: w, cachedHeaders: headers}

	// Initialize context with minimal overhead
	c.Request = r
	c.Response = wrappedWriter
	c.app = a
	c.params = params
	c.headers = headers

	// Fast path: Static routes with no query (most common case for middleware benchmarks)
	if params == nil && r.URL.RawQuery == "" {
		err := handler(c)
		if err != nil {
			a.errorHandler(c, err)
		}
		return
	}

	// Slow path: Dynamic routes or query parameters
	if r.URL.RawQuery != "" {
		if c.query == nil {
			c.query = make(QueryValues)
		}
		parsedQuery := r.URL.Query()
		for k, v := range parsedQuery {
			c.query[k] = v
		}
	}

	if err := handler(c); err != nil {
		a.errorHandler(c, err)
	}

	if params != nil && a.router.paramPool != nil {
		a.router.releaseParamMap(params)
	}
}

// DefaultErrorHandler provides a zero-allocation error handling mechanism.
func DefaultErrorHandler(c *Context, err error) {
	if err == nil {
		return
	}

	var code int
	var body []byte

	switch err {
	case ErrNotFound:
		code = http.StatusNotFound
		body = errNotFoundResponse
	case ErrBadRequest:
		code = http.StatusBadRequest
		body = errBadRequestResponse
	default:
		code = http.StatusInternalServerError
		body = errInternalServerResponse
	}

	// Avoid writing header twice
	if c.StatusCode == 0 {
		_ = c.Bytes(StatusCode(code), ContentTypeJSON, body)
	}
}

// SetErrorHandler sets a custom error handler for the application.
func (a *App) SetErrorHandler(handler ErrorHandler) *App {
	a.errorHandler = handler
	return a
}

// setupDocs configures documentation endpoints.
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

// Listen starts the HTTP server.
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
	a.setupDocs()
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return a.server.Serve(listener)
}

// handleShutdown gracefully shuts down the server.
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

// Shutdown provides a way to programmatically shut down the server.
func (a *App) Shutdown(ctx context.Context) error {
	if a.server == nil {
		return nil
	}
	return a.server.Shutdown(ctx)
}

