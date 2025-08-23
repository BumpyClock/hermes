// ABOUTME: Object pooling optimizations for high-throughput API scenarios
// ABOUTME: Provides memory-efficient reuse of Result structs and parser instances to reduce GC pressure

package parser

import (
	"sync"
	"time"
)

// CoreParser interface for basic parsing without circular dependencies
type CoreParser interface {
	Parse(url string, opts *ParserOptions) (*Result, error)
	ParseHTML(html, url string, opts *ParserOptions) (*Result, error)
}

// coreParserImpl implements CoreParser using a simple approach
type coreParserImpl struct {
	defaultOpts *ParserOptions
}

// NewCoreParser creates a basic parser implementation
func NewCoreParser(opts *ParserOptions) CoreParser {
	if opts == nil {
		opts = DefaultParserOptions()
	}
	return &coreParserImpl{defaultOpts: opts}
}

func (cp *coreParserImpl) Parse(url string, opts *ParserOptions) (*Result, error) {
	// For now, create a temporary Mercury instance to handle the parsing
	// This avoids the circular dependency while keeping the optimization framework
	tempParser := &Mercury{options: *cp.defaultOpts}
	if opts == nil {
		opts = cp.defaultOpts
	}
	return tempParser.parseWithoutOptimization(url, opts)
}

func (cp *coreParserImpl) ParseHTML(html, url string, opts *ParserOptions) (*Result, error) {
	// For now, create a temporary Mercury instance to handle the parsing
	tempParser := &Mercury{options: *cp.defaultOpts}
	if opts == nil {
		opts = cp.defaultOpts
	}
	return tempParser.parseHTMLWithoutOptimization(html, url, opts)
}

// ResultPool manages a pool of Result structs to reduce allocations
type ResultPool struct {
	pool sync.Pool
}

// NewResultPool creates a new Result object pool
func NewResultPool() *ResultPool {
	return &ResultPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Result{
					Extended: make(map[string]interface{}),
				}
			},
		},
	}
}

// Get retrieves a Result from the pool
func (rp *ResultPool) Get() *Result {
	result := rp.pool.Get().(*Result)
	// Reset the result to clean state
	rp.resetResult(result)
	return result
}

// Put returns a Result to the pool
func (rp *ResultPool) Put(result *Result) {
	if result != nil {
		rp.pool.Put(result)
	}
}

// resetResult clears a Result struct for reuse
func (rp *ResultPool) resetResult(result *Result) {
	result.Title = ""
	result.Content = ""
	result.Author = ""
	result.DatePublished = nil
	result.LeadImageURL = ""
	result.Dek = ""
	result.NextPageURL = ""
	result.URL = ""
	result.Domain = ""
	result.Excerpt = ""
	result.WordCount = 0
	result.Direction = ""
	result.TotalPages = 0
	result.RenderedPages = 0
	result.ExtractorUsed = ""
	result.Error = false
	result.Message = ""
	
	// Clear site metadata fields
	result.SiteName = ""
	result.SiteTitle = ""
	result.SiteImage = ""
	result.Favicon = ""
	result.Description = ""
	result.Language = ""
	
	// Clear the Extended map without reallocating
	for k := range result.Extended {
		delete(result.Extended, k)
	}
}


// BufferPool manages byte slices for content processing
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool creates a new buffer pool with specified initial size
func NewBufferPool(initialSize int) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, initialSize)
			},
		},
	}
}

// Get retrieves a buffer from the pool
func (bp *BufferPool) Get() []byte {
	return bp.pool.Get().([]byte)
}

// Put returns a buffer to the pool (resets length to 0)
func (bp *BufferPool) Put(buf []byte) {
	if buf != nil {
		// Reset length but keep capacity
		buf = buf[:0]
		bp.pool.Put(buf)
	}
}

// HighThroughputParser combines object pooling for maximum performance
type HighThroughputParser struct {
	resultPool  *ResultPool
	bufferPool  *BufferPool
	defaultOpts *ParserOptions
	stats       *PoolStats
	coreParser  CoreParser
}

// PoolStats tracks pool performance metrics
type PoolStats struct {
	mu                    sync.RWMutex
	TotalRequests        int64     `json:"total_requests"`
	PoolHits             int64     `json:"pool_hits"`
	PoolMisses           int64     `json:"pool_misses"`
	AverageProcessingTime float64   `json:"avg_processing_time_ms"`
	LastReset            time.Time `json:"last_reset"`
}

// NewHighThroughputParser creates an optimized parser for high-volume usage
func NewHighThroughputParser(opts *ParserOptions) *HighThroughputParser {
	if opts == nil {
		opts = DefaultParserOptions()
	}
	
	return &HighThroughputParser{
		resultPool:  NewResultPool(),
		bufferPool:  NewBufferPool(64 * 1024), // 64KB initial buffer size
		defaultOpts: opts,
		stats: &PoolStats{
			LastReset: time.Now(),
		},
		coreParser: NewCoreParser(opts),
	}
}

// Parse extracts content using object pooling for optimal performance
func (htp *HighThroughputParser) Parse(targetURL string, opts *ParserOptions) (*Result, error) {
	start := time.Now()
	defer func() {
		htp.updateStats(time.Since(start))
	}()
	
	// Use provided options or defaults
	if opts == nil {
		opts = htp.defaultOpts
	}
	
	// Parse using core parser
	parsedResult, err := htp.coreParser.Parse(targetURL, opts)
	if err != nil {
		return nil, err
	}
	
	// Get result from pool and copy data
	result := htp.resultPool.Get()
	htp.copyResult(parsedResult, result)
	
	return result, nil
}

// ParseHTML extracts content from HTML using object pooling
func (htp *HighThroughputParser) ParseHTML(html, targetURL string, opts *ParserOptions) (*Result, error) {
	start := time.Now()
	defer func() {
		htp.updateStats(time.Since(start))
	}()
	
	// Use provided options or defaults
	if opts == nil {
		opts = htp.defaultOpts
	}
	
	// Parse using core parser
	parsedResult, err := htp.coreParser.ParseHTML(html, targetURL, opts)
	if err != nil {
		return nil, err
	}
	
	// Get result from pool and copy data
	result := htp.resultPool.Get()
	htp.copyResult(parsedResult, result)
	
	return result, nil
}

// ParseBatch processes multiple URLs efficiently using object pooling
func (htp *HighThroughputParser) ParseBatch(urls []string, opts *ParserOptions) ([]*Result, []error) {
	if len(urls) == 0 {
		return nil, nil
	}
	
	results := make([]*Result, len(urls))
	errors := make([]error, 0, len(urls))
	
	// Use provided options or defaults
	if opts == nil {
		opts = htp.defaultOpts
	}
	
	for i, url := range urls {
		result, err := htp.Parse(url, opts)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		results[i] = result
	}
	
	if len(errors) > 0 {
		return results, errors
	}
	return results, nil
}

// ReturnResult returns a result to the pool (call this when done with the result)
func (htp *HighThroughputParser) ReturnResult(result *Result) {
	htp.resultPool.Put(result)
}

// GetStats returns current pool performance statistics
func (htp *HighThroughputParser) GetStats() *PoolStats {
	htp.stats.mu.RLock()
	defer htp.stats.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	return &PoolStats{
		TotalRequests:        htp.stats.TotalRequests,
		PoolHits:             htp.stats.PoolHits,
		PoolMisses:           htp.stats.PoolMisses,
		AverageProcessingTime: htp.stats.AverageProcessingTime,
		LastReset:            htp.stats.LastReset,
	}
}

// ResetStats resets performance statistics
func (htp *HighThroughputParser) ResetStats() {
	htp.stats.mu.Lock()
	defer htp.stats.mu.Unlock()
	
	htp.stats.TotalRequests = 0
	htp.stats.PoolHits = 0
	htp.stats.PoolMisses = 0
	htp.stats.AverageProcessingTime = 0
	htp.stats.LastReset = time.Now()
}

// copyResult efficiently copies data from source to destination Result
func (htp *HighThroughputParser) copyResult(src, dst *Result) {
	dst.Title = src.Title
	dst.Content = src.Content
	dst.Author = src.Author
	dst.DatePublished = src.DatePublished
	dst.LeadImageURL = src.LeadImageURL
	dst.Dek = src.Dek
	dst.NextPageURL = src.NextPageURL
	dst.URL = src.URL
	dst.Domain = src.Domain
	dst.Excerpt = src.Excerpt
	dst.WordCount = src.WordCount
	dst.Direction = src.Direction
	dst.TotalPages = src.TotalPages
	dst.RenderedPages = src.RenderedPages
	dst.ExtractorUsed = src.ExtractorUsed
	dst.Error = src.Error
	dst.Message = src.Message
	
	// Copy site metadata fields
	dst.SiteName = src.SiteName
	dst.SiteTitle = src.SiteTitle
	dst.SiteImage = src.SiteImage
	dst.Favicon = src.Favicon
	dst.Description = src.Description
	dst.Language = src.Language
	
	// Copy Extended map
	if src.Extended != nil {
		for k, v := range src.Extended {
			dst.Extended[k] = v
		}
	}
}

// updateStats updates performance metrics
func (htp *HighThroughputParser) updateStats(duration time.Duration) {
	htp.stats.mu.Lock()
	defer htp.stats.mu.Unlock()
	
	htp.stats.TotalRequests++
	
	// Update rolling average processing time
	durationMs := float64(duration.Nanoseconds()) / 1e6
	if htp.stats.TotalRequests == 1 {
		htp.stats.AverageProcessingTime = durationMs
	} else {
		// Rolling average with weight towards recent requests
		alpha := 0.1 // 10% weight to new value, 90% to existing average
		htp.stats.AverageProcessingTime = alpha*durationMs + (1-alpha)*htp.stats.AverageProcessingTime
	}
}

// Global high-throughput parser instance for convenience
var GlobalHTParser = NewHighThroughputParser(nil)

// Convenience functions for global parser
func Parse(url string, opts *ParserOptions) (*Result, error) {
	return GlobalHTParser.Parse(url, opts)
}

func ParseHTML(html, url string, opts *ParserOptions) (*Result, error) {
	return GlobalHTParser.ParseHTML(html, url, opts)
}

func ParseBatch(urls []string, opts *ParserOptions) ([]*Result, []error) {
	return GlobalHTParser.ParseBatch(urls, opts)
}

func ReturnResultToPool(result *Result) {
	GlobalHTParser.ReturnResult(result)
}

func ReturnResult(result *Result) {
	GlobalHTParser.ReturnResult(result)
}

func GetGlobalStats() *PoolStats {
	return GlobalHTParser.GetStats()
}