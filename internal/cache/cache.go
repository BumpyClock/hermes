// ABOUTME: Thread-safe caching layer using sync.Map for high-performance concurrent access
// ABOUTME: Caches DOM elements, selector results, and extraction results to avoid repeated computation

package cache

import (
	"hash/fnv"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Value     interface{} `json:"value"`
	CreatedAt time.Time   `json:"created_at"`
	AccessCount int64     `json:"access_count"`
	LastAccess  time.Time `json:"last_access"`
	TTL        time.Duration `json:"ttl,omitempty"`
}

// DOMCache provides thread-safe caching for DOM-related operations
type DOMCache struct {
	selectorResults sync.Map // map[string]*CacheEntry - selector query results
	elementCache    sync.Map // map[string]*CacheEntry - cached DOM elements
	textCache       sync.Map // map[string]*CacheEntry - extracted text content
	attributeCache  sync.Map // map[string]*CacheEntry - element attributes
	stats           CacheStats
	mutex           sync.RWMutex
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	Hits            int64   `json:"hits"`
	Misses          int64   `json:"misses"`
	Sets            int64   `json:"sets"`
	Evictions       int64   `json:"evictions"`
	HitRatio        float64 `json:"hit_ratio"`
	TotalEntries    int64   `json:"total_entries"`
	MemoryUsageKB   int64   `json:"memory_usage_kb"`
	LastCleanup     time.Time `json:"last_cleanup"`
}

// SelectorCacheKey generates a cache key for selector operations
type SelectorCacheKey struct {
	DocumentHash string
	Selector     string
	Operation    string // "find", "text", "attr", etc.
	Attribute    string // for attribute operations
}

// NewDOMCache creates a new thread-safe DOM cache
func NewDOMCache() *DOMCache {
	return &DOMCache{
		stats: CacheStats{
			LastCleanup: time.Now(),
		},
	}
}

// generateKey creates a fast hash-based cache key
func (key SelectorCacheKey) String() string {
	h := fnv.New64a()
	h.Write([]byte(key.DocumentHash))
	h.Write([]byte(key.Selector))
	h.Write([]byte(key.Operation))
	h.Write([]byte(key.Attribute))
	return strconv.FormatUint(h.Sum64(), 36)
}

// GetSelectorResult retrieves cached selector query results
func (dc *DOMCache) GetSelectorResult(key SelectorCacheKey) (*goquery.Selection, bool) {
	cacheKey := key.String()
	if entry, ok := dc.selectorResults.Load(cacheKey); ok {
		if cacheEntry, ok := entry.(*CacheEntry); ok {
			// Update access statistics
			dc.updateAccess(cacheEntry)
			dc.incrementStat("hits")
			
			if selection, ok := cacheEntry.Value.(*goquery.Selection); ok {
				return selection, true
			}
		}
	}
	
	dc.incrementStat("misses")
	return nil, false
}

// SetSelectorResult caches selector query results
func (dc *DOMCache) SetSelectorResult(key SelectorCacheKey, selection *goquery.Selection, ttl time.Duration) {
	cacheKey := key.String()
	entry := &CacheEntry{
		Value:       selection,
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
		TTL:         ttl,
	}
	
	dc.selectorResults.Store(cacheKey, entry)
	dc.incrementStat("sets")
}

// GetTextContent retrieves cached text extraction results
func (dc *DOMCache) GetTextContent(documentHash, selector string) (string, bool) {
	key := SelectorCacheKey{
		DocumentHash: documentHash,
		Selector:     selector,
		Operation:    "text",
	}
	
	cacheKey := key.String()
	if entry, ok := dc.textCache.Load(cacheKey); ok {
		if cacheEntry, ok := entry.(*CacheEntry); ok {
			dc.updateAccess(cacheEntry)
			dc.incrementStat("hits")
			
			if text, ok := cacheEntry.Value.(string); ok {
				return text, true
			}
		}
	}
	
	dc.incrementStat("misses")
	return "", false
}

// SetTextContent caches text extraction results
func (dc *DOMCache) SetTextContent(documentHash, selector, text string, ttl time.Duration) {
	key := SelectorCacheKey{
		DocumentHash: documentHash,
		Selector:     selector,
		Operation:    "text",
	}
	
	cacheKey := key.String()
	entry := &CacheEntry{
		Value:       text,
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
		TTL:         ttl,
	}
	
	dc.textCache.Store(cacheKey, entry)
	dc.incrementStat("sets")
}

// GetAttribute retrieves cached attribute values
func (dc *DOMCache) GetAttribute(documentHash, selector, attribute string) (string, bool) {
	key := SelectorCacheKey{
		DocumentHash: documentHash,
		Selector:     selector,
		Operation:    "attr",
		Attribute:    attribute,
	}
	
	cacheKey := key.String()
	if entry, ok := dc.attributeCache.Load(cacheKey); ok {
		if cacheEntry, ok := entry.(*CacheEntry); ok {
			dc.updateAccess(cacheEntry)
			dc.incrementStat("hits")
			
			if attrValue, ok := cacheEntry.Value.(string); ok {
				return attrValue, true
			}
		}
	}
	
	dc.incrementStat("misses")
	return "", false
}

// SetAttribute caches attribute values
func (dc *DOMCache) SetAttribute(documentHash, selector, attribute, value string, ttl time.Duration) {
	key := SelectorCacheKey{
		DocumentHash: documentHash,
		Selector:     selector,
		Operation:    "attr",
		Attribute:    attribute,
	}
	
	cacheKey := key.String()
	entry := &CacheEntry{
		Value:       value,
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
		TTL:         ttl,
	}
	
	dc.attributeCache.Store(cacheKey, entry)
	dc.incrementStat("sets")
}

// updateAccess updates access statistics for a cache entry
func (dc *DOMCache) updateAccess(entry *CacheEntry) {
	entry.AccessCount++
	entry.LastAccess = time.Now()
}

// incrementStat safely increments a statistic
func (dc *DOMCache) incrementStat(stat string) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	
	switch stat {
	case "hits":
		dc.stats.Hits++
	case "misses":
		dc.stats.Misses++
	case "sets":
		dc.stats.Sets++
	case "evictions":
		dc.stats.Evictions++
	}
	
	// Update hit ratio
	total := dc.stats.Hits + dc.stats.Misses
	if total > 0 {
		dc.stats.HitRatio = float64(dc.stats.Hits) / float64(total)
	}
}

// GetStats returns current cache statistics
func (dc *DOMCache) GetStats() CacheStats {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	
	// Count total entries
	totalEntries := int64(0)
	dc.selectorResults.Range(func(_, _ interface{}) bool {
		totalEntries++
		return true
	})
	dc.textCache.Range(func(_, _ interface{}) bool {
		totalEntries++
		return true
	})
	dc.attributeCache.Range(func(_, _ interface{}) bool {
		totalEntries++
		return true
	})
	
	stats := dc.stats
	stats.TotalEntries = totalEntries
	return stats
}

// CleanupExpired removes expired cache entries
func (dc *DOMCache) CleanupExpired() int {
	now := time.Now()
	evicted := 0
	
	// Clean selector results
	dc.selectorResults.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*CacheEntry); ok {
			if entry.TTL > 0 && now.Sub(entry.CreatedAt) > entry.TTL {
				dc.selectorResults.Delete(key)
				evicted++
			}
		}
		return true
	})
	
	// Clean text cache
	dc.textCache.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*CacheEntry); ok {
			if entry.TTL > 0 && now.Sub(entry.CreatedAt) > entry.TTL {
				dc.textCache.Delete(key)
				evicted++
			}
		}
		return true
	})
	
	// Clean attribute cache
	dc.attributeCache.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*CacheEntry); ok {
			if entry.TTL > 0 && now.Sub(entry.CreatedAt) > entry.TTL {
				dc.attributeCache.Delete(key)
				evicted++
			}
		}
		return true
	})
	
	// Update statistics
	dc.mutex.Lock()
	dc.stats.Evictions += int64(evicted)
	dc.stats.LastCleanup = now
	dc.mutex.Unlock()
	
	return evicted
}

// Clear removes all cache entries
func (dc *DOMCache) Clear() {
	dc.selectorResults = sync.Map{}
	dc.textCache = sync.Map{}
	dc.attributeCache = sync.Map{}
	
	dc.mutex.Lock()
	dc.stats = CacheStats{
		LastCleanup: time.Now(),
	}
	dc.mutex.Unlock()
}

// ExtractionCache provides thread-safe caching for extraction results
type ExtractionCache struct {
	results sync.Map // map[string]*CacheEntry - full extraction results
	fields  sync.Map // map[string]*CacheEntry - individual field results
	stats   CacheStats
	mutex   sync.RWMutex
}

// NewExtractionCache creates a new extraction result cache
func NewExtractionCache() *ExtractionCache {
	return &ExtractionCache{
		stats: CacheStats{
			LastCleanup: time.Now(),
		},
	}
}

// GetExtractionResult retrieves cached extraction results for a URL
func (ec *ExtractionCache) GetExtractionResult(url string) (interface{}, bool) {
	if entry, ok := ec.results.Load(url); ok {
		if cacheEntry, ok := entry.(*CacheEntry); ok {
			ec.updateAccessExtraction(cacheEntry)
			ec.incrementStatExtraction("hits")
			return cacheEntry.Value, true
		}
	}
	
	ec.incrementStatExtraction("misses")
	return nil, false
}

// SetExtractionResult caches extraction results for a URL
func (ec *ExtractionCache) SetExtractionResult(url string, result interface{}, ttl time.Duration) {
	entry := &CacheEntry{
		Value:       result,
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
		TTL:         ttl,
	}
	
	ec.results.Store(url, entry)
	ec.incrementStatExtraction("sets")
}

// GetFieldResult retrieves cached field extraction results
func (ec *ExtractionCache) GetFieldResult(url, field string) (interface{}, bool) {
	key := url + "::" + field
	if entry, ok := ec.fields.Load(key); ok {
		if cacheEntry, ok := entry.(*CacheEntry); ok {
			ec.updateAccessExtraction(cacheEntry)
			ec.incrementStatExtraction("hits")
			return cacheEntry.Value, true
		}
	}
	
	ec.incrementStatExtraction("misses")
	return nil, false
}

// SetFieldResult caches individual field extraction results
func (ec *ExtractionCache) SetFieldResult(url, field string, result interface{}, ttl time.Duration) {
	key := url + "::" + field
	entry := &CacheEntry{
		Value:       result,
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
		TTL:         ttl,
	}
	
	ec.fields.Store(key, entry)
	ec.incrementStatExtraction("sets")
}

// Helper methods for ExtractionCache
func (ec *ExtractionCache) updateAccessExtraction(entry *CacheEntry) {
	entry.AccessCount++
	entry.LastAccess = time.Now()
}

func (ec *ExtractionCache) incrementStatExtraction(stat string) {
	ec.mutex.Lock()
	defer ec.mutex.Unlock()
	
	switch stat {
	case "hits":
		ec.stats.Hits++
	case "misses":
		ec.stats.Misses++
	case "sets":
		ec.stats.Sets++
	case "evictions":
		ec.stats.Evictions++
	}
	
	total := ec.stats.Hits + ec.stats.Misses
	if total > 0 {
		ec.stats.HitRatio = float64(ec.stats.Hits) / float64(total)
	}
}

// Global cache instances
var (
	GlobalDOMCache        = NewDOMCache()
	GlobalExtractionCache = NewExtractionCache()
)

// CacheManager coordinates multiple cache types
type CacheManager struct {
	domCache        *DOMCache
	extractionCache *ExtractionCache
	cleanupTicker   *time.Ticker
	stopCleanup     chan bool
}

// NewCacheManager creates a new cache manager with automatic cleanup
func NewCacheManager(cleanupInterval time.Duration) *CacheManager {
	cm := &CacheManager{
		domCache:        NewDOMCache(),
		extractionCache: NewExtractionCache(),
		cleanupTicker:   time.NewTicker(cleanupInterval),
		stopCleanup:     make(chan bool, 1),
	}
	
	// Start automatic cleanup goroutine
	go cm.runCleanup()
	
	return cm
}

// runCleanup runs periodic cache cleanup
func (cm *CacheManager) runCleanup() {
	for {
		select {
		case <-cm.cleanupTicker.C:
			cm.domCache.CleanupExpired()
			cm.extractionCache.CleanupExpired()
		case <-cm.stopCleanup:
			cm.cleanupTicker.Stop()
			return
		}
	}
}

// GetDOMCache returns the DOM cache instance
func (cm *CacheManager) GetDOMCache() *DOMCache {
	return cm.domCache
}

// GetExtractionCache returns the extraction cache instance
func (cm *CacheManager) GetExtractionCache() *ExtractionCache {
	return cm.extractionCache
}

// Stop stops the cache manager and cleanup goroutine
func (cm *CacheManager) Stop() {
	cm.stopCleanup <- true
}

// GetAllStats returns statistics for all caches
func (cm *CacheManager) GetAllStats() map[string]CacheStats {
	return map[string]CacheStats{
		"dom":        cm.domCache.GetStats(),
		"extraction": cm.extractionCache.GetStats(),
	}
}

// CleanupExpired method for ExtractionCache
func (ec *ExtractionCache) CleanupExpired() int {
	now := time.Now()
	evicted := 0
	
	// Clean extraction results
	ec.results.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*CacheEntry); ok {
			if entry.TTL > 0 && now.Sub(entry.CreatedAt) > entry.TTL {
				ec.results.Delete(key)
				evicted++
			}
		}
		return true
	})
	
	// Clean field results
	ec.fields.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*CacheEntry); ok {
			if entry.TTL > 0 && now.Sub(entry.CreatedAt) > entry.TTL {
				ec.fields.Delete(key)
				evicted++
			}
		}
		return true
	})
	
	ec.mutex.Lock()
	ec.stats.Evictions += int64(evicted)
	ec.stats.LastCleanup = now
	ec.mutex.Unlock()
	
	return evicted
}

// GetStats returns cache statistics for ExtractionCache
func (ec *ExtractionCache) GetStats() CacheStats {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	
	totalEntries := int64(0)
	ec.results.Range(func(_, _ interface{}) bool {
		totalEntries++
		return true
	})
	ec.fields.Range(func(_, _ interface{}) bool {
		totalEntries++
		return true
	})
	
	stats := ec.stats
	stats.TotalEntries = totalEntries
	return stats
}