package bolt

import (
	"io"
	"sync"

	json "github.com/goccy/go-json"
)

// StreamingJSONPool manages reusable JSON encoders and decoders
type StreamingJSONPool struct {
	encoderPool sync.Pool
	decoderPool sync.Pool
}

// NewStreamingJSONPool creates a new streaming JSON pool
func NewStreamingJSONPool() *StreamingJSONPool {
	return &StreamingJSONPool{
		encoderPool: sync.Pool{
			New: func() interface{} {
				return json.NewEncoder(nil)
			},
		},
		decoderPool: sync.Pool{
			New: func() interface{} {
				return json.NewDecoder(nil)
			},
		},
	}
}

// AcquireEncoder gets a JSON encoder from the pool
func (p *StreamingJSONPool) AcquireEncoder(w io.Writer) *json.Encoder {
	// Since go-json doesn't have Reset, we create a new encoder each time
	// but this is still faster than alternatives due to go-json's speed
	return json.NewEncoder(w)
}

// ReleaseEncoder is a no-op for go-json compatibility
func (p *StreamingJSONPool) ReleaseEncoder(encoder *json.Encoder) {
	// No-op since we can't reuse go-json encoders
}

// AcquireDecoder gets a JSON decoder from the pool
func (p *StreamingJSONPool) AcquireDecoder(r io.Reader) *json.Decoder {
	// Since go-json doesn't have Reset, we create a new decoder each time
	return json.NewDecoder(r)
}

// ReleaseDecoder is a no-op for go-json compatibility
func (p *StreamingJSONPool) ReleaseDecoder(decoder *json.Decoder) {
	// No-op since we can't reuse go-json decoders
}

// ByteSlicePool manages reusable byte slices for JSON operations
type ByteSlicePool struct {
	smallPool  sync.Pool // 512B
	mediumPool sync.Pool // 2KB
	largePool  sync.Pool // 8KB
}

// NewByteSlicePool creates a new byte slice pool
func NewByteSlicePool() *ByteSlicePool {
	return &ByteSlicePool{
		smallPool: sync.Pool{
			New: func() interface{} {
				slice := make([]byte, 0, 512)
				return &slice
			},
		},
		mediumPool: sync.Pool{
			New: func() interface{} {
				slice := make([]byte, 0, 2048)
				return &slice
			},
		},
		largePool: sync.Pool{
			New: func() interface{} {
				slice := make([]byte, 0, 8192)
				return &slice
			},
		},
	}
}

// Acquire gets a byte slice based on expected size
func (p *ByteSlicePool) Acquire(expectedSize int) *[]byte {
	switch {
	case expectedSize <= 512:
		return p.smallPool.Get().(*[]byte)
	case expectedSize <= 2048:
		return p.mediumPool.Get().(*[]byte)
	default:
		return p.largePool.Get().(*[]byte)
	}
}

// Release returns a byte slice to the appropriate pool
func (p *ByteSlicePool) Release(slice *[]byte) {
	if slice == nil {
		return
	}

	// Reset the slice but keep capacity
	*slice = (*slice)[:0]

	capacity := cap(*slice)
	switch {
	case capacity <= 512:
		p.smallPool.Put(slice)
	case capacity <= 2048:
		p.mediumPool.Put(slice)
	case capacity <= 8192:
		p.largePool.Put(slice)
	}
	// Large slices are not pooled to prevent memory bloat
}