package bolt

// Node represents a radix tree node  
type Node struct {  
    path      string  
    indices   string  
    children  []*Node  
    handlers  map[HTTPMethod]Handler  
    priority  int  
    wildcard  bool  
    paramName string  
}

// Router implements a radix tree router  
type Router struct {  
    trees map[HTTPMethod]*Node  
}

// NewRouter creates a new router  
func NewRouter() *Router {  
    return &Router{  
        trees: make(map[HTTPMethod]*Node),  
    }  
}

// AddRoute adds a route to the router  
func (r *Router) AddRoute(method HTTPMethod, path string, handler Handler) {  
    if path[0] != '/' {  
        panic("path must begin with '/'")  
    }  
      
    root := r.trees[method]  
    if root == nil {  
        root = &Node{  
            handlers: make(map[HTTPMethod]Handler),  
        }  
        r.trees[method] = root  
    }  
      
    root.addRoute(path, method, handler)  
}

// addRoute adds a route to a node  
func (n *Node) addRoute(path string, method HTTPMethod, handler Handler) {  
    n.priority++  
      
    if n.path == "" && len(n.children) == 0 {  
        n.insertChild(path, method, handler)  
        return  
    }  
      
walk:  
    for {  
        i := longestCommonPrefix(path, n.path)  
          
        if i < len(n.path) {  
            child := &Node{  
                path:     n.path[i:],  
                indices:  n.indices,  
                children: n.children,  
                handlers: n.handlers,  
                priority: n.priority - 1,  
                wildcard: n.wildcard,  
            }  
              
            n.children = []*Node{child}  
            n.indices = string([]byte{n.path[i]})  
            n.path = path[:i]  
            n.handlers = make(map[HTTPMethod]Handler)  
            n.wildcard = false  
        }  
          
        if i < len(path) {  
            path = path[i:]  
            c := path[0]  
              
            for idx, index := range []byte(n.indices) {  
                if c == index {  
                    n.children[idx].priority++  
                    n = n.children[idx]  
                    continue walk  
                }  
            }  
              
            if c == ':' || c == '*' {  
                n.insertChild(path, method, handler)  
                return  
            }  
              
            n.indices += string([]byte{c})  
            child := &Node{  
                priority: 1,  
                handlers: make(map[HTTPMethod]Handler),  
            }  
            n.children = append(n.children, child)  
            n = child  
            n.insertChild(path, method, handler)  
            return  
        }  
          
        if n.handlers == nil {  
            n.handlers = make(map[HTTPMethod]Handler)  
        }  
        n.handlers[method] = handler  
        return  
    }  
}

// insertChild inserts a child node  
func (n *Node) insertChild(path string, method HTTPMethod, handler Handler) {  
    offset := 0  
      
    for i, max := 0, len(path); i < max; i++ {  
        c := path[i]  
        if c == ':' || c == '*' {  
            end := i + 1  
            for end < max && path[end] != '/' {  
                end++  
            }  
              
            if len(n.children) > 0 {  
                panic("wildcard conflicts with existing children")  
            }  
              
            paramName := path[i+1 : end]  
              
            if i > 0 {  
                n.path = path[:i]  
                path = path[i:]  
            }  
              
            child := &Node{  
                wildcard:  c == '*',  
                paramName: paramName,  
                priority:  1,  
                handlers:  make(map[HTTPMethod]Handler),  
            }  
              
            n.children = []*Node{child}  
            n.indices = string([]byte{c})  
            n = child  
              
            if c == '*' {  
                n.handlers[method] = handler  
                return  
            }  
              
            if end < max {  
                n.path = path[:end-i]  
                path = path[end-i:]  
                  
                child := &Node{  
                    priority: 1,  
                    handlers: make(map[HTTPMethod]Handler),  
                }  
                n.children = []*Node{child}  
                n = child  
            }  
              
            offset = i  
        }  
    }  
      
    n.path = path[offset:]  
    n.handlers[method] = handler  
}

// GetValue finds a handler and extracts parameters  
func (r *Router) GetValue(method HTTPMethod, path string) (Handler, ParamMap) {  
    root := r.trees[method]  
    if root == nil {  
        return nil, nil  
    }  
      
    params := make(ParamMap)  
    handler := root.getValue(method, path, params)  
      
    if handler == nil {  
        return nil, nil  
    }  
      
    return handler, params  
}

// getValue gets a handler from a node  
func (n *Node) getValue(method HTTPMethod, path string, params ParamMap) Handler {  
walk:  
    for {  
        prefix := n.path  
        if len(path) > len(prefix) {  
            if path[:len(prefix)] == prefix {  
                path = path[len(prefix):]  
                  
                if len(n.paramName) > 0 {  
                    end := 0  
                    for end < len(path) && path[end] != '/' {  
                        end++  
                    }  
                      
                    params[n.paramName] = path[:end]  
                      
                    if end < len(path) {  
                        if len(n.children) > 0 {  
                            path = path[end:]  
                            n = n.children[0]  
                            continue walk  
                        }  
                        return nil  
                    }  
                      
                    if handler := n.handlers[method]; handler != nil {  
                        return handler  
                    }  
                    return nil  
                }  
                  
                c := path[0]  
                for i, index := range []byte(n.indices) {  
                    if c == index {  
                        n = n.children[i]  
                        continue walk  
                    }  
                }  
                  
                if n.wildcard {  
                    params[n.paramName] = path  
                    return n.handlers[method]  
                }  
                  
                return nil  
            }  
        } else if path == prefix {  
            if handler := n.handlers[method]; handler != nil {  
                return handler  
            }  
        }  
          
        return nil  
    }  
}

// longestCommonPrefix finds the longest common prefix  
func longestCommonPrefix(a, b string) int {  
    i := 0  
    max := len(a)  
    if len(b) < max {  
        max = len(b)  
    }  
    for i < max && a[i] == b[i] {  
        i++  
    }  
    return i  
}