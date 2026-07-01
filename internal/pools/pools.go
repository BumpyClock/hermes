// ABOUTME: Shared sync.Pool helpers for reusable buffers and string builders.
// ABOUTME: Keeps pooling at the primitive level so callers do not maintain duplicate wrappers.
package pools

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"
)

// BufferPool manages reusable bytes.Buffer values.
type BufferPool struct {
	pool sync.Pool
}

// StringBuilderPool manages reusable strings.Builder values.
type StringBuilderPool struct {
	pool sync.Pool
}

// Global pool instances for efficient object reuse.
var (
	GlobalBufferPool        = NewBufferPool()
	GlobalResponseBodyPool  = GlobalBufferPool
	GlobalStringBuilderPool = NewStringBuilderPool()
)

// NewBufferPool creates a new BufferPool.
func NewBufferPool() *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

// Get retrieves a bytes.Buffer from the pool.
func (bp *BufferPool) Get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

// Put returns a bytes.Buffer to the pool.
func (bp *BufferPool) Put(buf *bytes.Buffer) {
	if buf == nil {
		return
	}

	buf.Reset()
	if buf.Cap() < 64*1024 {
		bp.pool.Put(buf)
	}
}

// ReadResponseBody reads and closes an HTTP response body using a pooled buffer.
func (bp *BufferPool) ReadResponseBody(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, nil
	}
	defer func() { _ = resp.Body.Close() }()

	buf := bp.Get()
	defer bp.Put(buf)

	if _, err := io.Copy(buf, resp.Body); err != nil {
		return nil, err
	}

	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

// NewStringBuilderPool creates a new StringBuilderPool.
func NewStringBuilderPool() *StringBuilderPool {
	return &StringBuilderPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &strings.Builder{}
			},
		},
	}
}

// Get retrieves a strings.Builder from the pool.
func (sbp *StringBuilderPool) Get() *strings.Builder {
	return sbp.pool.Get().(*strings.Builder)
}

// Put returns a strings.Builder to the pool.
func (sbp *StringBuilderPool) Put(sb *strings.Builder) {
	if sb == nil {
		return
	}

	sb.Reset()
	if sb.Cap() < 64*1024 {
		sbp.pool.Put(sb)
	}
}

// WithPooledStringBuilder executes a function with a pooled string builder.
func WithPooledStringBuilder(fn func(*strings.Builder) error) (string, error) {
	sb := GlobalStringBuilderPool.Get()
	defer GlobalStringBuilderPool.Put(sb)

	if err := fn(sb); err != nil {
		return "", err
	}

	return sb.String(), nil
}

// WithPooledBuffer executes a function with a pooled buffer.
func WithPooledBuffer(fn func(*bytes.Buffer) error) ([]byte, error) {
	buf := GlobalBufferPool.Get()
	defer GlobalBufferPool.Put(buf)

	if err := fn(buf); err != nil {
		return nil, err
	}

	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}
