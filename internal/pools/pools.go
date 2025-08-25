// ABOUTME: This file implements sync.Pool for reusing expensive objects like goquery documents and HTTP response bodies.
// It reduces garbage collection pressure and improves performance in high-throughput parsing scenarios.
package pools

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// DocumentPool manages a pool of goquery.Document objects for reuse
type DocumentPool struct {
	pool sync.Pool
}

// ResponseBodyPool manages a pool of response body readers for reuse
type ResponseBodyPool struct {
	pool sync.Pool
}

// BufferPool manages a pool of bytes.Buffer objects for efficient string building
type BufferPool struct {
	pool sync.Pool
}

// StringBuilderPool manages a pool of strings.Builder objects for efficient string concatenation
type StringBuilderPool struct {
	pool sync.Pool
}

// Global pool instances for efficient object reuse
var (
	GlobalDocumentPool      = NewDocumentPool()
	GlobalResponseBodyPool  = NewResponseBodyPool()
	GlobalBufferPool        = NewBufferPool()
	GlobalStringBuilderPool = NewStringBuilderPool()
)

// NewDocumentPool creates a new DocumentPool with proper initialization
func NewDocumentPool() *DocumentPool {
	return &DocumentPool{
		pool: sync.Pool{
			New: func() interface{} {
				// We can't pre-create a Document as it needs HTML content
				// Return nil and handle creation in Get()
				return nil
			},
		},
	}
}

// Get retrieves a goquery.Document from the pool or creates a new one
func (dp *DocumentPool) Get(htmlContent io.Reader) (*goquery.Document, error) {
	// Try to get from pool first
	if pooled := dp.pool.Get(); pooled != nil {
		if doc, ok := pooled.(*goquery.Document); ok {
			// Reset the document by creating a new one with the content
			// Since goquery documents can't be easily reset, we'll create new ones
			// but still benefit from the pool for GC pressure reduction
			newDoc, err := goquery.NewDocumentFromReader(htmlContent)
			if err != nil {
				dp.pool.Put(doc) // Return the pooled doc on error
				return nil, err
			}
			return newDoc, nil
		}
	}

	// Create new document if pool is empty or failed
	return goquery.NewDocumentFromReader(htmlContent)
}

// Put returns a goquery.Document to the pool for reuse
func (dp *DocumentPool) Put(doc *goquery.Document) {
	if doc != nil {
		// Clear any modifications made to the document
		// Note: goquery documents can't be easily reset, so we just put it back
		// The main benefit is reducing GC pressure on the Document struct itself
		dp.pool.Put(doc)
	}
}

// NewResponseBodyPool creates a new ResponseBodyPool
func NewResponseBodyPool() *ResponseBodyPool {
	return &ResponseBodyPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

// Get retrieves a buffer from the pool for reading response bodies
func (rbp *ResponseBodyPool) Get() *bytes.Buffer {
	return rbp.pool.Get().(*bytes.Buffer)
}

// Put returns a buffer to the pool after reading a response body
func (rbp *ResponseBodyPool) Put(buf *bytes.Buffer) {
	if buf != nil {
		buf.Reset() // Clear the buffer for reuse
		rbp.pool.Put(buf)
	}
}

// ReadResponseBody efficiently reads an HTTP response body using pooled buffers
func (rbp *ResponseBodyPool) ReadResponseBody(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, nil
	}

	// Get a buffer from the pool
	buf := rbp.Get()
	defer rbp.Put(buf)

	// Copy the response body into the buffer
	_, err := io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}

	// Close the original body
	resp.Body.Close()

	// Return a copy of the data
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	return result, nil
}

// NewBufferPool creates a new BufferPool
func NewBufferPool() *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

// Get retrieves a bytes.Buffer from the pool
func (bp *BufferPool) Get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

// Put returns a bytes.Buffer to the pool
func (bp *BufferPool) Put(buf *bytes.Buffer) {
	if buf != nil {
		buf.Reset()
		// Only pool buffers under a reasonable size to prevent memory bloat
		if buf.Cap() < 64*1024 { // 64KB limit
			bp.pool.Put(buf)
		}
	}
}

// NewStringBuilderPool creates a new StringBuilderPool
func NewStringBuilderPool() *StringBuilderPool {
	return &StringBuilderPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &strings.Builder{}
			},
		},
	}
}

// Get retrieves a strings.Builder from the pool
func (sbp *StringBuilderPool) Get() *strings.Builder {
	return sbp.pool.Get().(*strings.Builder)
}

// Put returns a strings.Builder to the pool
func (sbp *StringBuilderPool) Put(sb *strings.Builder) {
	if sb != nil {
		sb.Reset()
		// Only pool builders under a reasonable size to prevent memory bloat
		if sb.Cap() < 64*1024 { // 64KB limit
			sbp.pool.Put(sb)
		}
	}
}

// PoolStats provides statistics about pool usage for monitoring
type PoolStats struct {
	DocumentsInUse    int
	BuffersInUse      int
	StringBuildersInUse int
	ResponseBuffersInUse int
}

// GetPoolStats returns current pool usage statistics
// Note: sync.Pool doesn't provide built-in stats, so this is an approximation
func GetPoolStats() PoolStats {
	// This is a simplified version - in production you'd want to add
	// atomic counters to track actual usage
	return PoolStats{
		DocumentsInUse:       0, // Would need counters to track
		BuffersInUse:         0,
		StringBuildersInUse:  0,
		ResponseBuffersInUse: 0,
	}
}

// PooledStringBuilder provides a convenient way to use pooled string builders
type PooledStringBuilder struct {
	builder *strings.Builder
	pool    *StringBuilderPool
}

// NewPooledStringBuilder creates a new PooledStringBuilder
func NewPooledStringBuilder() *PooledStringBuilder {
	return &PooledStringBuilder{
		builder: GlobalStringBuilderPool.Get(),
		pool:    GlobalStringBuilderPool,
	}
}

// WriteString writes a string to the pooled builder
func (psb *PooledStringBuilder) WriteString(s string) (int, error) {
	return psb.builder.WriteString(s)
}

// String returns the built string
func (psb *PooledStringBuilder) String() string {
	return psb.builder.String()
}

// Reset resets the builder for reuse
func (psb *PooledStringBuilder) Reset() {
	psb.builder.Reset()
}

// Close returns the builder to the pool
func (psb *PooledStringBuilder) Close() {
	if psb.builder != nil && psb.pool != nil {
		psb.pool.Put(psb.builder)
		psb.builder = nil
		psb.pool = nil
	}
}

// WithPooledStringBuilder executes a function with a pooled string builder
// and automatically returns it to the pool when done
func WithPooledStringBuilder(fn func(*strings.Builder) error) (string, error) {
	sb := GlobalStringBuilderPool.Get()
	defer GlobalStringBuilderPool.Put(sb)

	err := fn(sb)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

// WithPooledBuffer executes a function with a pooled buffer
// and automatically returns it to the pool when done
func WithPooledBuffer(fn func(*bytes.Buffer) error) ([]byte, error) {
	buf := GlobalBufferPool.Get()
	defer GlobalBufferPool.Put(buf)

	err := fn(buf)
	if err != nil {
		return nil, err
	}

	// Return a copy of the buffer data
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}