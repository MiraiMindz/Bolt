package bolt

import "unsafe"

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
	trees map[HTTPMethod]*Node
}

// NewRouter creates a new router.
func NewRouter() *Router {
	return &Router{
		trees: make(map[HTTPMethod]*Node),
	}
}

// unsafeBytesToString converts a byte slice to a string without allocation.
// IMPORTANT: This is unsafe and should only be used when the underlying
// []byte is not going to change for the lifetime of the string.
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

func addRoute(n *Node, path string, handler Handler, method HTTPMethod) {
	n.priority++

walk:
	for {
		// Find the longest common prefix.
		i := 0
		max := len(path)
		if len(n.path) < max {
			max = len(n.path)
		}
		for i < max && path[i] == n.path[i] {
			i++
		}

		// Case 1: Split the current node's path.
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

		// Case 2: The path continues, find or create the next child node.
		if i < len(path) {
			path = path[i:]
			c := path[0]

			// Check for existing children with the next character
			for j := 0; j < len(n.indices); j++ {
				if c == n.indices[j] {
					n = n.children[j]
					continue walk
				}
			}

			// No existing child, create a new one.
			if c == ':' || c == '*' {
				if len(n.children) > 0 {
					// Logic for conflicting wildcards/params can be complex.
				}
			}

			n.indices = append(n.indices, c)
			child := &Node{}
			n.children = append(n.children, child)
			n = child
			insertChild(n, path, handler, method)
			return
		}

		// Case 3: The path matches the node's path exactly. Set the handler.
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
		if c == ':' { // Parameter
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
		} else if c == '*' { // Wildcard
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
func (r *Router) GetValue(method HTTPMethod, path string) (Handler, ParamMap) {
	root := r.trees[method]
	if root == nil {
		return nil, nil
	}

	var handler Handler
	params := make(ParamMap)
	pathBytes := []byte(path) // Work with byte slice to avoid allocations

walk:
	for {
		prefixLen := len(root.path)
		if len(pathBytes) > prefixLen {
			if unsafeBytesToString(pathBytes[:prefixLen]) == root.path {
				pathBytes = pathBytes[prefixLen:]
				c := pathBytes[0]
				for i, index := range root.indices {
					if c == index {
						root = root.children[i]
						continue walk
					}
				}

				// Handle parameter
				if len(root.children) > 0 && root.children[0].nodeType == paramNode {
					root = root.children[0]
					end := 0
					for end < len(pathBytes) && pathBytes[end] != '/' {
						end++
					}
					params[root.paramName] = unsafeBytesToString(pathBytes[:end])
					if end < len(pathBytes) {
						pathBytes = pathBytes[end:]
						continue walk
					}
					handler = root.handlers[method]
					return handler, params
				}

				// Handle wildcard
				if len(root.children) > 0 && root.children[0].nodeType == wildcardNode {
					root = root.children[0]
					params[root.paramName] = unsafeBytesToString(pathBytes)
					handler = root.handlers[method]
					return handler, params
				}

				return nil, nil
			}
		} else if unsafeBytesToString(pathBytes) == root.path {
			if root.handlers != nil {
				handler = root.handlers[method]
				return handler, params
			}
		}

		return nil, nil
	}
}
