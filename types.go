package bolt

import (
	"net/http"
	"net/url"
	"time"
)

// Handler is the standard request handler function
type Handler func(*Context) error

// TypedHandler is a generic handler that receives parsed request body
type TypedHandler[T any] func(*Context, T) error

// Middleware wraps a handler to add functionality
type Middleware func(Handler) Handler

// ErrorHandler handles errors from handlers
type ErrorHandler func(*Context, error)

// GroupFunc defines routes within a route group
type GroupFunc func(*App)

// DocGenerator generates OpenAPI documentation
type DocGenerator func(*App) *OpenAPISpec

// HTTPMethod represents HTTP methods
type HTTPMethod string

const (
	MethodGet     HTTPMethod = "GET"
	MethodPost    HTTPMethod = "POST"
	MethodPut     HTTPMethod = "PUT"
	MethodDelete  HTTPMethod = "DELETE"
	MethodPatch   HTTPMethod = "PATCH"
	MethodHead    HTTPMethod = "HEAD"
	MethodOptions HTTPMethod = "OPTIONS"
)

// RouteGroup represents a group of routes with a shared prefix and documentation.
type RouteGroup struct {
	Prefix string
	Doc    RouteDoc
}

// RouteInfo stores metadata about a registered route.
type RouteInfo struct {
	Method  HTTPMethod
	Path    string
	Handler Handler
	Doc     RouteDoc
	Group   *RouteGroup // Link to the parent group
}

// ChainLink represents the current state of a fluent configuration chain.
// It holds a "subject" which is the focus of subsequent calls.
type ChainLink struct {
	app     *App
	subject interface{} // The subject can be *RouteInfo or *RouteGroup
}

// RouteDoc stores documentation metadata for a route
type RouteDoc struct {
	Summary     string
	Description string
	Tags        []string
	Request     interface{}
	Response    interface{}
}

// Config configures the entire application
type Config struct {
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	GenerateDocs      bool
	DocsConfig        DocsConfig
	EnablePooling     bool
	MaxPoolSize       int
	PreallocateRoutes int
	DevMode           bool
}

// DocsConfig configures automatic documentation
type DocsConfig struct {
	Enabled     bool
	SpecPath    string
	UIPath      string
	Title       string
	Description string
	Version     string
	Generator   DocGenerator
}

// Option is a functional option for Config
type Option func(*Config)

// ParamMap stores URL parameters
type ParamMap map[string]string

// QueryValues stores query parameters
type QueryValues url.Values

// StatusCode represents HTTP status codes
type StatusCode int

// ContentType represents content types
type ContentType string

const (
	ContentTypeJSON ContentType = "application/json; charset=utf-8"
	ContentTypeText ContentType = "text/plain; charset=utf-8"
	ContentTypeHTML ContentType = "text/html; charset=utf-8"
)

// ResponseWriter wraps http.ResponseWriter
type ResponseWriter http.ResponseWriter

// Request wraps http.Request
type Request = http.Request

// Server wraps http.Server
type Server = http.Server

// Listener wraps net.Listener
type Listener interface {
	Accept() (interface{}, error)
	Close() error
	Addr() interface{}
}