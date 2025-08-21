// ABOUTME: Optimized batch processing API combining object pooling with concurrent processing
// ABOUTME: Ideal for high-throughput API scenarios where you need to parse multiple URLs efficiently

package parser

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BatchRequest represents a single parsing request in a batch
type BatchRequest struct {
	ID      string            // Unique identifier for this request
	URL     string            // URL to parse (required)
	HTML    string            // Pre-fetched HTML (optional, if provided URL fetching is skipped)
	Options *ParserOptions    // Parser options for this specific request
	Context context.Context   // Request-specific context
	Meta    map[string]interface{} // Custom metadata for this request
}

// BatchResponse represents the result of a batch parsing request
type BatchResponse struct {
	ID          string        // Request ID
	Result      *Result       // Parsed result (nil if error)
	Error       error         // Error if parsing failed
	Duration    time.Duration // Time taken to process this request
	WorkerID    int           // ID of worker that processed this request
	ProcessedAt time.Time     // When this request was completed
}

// BatchAPIConfig configures the batch processing API
type BatchAPIConfig struct {
	// Concurrency settings
	MaxWorkers      int           // Maximum number of concurrent workers (default: runtime.NumCPU())
	QueueSize       int           // Size of request queue (default: 1000)
	ProcessingTimeout time.Duration // Timeout per individual request (default: 30s)
	
	// Performance settings
	UseObjectPooling bool          // Whether to use object pooling (default: true)
	ParserOptions    *ParserOptions // Default parser options
	
	// Advanced settings
	EnableMetrics    bool          // Whether to collect detailed metrics (default: true)
	RetryCount       int           // Number of retries for failed requests (default: 1)
	RetryDelay       time.Duration // Delay between retries (default: 1s)
}

// DefaultBatchAPIConfig returns sensible defaults for batch processing
func DefaultBatchAPIConfig() *BatchAPIConfig {
	return &BatchAPIConfig{
		MaxWorkers:        8, // Conservative default
		QueueSize:         1000,
		ProcessingTimeout: 30 * time.Second,
		UseObjectPooling:  true,
		ParserOptions:     DefaultParserOptions(),
		EnableMetrics:     true,
		RetryCount:        1,
		RetryDelay:        1 * time.Second,
	}
}

// BatchAPI provides optimized batch processing for high-throughput scenarios
type BatchAPI struct {
	config           *BatchAPIConfig
	htParser         *HighThroughputParser
	requestQueue     chan *BatchRequest
	responseQueue    chan *BatchResponse
	workers          []*batchWorker
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	metrics          *BatchMetrics
	isRunning        bool
	mu               sync.RWMutex
}

// BatchMetrics tracks detailed performance metrics for the batch API
type BatchMetrics struct {
	mu                     sync.RWMutex
	TotalRequests         int64         `json:"total_requests"`
	CompletedRequests     int64         `json:"completed_requests"`
	FailedRequests        int64         `json:"failed_requests"`
	RetriedRequests       int64         `json:"retried_requests"`
	AverageResponseTime   float64       `json:"avg_response_time_ms"`
	ThroughputPerSecond   float64       `json:"throughput_per_sec"`
	ActiveWorkers         int           `json:"active_workers"`
	QueueLength           int           `json:"queue_length"`
	StartTime             time.Time     `json:"start_time"`
	LastRequestTime       time.Time     `json:"last_request_time"`
	PeakThroughput        float64       `json:"peak_throughput_per_sec"`
	ErrorsByType          map[string]int64 `json:"errors_by_type"`
}

// batchWorker represents a worker that processes batch requests
type batchWorker struct {
	id       int
	api      *BatchAPI
	ctx      context.Context
	requests chan *BatchRequest
}

// NewBatchAPI creates a new optimized batch processing API
func NewBatchAPI(config *BatchAPIConfig) *BatchAPI {
	if config == nil {
		config = DefaultBatchAPIConfig()
	}
	
	// Ensure required fields have sensible defaults
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = 8
	}
	if config.QueueSize <= 0 {
		config.QueueSize = 1000
	}
	if config.ProcessingTimeout <= 0 {
		config.ProcessingTimeout = 30 * time.Second
	}
	if config.ParserOptions == nil {
		config.ParserOptions = DefaultParserOptions()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	api := &BatchAPI{
		config:        config,
		requestQueue:  make(chan *BatchRequest, config.QueueSize),
		responseQueue: make(chan *BatchResponse, config.QueueSize),
		ctx:           ctx,
		cancel:        cancel,
		metrics: &BatchMetrics{
			StartTime:    time.Now(),
			ErrorsByType: make(map[string]int64),
		},
	}
	
	if config.UseObjectPooling {
		api.htParser = NewHighThroughputParser(config.ParserOptions)
	}
	
	return api
}

// Start begins processing batch requests
func (api *BatchAPI) Start() error {
	api.mu.Lock()
	defer api.mu.Unlock()
	
	if api.isRunning {
		return fmt.Errorf("batch API is already running")
	}
	
	// Create workers
	api.workers = make([]*batchWorker, api.config.MaxWorkers)
	for i := 0; i < api.config.MaxWorkers; i++ {
		worker := &batchWorker{
			id:       i,
			api:      api,
			ctx:      api.ctx,
			requests: api.requestQueue,
		}
		api.workers[i] = worker
		
		api.wg.Add(1)
		go worker.run()
	}
	
	api.isRunning = true
	return nil
}

// Stop gracefully shuts down the batch API
func (api *BatchAPI) Stop() error {
	api.mu.Lock()
	defer api.mu.Unlock()
	
	if !api.isRunning {
		return fmt.Errorf("batch API is not running")
	}
	
	// Signal shutdown
	api.cancel()
	
	// Close request queue to stop accepting new requests
	close(api.requestQueue)
	
	// Wait for all workers to finish
	api.wg.Wait()
	
	// Close response queue
	close(api.responseQueue)
	
	api.isRunning = false
	return nil
}

// Submit submits a batch request for processing
func (api *BatchAPI) Submit(request *BatchRequest) error {
	api.mu.RLock()
	defer api.mu.RUnlock()
	
	if !api.isRunning {
		return fmt.Errorf("batch API is not running")
	}
	
	// Generate ID if not provided
	if request.ID == "" {
		request.ID = fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), len(api.requestQueue))
	}
	
	// Add context if not provided
	if request.Context == nil {
		var cancel context.CancelFunc
		request.Context, cancel = context.WithTimeout(api.ctx, api.config.ProcessingTimeout)
		defer cancel()
	}
	
	// Update metrics
	if api.config.EnableMetrics {
		api.updateSubmissionMetrics()
	}
	
	select {
	case api.requestQueue <- request:
		return nil
	case <-api.ctx.Done():
		return fmt.Errorf("batch API is shutting down")
	default:
		return fmt.Errorf("request queue is full")
	}
}

// SubmitBatch submits multiple requests as a batch
func (api *BatchAPI) SubmitBatch(requests []*BatchRequest) []error {
	errors := make([]error, 0, len(requests))
	
	for _, request := range requests {
		if err := api.Submit(request); err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return errors
	}
	return nil
}

// GetResponse retrieves a processed response (blocking)
func (api *BatchAPI) GetResponse() *BatchResponse {
	select {
	case response := <-api.responseQueue:
		return response
	case <-api.ctx.Done():
		return nil
	}
}

// GetResponseNonBlocking retrieves a processed response (non-blocking)
func (api *BatchAPI) GetResponseNonBlocking() *BatchResponse {
	select {
	case response := <-api.responseQueue:
		return response
	default:
		return nil
	}
}

// ProcessBatch is a convenience method that processes a batch and waits for all results
func (api *BatchAPI) ProcessBatch(requests []*BatchRequest) ([]*BatchResponse, error) {
	if !api.isRunning {
		return nil, fmt.Errorf("batch API is not running")
	}
	
	// Submit all requests
	submitErrors := api.SubmitBatch(requests)
	if len(submitErrors) > 0 {
		return nil, fmt.Errorf("failed to submit %d requests", len(submitErrors))
	}
	
	// Collect responses
	responses := make([]*BatchResponse, 0, len(requests))
	responseMap := make(map[string]*BatchResponse)
	
	// Create a timeout for the entire batch
	timeout := time.After(api.config.ProcessingTimeout * time.Duration(len(requests)))
	
	for len(responses) < len(requests) {
		select {
		case response := <-api.responseQueue:
			if response != nil {
				responseMap[response.ID] = response
				responses = append(responses, response)
			}
		case <-timeout:
			return responses, fmt.Errorf("batch processing timed out, got %d/%d responses", len(responses), len(requests))
		case <-api.ctx.Done():
			return responses, fmt.Errorf("batch API is shutting down")
		}
	}
	
	return responses, nil
}

// GetMetrics returns current performance metrics
func (api *BatchAPI) GetMetrics() *BatchMetrics {
	if !api.config.EnableMetrics {
		return nil
	}
	
	api.metrics.mu.RLock()
	defer api.metrics.mu.RUnlock()
	
	// Update dynamic metrics
	api.metrics.QueueLength = len(api.requestQueue)
	api.metrics.ActiveWorkers = api.config.MaxWorkers
	
	// Calculate throughput
	elapsed := time.Since(api.metrics.StartTime).Seconds()
	if elapsed > 0 {
		api.metrics.ThroughputPerSecond = float64(api.metrics.CompletedRequests) / elapsed
	}
	
	// Return copy to avoid race conditions
	errorsCopy := make(map[string]int64)
	for k, v := range api.metrics.ErrorsByType {
		errorsCopy[k] = v
	}
	
	return &BatchMetrics{
		TotalRequests:       api.metrics.TotalRequests,
		CompletedRequests:   api.metrics.CompletedRequests,
		FailedRequests:      api.metrics.FailedRequests,
		RetriedRequests:     api.metrics.RetriedRequests,
		AverageResponseTime: api.metrics.AverageResponseTime,
		ThroughputPerSecond: api.metrics.ThroughputPerSecond,
		ActiveWorkers:       api.metrics.ActiveWorkers,
		QueueLength:         api.metrics.QueueLength,
		StartTime:           api.metrics.StartTime,
		LastRequestTime:     api.metrics.LastRequestTime,
		PeakThroughput:      api.metrics.PeakThroughput,
		ErrorsByType:        errorsCopy,
	}
}

// IsRunning returns whether the batch API is currently running
func (api *BatchAPI) IsRunning() bool {
	api.mu.RLock()
	defer api.mu.RUnlock()
	return api.isRunning
}

// Worker implementation
func (worker *batchWorker) run() {
	defer worker.api.wg.Done()
	
	for {
		select {
		case request := <-worker.requests:
			if request == nil {
				return // Channel closed
			}
			worker.processRequest(request)
		case <-worker.ctx.Done():
			return
		}
	}
}

func (worker *batchWorker) processRequest(request *BatchRequest) {
	start := time.Now()
	
	response := &BatchResponse{
		ID:          request.ID,
		WorkerID:    worker.id,
		ProcessedAt: time.Now(),
	}
	
	// Process the request with retries
	var err error
	for attempt := 0; attempt <= worker.api.config.RetryCount; attempt++ {
		if attempt > 0 {
			// Update retry metrics
			worker.api.updateRetryMetrics()
			time.Sleep(worker.api.config.RetryDelay)
		}
		
		// Use provided options or defaults
		opts := request.Options
		if opts == nil {
			opts = worker.api.config.ParserOptions
		}
		
		// Parse based on whether HTML is provided
		if request.HTML != "" {
			if worker.api.config.UseObjectPooling {
				response.Result, err = worker.api.htParser.ParseHTML(request.HTML, request.URL, opts)
			} else {
				parser := New(opts)
				response.Result, err = parser.ParseHTML(request.HTML, request.URL, opts)
			}
		} else {
			if worker.api.config.UseObjectPooling {
				response.Result, err = worker.api.htParser.Parse(request.URL, opts)
			} else {
				parser := New(opts)
				response.Result, err = parser.Parse(request.URL, opts)
			}
		}
		
		if err == nil {
			break // Success, no need to retry
		}
	}
	
	response.Error = err
	response.Duration = time.Since(start)
	
	// Update metrics
	if worker.api.config.EnableMetrics {
		worker.api.updateCompletionMetrics(response)
	}
	
	// Send response
	select {
	case worker.api.responseQueue <- response:
	case <-worker.ctx.Done():
		// If we can't send the response due to shutdown, return result to pool if using object pooling
		if worker.api.config.UseObjectPooling && response.Result != nil {
			worker.api.htParser.ReturnResult(response.Result)
		}
	}
}

// Metrics update methods
func (api *BatchAPI) updateSubmissionMetrics() {
	api.metrics.mu.Lock()
	defer api.metrics.mu.Unlock()
	
	api.metrics.TotalRequests++
	api.metrics.LastRequestTime = time.Now()
}

func (api *BatchAPI) updateRetryMetrics() {
	api.metrics.mu.Lock()
	defer api.metrics.mu.Unlock()
	
	api.metrics.RetriedRequests++
}

func (api *BatchAPI) updateCompletionMetrics(response *BatchResponse) {
	api.metrics.mu.Lock()
	defer api.metrics.mu.Unlock()
	
	if response.Error != nil {
		api.metrics.FailedRequests++
		errorType := fmt.Sprintf("%T", response.Error)
		api.metrics.ErrorsByType[errorType]++
	} else {
		api.metrics.CompletedRequests++
	}
	
	// Update average response time
	durationMs := float64(response.Duration.Nanoseconds()) / 1e6
	if api.metrics.CompletedRequests == 1 {
		api.metrics.AverageResponseTime = durationMs
	} else {
		// Rolling average
		alpha := 0.1
		api.metrics.AverageResponseTime = alpha*durationMs + (1-alpha)*api.metrics.AverageResponseTime
	}
	
	// Update peak throughput
	elapsed := time.Since(api.metrics.StartTime).Seconds()
	if elapsed > 0 {
		currentThroughput := float64(api.metrics.CompletedRequests) / elapsed
		if currentThroughput > api.metrics.PeakThroughput {
			api.metrics.PeakThroughput = currentThroughput
		}
	}
}

// Global batch API instance for convenience - initialized lazily
var GlobalBatchAPI *BatchAPI

// GetGlobalBatchAPI returns the global batch API, initializing it if needed
func GetGlobalBatchAPI() *BatchAPI {
	if GlobalBatchAPI == nil {
		config := DefaultBatchAPIConfig()
		GlobalBatchAPI = NewBatchAPI(config)
		GlobalBatchAPI.Start()
	}
	return GlobalBatchAPI
}

// ProcessURLsBatch processes multiple URLs concurrently using the global batch API
func ProcessURLsBatch(urls []string, opts *ParserOptions) ([]*BatchResponse, error) {
	requests := make([]*BatchRequest, len(urls))
	for i, url := range urls {
		requests[i] = &BatchRequest{
			URL:     url,
			Options: opts,
		}
	}
	
	return GetGlobalBatchAPI().ProcessBatch(requests)
}