// ABOUTME: High-performance worker pool for concurrent URL processing
// ABOUTME: Enables parsing hundreds of URLs simultaneously with controlled concurrency and resource management

package parser

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// WorkerPoolConfig configures the worker pool behavior
type WorkerPoolConfig struct {
	NumWorkers      int           // Number of worker goroutines
	QueueSize       int           // Size of job queue buffer
	JobTimeout      time.Duration // Timeout for individual jobs
	ShutdownTimeout time.Duration // Time to wait for graceful shutdown
	EnableMetrics   bool          // Whether to collect detailed metrics
	MaxRetries      int           // Maximum retry attempts for failed jobs
}

// BatchJob represents a single URL parsing job
type BatchJob struct {
	ID       string         // Unique job identifier
	URL      string         // URL to parse
	Options  *ParserOptions // Parser configuration
	Priority int            // Job priority (lower = higher priority)
	Context  context.Context // Job-specific context
	Retries  int            // Current retry count
}

// BatchResult represents the result of a batch job
type BatchResult struct {
	JobID     string        // Job identifier
	URL       string        // URL that was parsed
	Result    *Result       // Parsed result (nil if error)
	Error     error         // Error if parsing failed
	Duration  time.Duration // Time taken to process
	WorkerID  int           // ID of worker that processed this job
	Completed time.Time     // When the job completed
}

// WorkerPoolStats tracks worker pool performance
type WorkerPoolStats struct {
	JobsQueued       int64   `json:"jobs_queued"`
	JobsCompleted    int64   `json:"jobs_completed"`
	JobsFailed       int64   `json:"jobs_failed"`
	JobsRetried      int64   `json:"jobs_retried"`
	ActiveWorkers    int64   `json:"active_workers"`
	QueueLength      int64   `json:"queue_length"`
	AverageJobTime   float64 `json:"average_job_time_ms"`
	ThroughputPerSec float64 `json:"throughput_per_sec"`
	ErrorRate        float64 `json:"error_rate"`
	SuccessRate      float64 `json:"success_rate"`
	UptimeSeconds    int64   `json:"uptime_seconds"`
	startTime        time.Time
}

// WorkerPool manages concurrent URL parsing with controlled concurrency
type WorkerPool struct {
	config      *WorkerPoolConfig
	parser      Parser
	jobs        chan *BatchJob
	results     chan *BatchResult
	workers     []*worker
	stats       *WorkerPoolStats
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	isRunning   int64 // atomic flag
	shutdownOnce sync.Once
}

// worker represents an individual worker goroutine
type worker struct {
	id       int
	pool     *WorkerPool
	ctx      context.Context
	jobCount int64
}

// NewWorkerPool creates a new worker pool for batch URL processing
func NewWorkerPool(parser Parser, config *WorkerPoolConfig) *WorkerPool {
	if config == nil {
		config = DefaultWorkerPoolConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	wp := &WorkerPool{
		config:  config,
		parser:  parser,
		jobs:    make(chan *BatchJob, config.QueueSize),
		results: make(chan *BatchResult, config.QueueSize),
		workers: make([]*worker, config.NumWorkers),
		stats: &WorkerPoolStats{
			startTime: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Create workers
	for i := 0; i < config.NumWorkers; i++ {
		wp.workers[i] = &worker{
			id:   i,
			pool: wp,
			ctx:  ctx,
		}
	}

	return wp
}

// DefaultWorkerPoolConfig returns sensible default configuration
func DefaultWorkerPoolConfig() *WorkerPoolConfig {
	numCPU := runtime.NumCPU()
	return &WorkerPoolConfig{
		NumWorkers:      numCPU * 2,      // 2x CPU cores for I/O bound work
		QueueSize:       numCPU * 10,     // Large queue for batching
		JobTimeout:      30 * time.Second, // Per-job timeout
		ShutdownTimeout: 10 * time.Second, // Graceful shutdown timeout
		EnableMetrics:   true,
		MaxRetries:      2, // Retry failed jobs twice
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() error {
	if !atomic.CompareAndSwapInt64(&wp.isRunning, 0, 1) {
		return NewParseError("worker_pool", "", fmt.Errorf("worker pool already running"))
	}

	// Start workers
	for _, worker := range wp.workers {
		wp.wg.Add(1)
		go worker.run()
	}

	// Start metrics collector if enabled
	if wp.config.EnableMetrics {
		wp.wg.Add(1)
		go wp.collectMetrics()
	}

	return nil
}

// Stop gracefully stops the worker pool
func (wp *WorkerPool) Stop() error {
	wp.shutdownOnce.Do(func() {
		// Set running flag to false
		atomic.StoreInt64(&wp.isRunning, 0)

		// Close job channel to stop accepting new jobs
		close(wp.jobs)

		// Wait for workers to finish or timeout
		done := make(chan struct{})
		go func() {
			wp.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// All workers finished gracefully
		case <-time.After(wp.config.ShutdownTimeout):
			// Force shutdown
			wp.cancel()
			<-done
		}

		// Close results channel
		close(wp.results)
	})

	return nil
}

// SubmitJob submits a job for processing
func (wp *WorkerPool) SubmitJob(job *BatchJob) error {
	if atomic.LoadInt64(&wp.isRunning) == 0 {
		return NewParseError("worker_pool", job.URL, fmt.Errorf("worker pool not running"))
	}

	// Set default context if none provided
	if job.Context == nil {
		ctx, cancel := context.WithTimeout(wp.ctx, wp.config.JobTimeout)
		_ = cancel // Will be cancelled when context times out
		job.Context = ctx
	}

	select {
	case wp.jobs <- job:
		atomic.AddInt64(&wp.stats.JobsQueued, 1)
		return nil
	case <-wp.ctx.Done():
		return NewParseError("worker_pool", job.URL, fmt.Errorf("worker pool shutting down"))
	}
}

// SubmitBatch submits multiple jobs at once
func (wp *WorkerPool) SubmitBatch(jobs []*BatchJob) []error {
	errors := make([]error, len(jobs))
	
	for i, job := range jobs {
		errors[i] = wp.SubmitJob(job)
	}
	
	return errors
}

// SubmitURLs is a convenience method to submit multiple URLs with default options
func (wp *WorkerPool) SubmitURLs(urls []string, options *ParserOptions) []error {
	jobs := make([]*BatchJob, len(urls))
	
	for i, url := range urls {
		jobs[i] = &BatchJob{
			ID:      fmt.Sprintf("job_%d_%d", time.Now().UnixNano(), i),
			URL:     url,
			Options: options,
		}
	}
	
	return wp.SubmitBatch(jobs)
}

// GetResults returns the results channel for consuming parsed results
func (wp *WorkerPool) GetResults() <-chan *BatchResult {
	return wp.results
}

// GetStats returns current worker pool statistics
func (wp *WorkerPool) GetStats() WorkerPoolStats {
	stats := *wp.stats // Copy
	
	// Calculate derived statistics
	total := stats.JobsCompleted + stats.JobsFailed
	if total > 0 {
		stats.ErrorRate = float64(stats.JobsFailed) / float64(total)
		stats.SuccessRate = float64(stats.JobsCompleted) / float64(total)
	}
	
	stats.UptimeSeconds = int64(time.Since(stats.startTime).Seconds())
	stats.ActiveWorkers = int64(len(wp.workers))
	stats.QueueLength = int64(len(wp.jobs))
	
	if stats.UptimeSeconds > 0 {
		stats.ThroughputPerSec = float64(total) / float64(stats.UptimeSeconds)
	}
	
	return stats
}

// IsRunning returns true if the worker pool is currently running
func (wp *WorkerPool) IsRunning() bool {
	return atomic.LoadInt64(&wp.isRunning) == 1
}

// worker.run is the main worker loop
func (w *worker) run() {
	defer w.pool.wg.Done()

	for {
		select {
		case job, ok := <-w.pool.jobs:
			if !ok {
				// Job channel closed, worker should exit
				return
			}
			
			w.processJob(job)
			
		case <-w.ctx.Done():
			// Context cancelled, worker should exit
			return
		}
	}
}

// processJob processes a single job
func (w *worker) processJob(job *BatchJob) {
	start := time.Now()
	atomic.AddInt64(&w.jobCount, 1)

	// Create job-specific context with timeout
	ctx := job.Context
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(w.ctx, w.pool.config.JobTimeout)
		defer cancel()
	}

	// Parse the URL
	result, err := w.parseWithContext(ctx, job)
	duration := time.Since(start)

	// Create result
	batchResult := &BatchResult{
		JobID:     job.ID,
		URL:       job.URL,
		Result:    result,
		Error:     err,
		Duration:  duration,
		WorkerID:  w.id,
		Completed: time.Now(),
	}

	// Handle errors and retries
	if err != nil && job.Retries < w.pool.config.MaxRetries {
		// Retry the job
		job.Retries++
		atomic.AddInt64(&w.pool.stats.JobsRetried, 1)
		
		select {
		case w.pool.jobs <- job:
			// Job requeued for retry
			return
		case <-w.ctx.Done():
			// Can't requeue, send error result
		}
	}

	// Send result
	select {
	case w.pool.results <- batchResult:
		if err != nil {
			atomic.AddInt64(&w.pool.stats.JobsFailed, 1)
		} else {
			atomic.AddInt64(&w.pool.stats.JobsCompleted, 1)
		}
	case <-w.ctx.Done():
		// Worker pool shutting down
		return
	}
}

// parseWithContext performs URL parsing with context cancellation
func (w *worker) parseWithContext(ctx context.Context, job *BatchJob) (*Result, error) {
	// Create a channel to receive the result
	resultChan := make(chan *Result, 1)
	errorChan := make(chan error, 1)

	// Start parsing in a goroutine
	go func() {
		result, err := w.pool.parser.Parse(job.URL, job.Options)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- result
		}
	}()

	// Wait for result or context cancellation
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		return nil, NewTimeoutError(job.URL, "parse", w.pool.config.JobTimeout)
	}
}

// collectMetrics runs in a separate goroutine to collect performance metrics
func (wp *WorkerPool) collectMetrics() {
	defer wp.wg.Done()
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	var jobTimes []time.Duration
	
	for {
		select {
		case <-ticker.C:
			// Sample job durations from recent results
			timeout := time.After(100 * time.Millisecond)
			for {
				select {
				case result := <-wp.results:
					jobTimes = append(jobTimes, result.Duration)
					
					// Keep only last 100 measurements for moving average
					if len(jobTimes) > 100 {
						jobTimes = jobTimes[1:]
					}
					
					// Forward result back to the channel (non-blocking)
					select {
					case wp.results <- result:
					default:
						// Channel full, drop the result
					}
					
				case <-timeout:
					goto updateStats
				}
			}
			
		updateStats:
			// Update average job time
			if len(jobTimes) > 0 {
				var total time.Duration
				for _, t := range jobTimes {
					total += t
				}
				wp.stats.AverageJobTime = float64(total/time.Duration(len(jobTimes))) / float64(time.Millisecond)
			}
			
		case <-wp.ctx.Done():
			return
		}
	}
}

// BatchProcessor provides a high-level interface for batch URL processing
type BatchProcessor struct {
	pool    *WorkerPool
	results []BatchResult
	mutex   sync.RWMutex
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(parser Parser, config *WorkerPoolConfig) *BatchProcessor {
	return &BatchProcessor{
		pool: NewWorkerPool(parser, config),
	}
}

// ProcessURLs processes a list of URLs and returns all results
func (bp *BatchProcessor) ProcessURLs(urls []string, options *ParserOptions) ([]BatchResult, error) {
	// Start worker pool
	if err := bp.pool.Start(); err != nil {
		return nil, err
	}
	defer bp.pool.Stop()

	// Submit all URLs
	errors := bp.pool.SubmitURLs(urls, options)
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	// Collect results
	var results []BatchResult
	expectedResults := len(urls)
	
	for i := 0; i < expectedResults; i++ {
		select {
		case result := <-bp.pool.GetResults():
			results = append(results, *result)
		case <-time.After(60 * time.Second): // Overall timeout
			return results, NewTimeoutError("batch", "process_urls", 60*time.Second)
		}
	}

	return results, nil
}

// ProcessURLsConcurrent processes URLs and calls a callback for each result as it completes
func (bp *BatchProcessor) ProcessURLsConcurrent(urls []string, options *ParserOptions, callback func(*BatchResult)) error {
	// Start worker pool
	if err := bp.pool.Start(); err != nil {
		return err
	}
	defer bp.pool.Stop()

	// Submit all URLs
	errors := bp.pool.SubmitURLs(urls, options)
	for _, err := range errors {
		if err != nil {
			return err
		}
	}

	// Process results as they arrive
	expectedResults := len(urls)
	for i := 0; i < expectedResults; i++ {
		select {
		case result := <-bp.pool.GetResults():
			callback(result)
		case <-time.After(60 * time.Second):
			return NewTimeoutError("batch", "process_urls_concurrent", 60*time.Second)
		}
	}

	return nil
}