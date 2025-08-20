// ABOUTME: High-performance HTTP connection pooling for concurrent URL processing
// ABOUTME: Optimized for parsing multiple URLs simultaneously with connection reuse

package resource

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

// ConnectionPoolConfig configures the HTTP connection pool for optimal performance
type ConnectionPoolConfig struct {
	// Connection limits
	MaxIdleConns        int           // Total pool size across all hosts
	MaxIdleConnsPerHost int           // Connections per individual host
	MaxConnsPerHost     int           // Maximum active connections per host
	
	// Timeouts
	IdleConnTimeout       time.Duration // How long to keep idle connections
	ConnectTimeout        time.Duration // TCP connection timeout
	TLSHandshakeTimeout   time.Duration // TLS handshake timeout
	ResponseHeaderTimeout time.Duration // Time to read response headers
	
	// Keepalive settings
	KeepAlive           time.Duration // TCP keepalive interval
	DisableKeepAlives   bool          // Disable HTTP keepalive
	DisableCompression  bool          // Disable compression
	
	// HTTP/2 settings
	EnableHTTP2         bool          // Enable HTTP/2 support
	HTTP2MaxConcurrent  int           // Max concurrent streams per connection
	
	// Advanced settings
	ExpectContinueTimeout time.Duration // 100-continue timeout
	WriteBufferSize       int           // Write buffer size
	ReadBufferSize        int           // Read buffer size
}

// OptimizedTransport creates a high-performance HTTP transport for concurrent parsing
type OptimizedTransport struct {
	*http.Transport
	config *ConnectionPoolConfig
	stats  *ConnectionStats
	mutex  sync.RWMutex
}

// ConnectionStats tracks connection pool performance
type ConnectionStats struct {
	ActiveConnections    int64         // Current active connections
	IdleConnections      int64         // Current idle connections
	TotalConnections     int64         // Total connections created
	ConnectionsReused    int64         // Connections reused from pool
	ConnectionTimeouts   int64         // Connection timeout errors
	DNSLookupTime        time.Duration // Average DNS lookup time
	ConnectionTime       time.Duration // Average connection establishment time
	TLSHandshakeTime     time.Duration // Average TLS handshake time
	RequestsPerSecond    float64       // Current RPS
	mutex                sync.RWMutex
}

// NewHighPerformanceConfig returns optimized connection pool settings for concurrent parsing
func NewHighPerformanceConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		// Increased connection limits for high concurrency
		MaxIdleConns:        200,  // Up from default 100
		MaxIdleConnsPerHost: 20,   // Up from default 2
		MaxConnsPerHost:     50,   // Up from default 0 (unlimited)
		
		// Optimized timeouts
		IdleConnTimeout:       120 * time.Second, // Keep connections longer
		ConnectTimeout:        10 * time.Second,  // Fast connection establishment
		TLSHandshakeTimeout:   10 * time.Second,  // Fast TLS handshake
		ResponseHeaderTimeout: 15 * time.Second,  // Reasonable header timeout
		
		// Keepalive optimization
		KeepAlive:          30 * time.Second, // Standard keepalive
		DisableKeepAlives:  false,            // Enable connection reuse
		DisableCompression: false,            // Enable compression for bandwidth
		
		// HTTP/2 configuration
		EnableHTTP2:        true, // Enable HTTP/2 for better multiplexing
		HTTP2MaxConcurrent: 100,  // Many concurrent streams per connection
		
		// Advanced settings
		ExpectContinueTimeout: 1 * time.Second,
		WriteBufferSize:       64 * 1024, // 64KB write buffer
		ReadBufferSize:        64 * 1024, // 64KB read buffer
	}
}

// NewConservativeConfig returns safe connection pool settings for stable operation
func NewConservativeConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 5,
		MaxConnsPerHost:     20,
		
		IdleConnTimeout:       90 * time.Second,
		ConnectTimeout:        15 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		
		KeepAlive:          30 * time.Second,
		DisableKeepAlives:  false,
		DisableCompression: false,
		
		EnableHTTP2:        false, // Disable HTTP/2 for stability
		HTTP2MaxConcurrent: 10,
		
		ExpectContinueTimeout: 1 * time.Second,
		WriteBufferSize:       32 * 1024,
		ReadBufferSize:        32 * 1024,
	}
}

// NewOptimizedTransport creates a new optimized HTTP transport
func NewOptimizedTransport(config *ConnectionPoolConfig) *OptimizedTransport {
	if config == nil {
		config = NewHighPerformanceConfig()
	}

	// Create base transport
	transport := &http.Transport{
		// Connection pooling
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		
		// Timeouts
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
		
		// Keepalive and compression
		DisableKeepAlives:  config.DisableKeepAlives,
		DisableCompression: config.DisableCompression,
		
		// Custom dialer for connection tracking
		DialContext: (&net.Dialer{
			Timeout:   config.ConnectTimeout,
			KeepAlive: config.KeepAlive,
		}).DialContext,
		
		// TLS configuration
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
		
		// Buffer sizes
		WriteBufferSize: config.WriteBufferSize,
		ReadBufferSize:  config.ReadBufferSize,
	}

	// Configure HTTP/2
	if !config.EnableHTTP2 {
		// Disable HTTP/2 by setting TLSNextProto to non-nil empty map
		transport.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	}

	optimized := &OptimizedTransport{
		Transport: transport,
		config:    config,
		stats:     &ConnectionStats{},
	}

	return optimized
}

// RoundTrip implements http.RoundTripper with statistics tracking
func (ot *OptimizedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	
	// Track request
	ot.stats.mutex.Lock()
	ot.stats.ActiveConnections++
	ot.stats.mutex.Unlock()
	
	// Perform request
	resp, err := ot.Transport.RoundTrip(req)
	
	// Track completion
	duration := time.Since(start)
	ot.stats.mutex.Lock()
	ot.stats.ActiveConnections--
	if err == nil {
		ot.stats.ConnectionsReused++
	}
	ot.stats.mutex.Unlock()
	
	// Update average timing (simplified)
	if err == nil {
		ot.updateAverageTime(duration)
	}
	
	return resp, err
}

// updateAverageTime updates average connection timing
func (ot *OptimizedTransport) updateAverageTime(duration time.Duration) {
	ot.stats.mutex.Lock()
	defer ot.stats.mutex.Unlock()
	
	// Simple exponential moving average
	alpha := 0.1
	if ot.stats.ConnectionTime == 0 {
		ot.stats.ConnectionTime = duration
	} else {
		ot.stats.ConnectionTime = time.Duration(float64(ot.stats.ConnectionTime)*(1-alpha) + float64(duration)*alpha)
	}
}

// GetStats returns current connection pool statistics
func (ot *OptimizedTransport) GetStats() ConnectionStats {
	ot.stats.mutex.RLock()
	defer ot.stats.mutex.RUnlock()
	return *ot.stats
}

// NewOptimizedHTTPClient creates an HTTP client with optimized connection pooling
func NewOptimizedHTTPClient(config *ConnectionPoolConfig, headers map[string]string) *HTTPClient {
	if config == nil {
		config = NewHighPerformanceConfig()
	}

	transport := NewOptimizedTransport(config)
	
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second, // Overall request timeout
	}

	return &HTTPClient{
		client:  client,
		headers: headers,
	}
}

// ConnectionPoolManager manages multiple connection pools for different use cases
type ConnectionPoolManager struct {
	pools map[string]*OptimizedTransport
	mutex sync.RWMutex
}

// NewConnectionPoolManager creates a new pool manager
func NewConnectionPoolManager() *ConnectionPoolManager {
	return &ConnectionPoolManager{
		pools: make(map[string]*OptimizedTransport),
	}
}

// GetPool returns a connection pool for the specified profile
func (cpm *ConnectionPoolManager) GetPool(profile string) *OptimizedTransport {
	cpm.mutex.RLock()
	if pool, exists := cpm.pools[profile]; exists {
		cpm.mutex.RUnlock()
		return pool
	}
	cpm.mutex.RUnlock()

	// Create new pool
	cpm.mutex.Lock()
	defer cpm.mutex.Unlock()

	// Double-check pattern
	if pool, exists := cpm.pools[profile]; exists {
		return pool
	}

	var config *ConnectionPoolConfig
	switch profile {
	case "high_performance":
		config = NewHighPerformanceConfig()
	case "conservative":
		config = NewConservativeConfig()
	default:
		config = NewHighPerformanceConfig()
	}

	pool := NewOptimizedTransport(config)
	cpm.pools[profile] = pool
	return pool
}

// GetPoolStats returns statistics for all pools
func (cpm *ConnectionPoolManager) GetPoolStats() map[string]ConnectionStats {
	cpm.mutex.RLock()
	defer cpm.mutex.RUnlock()

	stats := make(map[string]ConnectionStats)
	for name, pool := range cpm.pools {
		stats[name] = pool.GetStats()
	}
	return stats
}

// Global connection pool manager instance
var globalPoolManager = NewConnectionPoolManager()

// GetGlobalPool returns a globally shared connection pool
func GetGlobalPool(profile string) *OptimizedTransport {
	return globalPoolManager.GetPool(profile)
}

// GetGlobalPoolStats returns statistics for all global pools
func GetGlobalPoolStats() map[string]ConnectionStats {
	return globalPoolManager.GetPoolStats()
}