package bolt

import "sync"

// pathBuilder helps build paths efficiently, avoiding string concatenations.
type pathBuilder struct {
	buf []byte
}

// newPathBuilder creates a new path builder.
func newPathBuilder() *pathBuilder {
	return &pathBuilder{
		buf: make([]byte, 0, 128), // Pre-allocate a reasonable size
	}
}

// build combines a prefix and a path into a single string.
func (pb *pathBuilder) build(prefix, path string) string {
	pb.buf = pb.buf[:0] // Reset buffer without reallocating
	pb.buf = append(pb.buf, prefix...)
	pb.buf = append(pb.buf, path...)
	return string(pb.buf)
}

// --- String Interning for Path Segments ---

var (
	pathIntern   = make(map[string]string)
	pathInternMu sync.RWMutex
)

// internPath returns a shared string for a given path segment,
// reducing memory usage for common route parts.
func internPath(path string) string {
	pathInternMu.RLock()
	if interned, ok := pathIntern[path]; ok {
		pathInternMu.RUnlock()
		return interned
	}
	pathInternMu.RUnlock()

	pathInternMu.Lock()
	// Double-check in case another goroutine just added it
	if interned, ok := pathIntern[path]; ok {
		pathInternMu.Unlock()
		return interned
	}
	pathIntern[path] = path
	pathInternMu.Unlock()
	return path
}
