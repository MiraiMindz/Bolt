package bolt

import (
	"sync"
	"unsafe"
)

// nodeType represents the type of a node in the radix tree.
type nodeType uint8

const (
	staticNode   nodeType = iota // A static path segment, e.g., "/users"
	paramNode                    // A parameter, e.g., "/:id"
	wildcardNode                 // A catch-all wildcard, e.g., "/*path"
)

// Node represents a node in the radix tree.
type Node struct {
	path      string
	indices   []byte // Changed from string to []byte
	children  []*Node
	handlers  map[HTTPMethod]Handler
	nodeType  nodeType
	priority  uint32
	paramName string
}

// Router implements a high-performance radix tree router.
type Router struct {
	trees     map[HTTPMethod]*Node
	paramPool *sync.Pool // Pool for ParamMap to reduce allocations
}

// NewRouter creates a new router.
func NewRouter() *Router {
	return &Router{
		trees: make(map[HTTPMethod]*Node),
		paramPool: &sync.Pool{
			New: func() interface{} {
				// Initialize with a default capacity
				return make(ParamMap, DefaultParamsSize)
			},
		},
	}
}

// acquireParamMap gets a ParamMap from the pool.
func (r *Router) acquireParamMap() ParamMap {
	return r.paramPool.Get().(ParamMap)
}

// releaseParamMap returns a ParamMap to the pool after clearing it.
func (r *Router) releaseParamMap(p ParamMap) {
	if p == nil {
		return
	}
	// If the map grew too large, don't return it to the pool.
	// Let the GC handle it and create a new, smaller one next time.
	if len(p) > MaxParams {
		return
	}
	for k := range p {
		delete(p, k)
	}
	r.paramPool.Put(p)
}

// unsafeBytesToString converts a byte slice to a string without allocation.
func unsafeBytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// AddRoute adds a new route to the router.
func (r *Router) AddRoute(method HTTPMethod, path string, handler Handler) {
	if path[0] != '/' {
		panic("path must begin with '/'")
	}

	root := r.trees[method]
	if root == nil {
		root = &Node{}
		r.trees[method] = root
	}

	addRoute(root, path, handler, method)
}

// (addRoute and insertChild logic remains the same as it's complex and not the primary bottleneck)
func addRoute(n *Node, path string, handler Handler, method HTTPMethod) {
	n.priority++
walk:
	for {
		i := 0
		max := len(path)
		if len(n.path) < max {
			max = len(n.path)
		}
		for i < max && path[i] == n.path[i] {
			i++
		}
		if i < len(n.path) {
			child := &Node{
				path:      n.path[i:],
				indices:   n.indices,
				children:  n.children,
				handlers:  n.handlers,
				priority:  n.priority - 1,
				nodeType:  n.nodeType,
				paramName: n.paramName,
			}
			n.children = []*Node{child}
			n.indices = []byte{n.path[i]}
			n.path = path[:i]
			n.handlers = nil
			n.paramName = ""
			n.nodeType = staticNode
		}
		if i < len(path) {
			path = path[i:]
			c := path[0]
			for j := 0; j < len(n.indices); j++ {
				if c == n.indices[j] {
					n = n.children[j]
					continue walk
				}
			}
			n.indices = append(n.indices, c)
			child := &Node{}
			n.children = append(n.children, child)
			n = child
			insertChild(n, path, handler, method)
			return
		}
		if n.handlers == nil {
			n.handlers = make(map[HTTPMethod]Handler)
		}
		n.handlers[method] = handler
		return
	}
}

func insertChild(n *Node, path string, handler Handler, method HTTPMethod) {
	var offset int
	for offset = 0; offset < len(path); offset++ {
		c := path[offset]
		if c == ':' || c == '*' {
			break
		}
	}
	if offset < len(path) {
		c := path[offset]
		if c == ':' {
			end := offset + 1
			for end < len(path) && path[end] != '/' {
				end++
			}
			n.path = path[:end]
			n.paramName = path[offset+1 : end]
			n.nodeType = paramNode
			if end < len(path) {
				child := &Node{}
				n.indices = []byte{path[end]}
				n.children = []*Node{child}
				addRoute(child, path[end:], handler, method)
				return
			}
		} else if c == '*' {
			n.path = path
			n.paramName = path[offset+1:]
			n.nodeType = wildcardNode
		}
	} else {
		n.path = path
	}
	if n.handlers == nil {
		n.handlers = make(map[HTTPMethod]Handler)
	}
	n.handlers[method] = handler
}


// GetValue finds a handler and extracts parameters for a given path.
// It now uses a pool for the parameters map to reduce allocations.
func (r *Router) GetValue(method HTTPMethod, path string) (Handler, ParamMap) {
	root := r.trees[method]
	if root == nil {
		return nil, nil
	}

	var handler Handler
	var params ParamMap = nil // Initialize to nil
	pathBytes := []byte(path)

walk:
	for {
		prefixLen := len(root.path)
		if len(pathBytes) >= prefixLen && unsafeBytesToString(pathBytes[:prefixLen]) == root.path {
			pathBytes = pathBytes[prefixLen:]

			if len(pathBytes) == 0 {
				// Path matches exactly
				if root.handlers != nil {
					handler = root.handlers[method]
				}
				return handler, params
			}

			// Path continues, look for a matching child
			c := pathBytes[0]
			for i, index := range root.indices {
				if c == index {
					root = root.children[i]
					continue walk
				}
			}

			// No static child found, check for param/wildcard
			if len(root.children) > 0 {
				child := root.children[0]
				if child.nodeType == paramNode {
					root = child
					end := 0
					for end < len(pathBytes) && pathBytes[end] != '/' {
						end++
					}

					if params == nil {
						params = r.acquireParamMap()
					}
					params[root.paramName] = unsafeBytesToString(pathBytes[:end])
					pathBytes = pathBytes[end:]
					continue walk
				}

				if child.nodeType == wildcardNode {
					root = child
					if params == nil {
						params = r.acquireParamMap()
					}
					params[root.paramName] = unsafeBytesToString(pathBytes)
					handler = root.handlers[method]
					return handler, params
				}
			}

			// No route matched
			if params != nil {
				r.releaseParamMap(params)
			}
			return nil, nil
		}
		
		// This node's path does not match
		if params != nil {
			r.releaseParamMap(params)
		}
		return nil, nil
	}
}
