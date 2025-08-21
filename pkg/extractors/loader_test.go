package extractors

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/parser-go/pkg/extractors/custom"
)

func TestDefaultLoaderConfig(t *testing.T) {
	config := DefaultLoaderConfig()
	
	if config.MaxCacheSize != 50 {
		t.Errorf("Expected MaxCacheSize 50, got %d", config.MaxCacheSize)
	}
	
	if config.CacheExpiration != 30*time.Minute {
		t.Errorf("Expected CacheExpiration 30 minutes, got %v", config.CacheExpiration)
	}
	
	if !config.PreloadCommonSites {
		t.Error("Expected PreloadCommonSites to be true")
	}
	
	if !config.EnableMetrics {
		t.Error("Expected EnableMetrics to be true")
	}
	
	if config.MaxLoadAttempts != 3 {
		t.Errorf("Expected MaxLoadAttempts 3, got %d", config.MaxLoadAttempts)
	}
	
	if config.LoadTimeout != 5*time.Second {
		t.Errorf("Expected LoadTimeout 5 seconds, got %v", config.LoadTimeout)
	}
}

func TestNewExtractorLoader(t *testing.T) {
	registry := custom.NewRegistryManager()
	config := DefaultLoaderConfig()
	
	loader := NewExtractorLoader(config, registry)
	defer loader.Close()
	
	if loader == nil {
		t.Fatal("Expected non-nil loader")
	}
	
	if loader.config != config {
		t.Error("Config not set correctly")
	}
	
	if loader.registry != registry {
		t.Error("Registry not set correctly")
	}
	
	if loader.cache == nil {
		t.Error("Cache not initialized")
	}
	
	if loader.lruList == nil {
		t.Error("LRU list not initialized")
	}
	
	if loader.metrics == nil {
		t.Error("Metrics not initialized")
	}
}

func TestNewExtractorLoaderWithNilConfig(t *testing.T) {
	loader := NewExtractorLoader(nil, nil)
	defer loader.Close()
	
	if loader.config.MaxCacheSize != 50 {
		t.Error("Should use default config when nil provided")
	}
	
	if loader.registry == nil {
		t.Error("Should use global registry when nil provided")
	}
}

func TestExtractorLoaderBasicLoad(t *testing.T) {
	registry := custom.NewRegistryManager()
	config := &LoaderConfig{
		MaxCacheSize:       10,
		CacheExpiration:    5 * time.Minute,
		PreloadCommonSites: false,
		EnableMetrics:      true,
		MaxLoadAttempts:    3,
		LoadTimeout:        5 * time.Second,
	}
	
	loader := NewExtractorLoader(config, registry)
	defer loader.Close()
	
	// Create a test extractor
	testExtractor := &custom.CustomExtractor{
		Domain: "example.com",
		Title: map[string]interface{}{
			"selectors": []string{"h1", ".title"},
		},
	}
	
	// Register the extractor
	err := registry.Register(testExtractor)
	if err != nil {
		t.Fatalf("Failed to register extractor: %v", err)
	}
	
	// Load the extractor
	loaded, err := loader.LoadExtractor("example.com")
	if err != nil {
		t.Fatalf("Failed to load extractor: %v", err)
	}
	
	if loaded == nil {
		t.Fatal("Expected non-nil loaded extractor")
	}
	
	if loaded.Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got '%s'", loaded.Domain)
	}
}

func TestExtractorLoaderCaching(t *testing.T) {
	registry := custom.NewRegistryManager()
	config := &LoaderConfig{
		MaxCacheSize:       5,
		CacheExpiration:    1 * time.Minute,
		PreloadCommonSites: false,
		EnableMetrics:      true,
		MaxLoadAttempts:    1,
		LoadTimeout:        1 * time.Second,
	}
	
	loader := NewExtractorLoader(config, registry)
	defer loader.Close()
	
	// Create test extractor
	testExtractor := &custom.CustomExtractor{
		Domain: "cache-test.com",
		Title: map[string]interface{}{
			"selectors": []string{"h1"},
		},
	}
	
	registry.Register(testExtractor)
	
	// First load - should miss cache
	metrics1 := loader.GetMetrics()
	initialMisses := metrics1.CacheMisses
	
	loaded1, err := loader.LoadExtractor("cache-test.com")
	if err != nil {
		t.Fatalf("First load failed: %v", err)
	}
	
	metrics2 := loader.GetMetrics()
	if metrics2.CacheMisses != initialMisses+1 {
		t.Error("Expected cache miss on first load")
	}
	
	// Second load - should hit cache
	loaded2, err := loader.LoadExtractor("cache-test.com")
	if err != nil {
		t.Fatalf("Second load failed: %v", err)
	}
	
	metrics3 := loader.GetMetrics()
	if metrics3.CacheHits <= metrics1.CacheHits {
		t.Error("Expected cache hit on second load")
	}
	
	// Should be same instance
	if loaded1 != loaded2 {
		t.Error("Expected same extractor instance from cache")
	}
}

func TestExtractorLoaderLRUEviction(t *testing.T) {
	registry := custom.NewRegistryManager()
	config := &LoaderConfig{
		MaxCacheSize:       2, // Very small cache for testing
		CacheExpiration:    1 * time.Minute,
		PreloadCommonSites: false,
		EnableMetrics:      true,
		MaxLoadAttempts:    1,
		LoadTimeout:        1 * time.Second,
	}
	
	loader := NewExtractorLoader(config, registry)
	defer loader.Close()
	
	// Create test extractors
	for i := 1; i <= 3; i++ {
		domain := fmt.Sprintf("test%d.com", i)
		extractor := &custom.CustomExtractor{
			Domain: domain,
			Title: map[string]interface{}{
				"selectors": []string{"h1"},
			},
		}
		registry.Register(extractor)
	}
	
	// Load first extractor
	_, err := loader.LoadExtractor("test1.com")
	if err != nil {
		t.Fatalf("Failed to load test1.com: %v", err)
	}
	
	// Load second extractor
	_, err = loader.LoadExtractor("test2.com")
	if err != nil {
		t.Fatalf("Failed to load test2.com: %v", err)
	}
	
	// Check cache size
	size, maxSize, _ := loader.GetCacheStats()
	if size != 2 {
		t.Errorf("Expected cache size 2, got %d", size)
	}
	if maxSize != 2 {
		t.Errorf("Expected max cache size 2, got %d", maxSize)
	}
	
	// Load third extractor - should evict first
	_, err = loader.LoadExtractor("test3.com")
	if err != nil {
		t.Fatalf("Failed to load test3.com: %v", err)
	}
	
	// Cache should still be size 2
	size, _, _ = loader.GetCacheStats()
	if size != 2 {
		t.Errorf("Expected cache size 2 after eviction, got %d", size)
	}
	
	// Check that eviction occurred
	metrics := loader.GetMetrics()
	if metrics.EvictionCount == 0 {
		t.Error("Expected at least one eviction")
	}
}

func TestExtractorLoaderCacheExpiration(t *testing.T) {
	registry := custom.NewRegistryManager()
	config := &LoaderConfig{
		MaxCacheSize:       10,
		CacheExpiration:    50 * time.Millisecond, // Very short expiration
		PreloadCommonSites: false,
		EnableMetrics:      true,
		MaxLoadAttempts:    1,
		LoadTimeout:        1 * time.Second,
	}
	
	loader := NewExtractorLoader(config, registry)
	defer loader.Close()
	
	// Create test extractor
	testExtractor := &custom.CustomExtractor{
		Domain: "expire-test.com",
		Title: map[string]interface{}{
			"selectors": []string{"h1"},
		},
	}
	registry.Register(testExtractor)
	
	// Load extractor
	_, err := loader.LoadExtractor("expire-test.com")
	if err != nil {
		t.Fatalf("Failed to load extractor: %v", err)
	}
	
	// Wait for expiration
	time.Sleep(100 * time.Millisecond)
	
	// Load again - should be cache miss due to expiration
	initialMisses := loader.GetMetrics().CacheMisses
	_, err = loader.LoadExtractor("expire-test.com")
	if err != nil {
		t.Fatalf("Failed to reload expired extractor: %v", err)
	}
	
	finalMisses := loader.GetMetrics().CacheMisses
	if finalMisses <= initialMisses {
		t.Error("Expected cache miss after expiration")
	}
}

func TestExtractorLoaderLoadExtractorByHTML(t *testing.T) {
	registry := custom.NewRegistryManager()
	loader := NewExtractorLoader(DefaultLoaderConfig(), registry)
	defer loader.Close()
	
	// Create test extractor with HTML detector
	testExtractor := &custom.CustomExtractor{
		Domain: "html-test.com",
		Title: map[string]interface{}{
			"selectors": []string{"h1"},
		},
	}
	
	registry.Register(testExtractor)
	registry.RegisterHTMLDetector(".special-class", testExtractor)
	
	// Create test HTML document
	html := `<html><body><div class="special-class">Test</div></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create test document: %v", err)
	}
	
	// Load by HTML
	loaded, err := loader.LoadExtractorByHTML(doc)
	if err != nil {
		t.Fatalf("Failed to load by HTML: %v", err)
	}
	
	if loaded == nil {
		t.Fatal("Expected non-nil loaded extractor")
	}
	
	if loaded.Domain != "html-test.com" {
		t.Errorf("Expected domain 'html-test.com', got '%s'", loaded.Domain)
	}
}

func TestExtractorLoaderLoadExtractorByHTMLNilDoc(t *testing.T) {
	loader := NewExtractorLoader(DefaultLoaderConfig(), custom.NewRegistryManager())
	defer loader.Close()
	
	_, err := loader.LoadExtractorByHTML(nil)
	if err == nil {
		t.Error("Expected error for nil document")
	}
	
	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Expected nil error message, got: %v", err)
	}
}

func TestExtractorLoaderMetrics(t *testing.T) {
	registry := custom.NewRegistryManager()
	config := &LoaderConfig{
		MaxCacheSize:       10,
		CacheExpiration:    1 * time.Minute,
		PreloadCommonSites: false,
		EnableMetrics:      true,
		MaxLoadAttempts:    1,
		LoadTimeout:        1 * time.Second,
	}
	
	loader := NewExtractorLoader(config, registry)
	defer loader.Close()
	
	// Create test extractor
	testExtractor := &custom.CustomExtractor{
		Domain: "metrics-test.com",
		Title: map[string]interface{}{
			"selectors": []string{"h1"},
		},
	}
	registry.Register(testExtractor)
	
	// Initial metrics
	initialMetrics := loader.GetMetrics()
	
	// Load extractor multiple times
	for i := 0; i < 5; i++ {
		_, err := loader.LoadExtractor("metrics-test.com")
		if err != nil {
			t.Fatalf("Failed to load extractor on iteration %d: %v", i, err)
		}
	}
	
	// Check final metrics
	finalMetrics := loader.GetMetrics()
	
	// Should have 1 miss (first load) and 4 hits (subsequent loads)
	expectedMisses := initialMetrics.CacheMisses + 1
	expectedHits := initialMetrics.CacheHits + 4
	
	if finalMetrics.CacheMisses != expectedMisses {
		t.Errorf("Expected %d cache misses, got %d", expectedMisses, finalMetrics.CacheMisses)
	}
	
	if finalMetrics.CacheHits != expectedHits {
		t.Errorf("Expected %d cache hits, got %d", expectedHits, finalMetrics.CacheHits)
	}
	
	if finalMetrics.LoadSuccesses <= initialMetrics.LoadSuccesses {
		t.Error("Expected increase in load successes")
	}
	
	if finalMetrics.AverageLoadTime == 0 {
		t.Error("Expected non-zero average load time")
	}
}

func TestExtractorLoaderClearCache(t *testing.T) {
	registry := custom.NewRegistryManager()
	loader := NewExtractorLoader(DefaultLoaderConfig(), registry)
	defer loader.Close()
	
	// Create and register test extractor
	testExtractor := &custom.CustomExtractor{
		Domain: "clear-test.com",
		Title: map[string]interface{}{
			"selectors": []string{"h1"},
		},
	}
	registry.Register(testExtractor)
	
	// Load extractor to populate cache
	_, err := loader.LoadExtractor("clear-test.com")
	if err != nil {
		t.Fatalf("Failed to load extractor: %v", err)
	}
	
	// Check cache has entry
	size, _, _ := loader.GetCacheStats()
	if size == 0 {
		t.Error("Expected cache to have entries")
	}
	
	// Clear cache
	loader.ClearCache()
	
	// Check cache is empty
	size, _, _ = loader.GetCacheStats()
	if size != 0 {
		t.Errorf("Expected empty cache after clear, got size %d", size)
	}
}

func TestExtractorLoaderWarmupCache(t *testing.T) {
	registry := custom.NewRegistryManager()
	loader := NewExtractorLoader(DefaultLoaderConfig(), registry)
	defer loader.Close()
	
	// Create test extractors
	domains := []string{"warmup1.com", "warmup2.com", "warmup3.com"}
	for _, domain := range domains {
		extractor := &custom.CustomExtractor{
			Domain: domain,
			Title: map[string]interface{}{
				"selectors": []string{"h1"},
			},
		}
		registry.Register(extractor)
	}
	
	// Warmup cache
	err := loader.WarmupCache(domains)
	if err != nil {
		t.Fatalf("Warmup failed: %v", err)
	}
	
	// Check that cache has entries
	size, _, _ := loader.GetCacheStats()
	if size != len(domains) {
		t.Errorf("Expected cache size %d after warmup, got %d", len(domains), size)
	}
}

func TestExtractorLoaderFailedLoad(t *testing.T) {
	registry := custom.NewRegistryManager()
	loader := NewExtractorLoader(DefaultLoaderConfig(), registry)
	defer loader.Close()
	
	// Try to load non-existent extractor
	_, err := loader.LoadExtractor("nonexistent.com")
	if err == nil {
		t.Error("Expected error for non-existent extractor")
	}
	
	// Check that failure was recorded in metrics
	metrics := loader.GetMetrics()
	if metrics.LoadFailures == 0 {
		t.Error("Expected load failure to be recorded in metrics")
	}
}

func TestExtractorLoaderCacheHitRate(t *testing.T) {
	registry := custom.NewRegistryManager()
	loader := NewExtractorLoader(DefaultLoaderConfig(), registry)
	defer loader.Close()
	
	// Create test extractor
	testExtractor := &custom.CustomExtractor{
		Domain: "hitrate-test.com",
		Title: map[string]interface{}{
			"selectors": []string{"h1"},
		},
	}
	registry.Register(testExtractor)
	
	// Load extractor 10 times
	for i := 0; i < 10; i++ {
		_, err := loader.LoadExtractor("hitrate-test.com")
		if err != nil {
			t.Fatalf("Failed to load extractor on iteration %d: %v", i, err)
		}
	}
	
	// Check hit rate
	_, _, hitRate := loader.GetCacheStats()
	expectedHitRate := 9.0 / 10.0 // 1 miss, 9 hits
	
	if hitRate < expectedHitRate-0.01 || hitRate > expectedHitRate+0.01 {
		t.Errorf("Expected hit rate around %.2f, got %.2f", expectedHitRate, hitRate)
	}
}

// Benchmark tests
func BenchmarkExtractorLoaderCacheHit(b *testing.B) {
	registry := custom.NewRegistryManager()
	loader := NewExtractorLoader(DefaultLoaderConfig(), registry)
	defer loader.Close()
	
	// Create test extractor
	testExtractor := &custom.CustomExtractor{
		Domain: "benchmark.com",
		Title: map[string]interface{}{
			"selectors": []string{"h1"},
		},
	}
	registry.Register(testExtractor)
	
	// Prime the cache
	loader.LoadExtractor("benchmark.com")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadExtractor("benchmark.com")
		if err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}
}

func BenchmarkExtractorLoaderCacheMiss(b *testing.B) {
	registry := custom.NewRegistryManager()
	loader := NewExtractorLoader(DefaultLoaderConfig(), registry)
	defer loader.Close()
	
	// Create many test extractors
	for i := 0; i < b.N; i++ {
		domain := fmt.Sprintf("benchmark%d.com", i)
		extractor := &custom.CustomExtractor{
			Domain: domain,
			Title: map[string]interface{}{
				"selectors": []string{"h1"},
			},
		}
		registry.Register(extractor)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := fmt.Sprintf("benchmark%d.com", i)
		_, err := loader.LoadExtractor(domain)
		if err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}
}