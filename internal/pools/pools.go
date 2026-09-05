package pools

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

// BufferPool manages reusable bytes.Buffer values.
type BufferPool struct {
	pool sync.Pool
}

// Global pool instances for efficient object reuse.
var (
	GlobalBufferPool       = NewBufferPool()
	GlobalResponseBodyPool = GlobalBufferPool
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
