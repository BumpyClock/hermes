// ABOUTME: Complete extractor registry system with domain mapping and JavaScript compatibility
// ABOUTME: Manages all 150+ custom extractors with domain resolution, HTML detection, and runtime extractor addition

package custom

import (
	"fmt"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// ExtractorFactory creates custom extractors
// Used for lazy loading and dynamic creation of extractors
type ExtractorFactory func() *CustomExtractor

// RegistryManager provides thread-safe management of custom extractors
// JavaScript equivalent: Combination of all.js, mergeSupportedDomains, and runtime extractor management
type RegistryManager struct {
	mu                   sync.RWMutex
	extractors           map[string]*CustomExtractor        // Primary domain -> extractor
	domainToExtractor    map[string]*CustomExtractor        // All domains -> extractor (including supported)
	htmlDetectors        map[string]*CustomExtractor        // HTML selector -> extractor
	extractorFactories   map[string]ExtractorFactory        // Domain -> factory for lazy loading
	initialized          map[string]bool                    // Track which extractors are loaded
}

// NewRegistryManager creates a new registry manager
func NewRegistryManager() *RegistryManager {
	return &RegistryManager{
		extractors:         make(map[string]*CustomExtractor),
		domainToExtractor:  make(map[string]*CustomExtractor),
		htmlDetectors:      make(map[string]*CustomExtractor),
		extractorFactories: make(map[string]ExtractorFactory),
		initialized:        make(map[string]bool),
	}
}

// Register adds a custom extractor to the registry
// JavaScript equivalent: Building the registry in all.js with mergeSupportedDomains
func (rm *RegistryManager) Register(extractor *CustomExtractor) error {
	if extractor == nil {
		return fmt.Errorf("cannot register nil extractor")
	}
	
	if extractor.Domain == "" {
		return fmt.Errorf("extractor must have a domain")
	}
	
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	return rm.registerLocked(extractor)
}

// registerLocked performs registration with existing lock
func (rm *RegistryManager) registerLocked(extractor *CustomExtractor) error {
	primaryDomain := extractor.Domain
	
	// Register by primary domain
	rm.extractors[primaryDomain] = extractor
	rm.domainToExtractor[primaryDomain] = extractor
	rm.initialized[primaryDomain] = true
	
	// Register by all supported domains (mergeSupportedDomains behavior)
	// JavaScript: merge(extractor, [extractor.domain, ...extractor.supportedDomains])
	for _, domain := range extractor.SupportedDomains {
		if domain != "" && domain != primaryDomain {
			rm.domainToExtractor[domain] = extractor
		}
	}
	
	return nil
}

// RegisterFactory adds a factory for lazy loading of extractors
// Useful for reducing memory usage when not all extractors are needed
func (rm *RegistryManager) RegisterFactory(domain string, factory ExtractorFactory) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	
	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}
	
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	rm.extractorFactories[domain] = factory
	return nil
}

// RegisterHTMLDetector adds an HTML-based extractor detector
// JavaScript equivalent: Entries in detect-by-html.js Detectors map
func (rm *RegistryManager) RegisterHTMLDetector(selector string, extractor *CustomExtractor) error {
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}
	
	if extractor == nil {
		return fmt.Errorf("extractor cannot be nil")
	}
	
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	rm.htmlDetectors[selector] = extractor
	return nil
}

// GetByDomain retrieves an extractor by domain with lazy loading support
// JavaScript equivalent: Extractors[hostname] || Extractors[baseDomain] lookup in get-extractor.js
func (rm *RegistryManager) GetByDomain(domain string) (*CustomExtractor, bool) {
	if domain == "" {
		return nil, false
	}
	
	rm.mu.RLock()
	
	// Check if already loaded
	if extractor, exists := rm.domainToExtractor[domain]; exists {
		rm.mu.RUnlock()
		return extractor, true
	}
	
	// Check for factory
	factory, hasFactory := rm.extractorFactories[domain]
	rm.mu.RUnlock()
	
	if hasFactory {
		// Lazy load the extractor
		extractor := factory()
		if extractor != nil {
			rm.Register(extractor) // This will handle locking
			return extractor, true
		}
	}
	
	return nil, false
}

// GetByHTML detects extractor using HTML selectors
// JavaScript equivalent: detectByHtml($) function in detect-by-html.js
func (rm *RegistryManager) GetByHTML(doc *goquery.Document) *CustomExtractor {
	if doc == nil {
		return nil
	}
	
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Try each registered HTML detector
	// JavaScript: Reflect.ownKeys(Detectors).find(s => $(s).length > 0)
	for selector, extractor := range rm.htmlDetectors {
		if doc.Find(selector).Length() > 0 {
			return extractor
		}
	}
	
	return nil
}

// GetBaseDomain calculates base domain from hostname
// JavaScript equivalent: hostname.split('.').slice(-2).join('.') in get-extractor.js
func (rm *RegistryManager) GetBaseDomain(hostname string) string {
	if hostname == "" {
		return ""
	}
	
	parts := strings.Split(hostname, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	
	return hostname
}

// GetByDomainWithFallback tries hostname first, then base domain
// JavaScript equivalent: Extractors[hostname] || Extractors[baseDomain] logic
func (rm *RegistryManager) GetByDomainWithFallback(hostname string) (*CustomExtractor, bool) {
	// Try exact hostname first
	if extractor, exists := rm.GetByDomain(hostname); exists {
		return extractor, true
	}
	
	// Try base domain
	baseDomain := rm.GetBaseDomain(hostname)
	if baseDomain != hostname {
		return rm.GetByDomain(baseDomain)
	}
	
	return nil, false
}

// GetAll returns all registered extractors (deduplicated by primary domain)
// JavaScript equivalent: Object.keys(CustomExtractors) processing in all.js
func (rm *RegistryManager) GetAll() map[string]*CustomExtractor {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Return copy to prevent external modification
	result := make(map[string]*CustomExtractor)
	for domain, extractor := range rm.extractors {
		result[domain] = extractor
	}
	
	return result
}

// GetDomainMapping returns the complete domain-to-extractor mapping
// JavaScript equivalent: The flattened domain mapping created by all.js + mergeSupportedDomains
func (rm *RegistryManager) GetDomainMapping() map[string]*CustomExtractor {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Return copy of complete domain mapping
	result := make(map[string]*CustomExtractor)
	for domain, extractor := range rm.domainToExtractor {
		result[domain] = extractor
	}
	
	return result
}

// Count returns statistics about registered extractors
func (rm *RegistryManager) Count() (int, int) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	return len(rm.extractors), len(rm.domainToExtractor)
}

// ListDomains returns all registered domains (including supported domains)
func (rm *RegistryManager) ListDomains() []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	domains := make([]string, 0, len(rm.domainToExtractor))
	for domain := range rm.domainToExtractor {
		domains = append(domains, domain)
	}
	
	return domains
}

// ListPrimaryDomains returns only primary domains (not supported domains)
func (rm *RegistryManager) ListPrimaryDomains() []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	domains := make([]string, 0, len(rm.extractors))
	for domain := range rm.extractors {
		domains = append(domains, domain)
	}
	
	return domains
}

// Remove removes an extractor from the registry
// Useful for testing and dynamic management
func (rm *RegistryManager) Remove(domain string) bool {
	if domain == "" {
		return false
	}
	
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	// Find the extractor to remove
	extractor, exists := rm.extractors[domain]
	if !exists {
		return false
	}
	
	// Remove from primary registry
	delete(rm.extractors, domain)
	delete(rm.initialized, domain)
	
	// Remove from domain mapping
	delete(rm.domainToExtractor, domain)
	
	// Remove supported domains
	for _, supportedDomain := range extractor.SupportedDomains {
		if rm.domainToExtractor[supportedDomain] == extractor {
			delete(rm.domainToExtractor, supportedDomain)
		}
	}
	
	// Remove factory if exists
	delete(rm.extractorFactories, domain)
	
	return true
}

// Clear removes all extractors from the registry
// Useful for testing
func (rm *RegistryManager) Clear() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	rm.extractors = make(map[string]*CustomExtractor)
	rm.domainToExtractor = make(map[string]*CustomExtractor)
	rm.htmlDetectors = make(map[string]*CustomExtractor)
	rm.extractorFactories = make(map[string]ExtractorFactory)
	rm.initialized = make(map[string]bool)
}

// Clone creates a copy of the registry
// Useful for testing and isolated environments
func (rm *RegistryManager) Clone() *RegistryManager {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	clone := NewRegistryManager()
	
	// Copy extractors
	for domain, extractor := range rm.extractors {
		clone.extractors[domain] = extractor
	}
	
	// Copy domain mapping
	for domain, extractor := range rm.domainToExtractor {
		clone.domainToExtractor[domain] = extractor
	}
	
	// Copy HTML detectors
	for selector, extractor := range rm.htmlDetectors {
		clone.htmlDetectors[selector] = extractor
	}
	
	// Copy factories
	for domain, factory := range rm.extractorFactories {
		clone.extractorFactories[domain] = factory
	}
	
	// Copy initialization status
	for domain, status := range rm.initialized {
		clone.initialized[domain] = status
	}
	
	return clone
}

// MergeSupportedDomains creates domain mappings for an extractor
// JavaScript equivalent: mergeSupportedDomains function in utils/merge-supported-domains.js
func MergeSupportedDomains(extractor *CustomExtractor) map[string]*CustomExtractor {
	result := make(map[string]*CustomExtractor)
	
	if extractor == nil {
		return result
	}
	
	// Add primary domain
	result[extractor.Domain] = extractor
	
	// Add supported domains
	for _, domain := range extractor.SupportedDomains {
		if domain != "" {
			result[domain] = extractor
		}
	}
	
	return result
}

// BuildAllExtractorsMap creates the complete domain mapping
// JavaScript equivalent: The result of all.js processing
func BuildAllExtractorsMap(extractors []*CustomExtractor) map[string]*CustomExtractor {
	result := make(map[string]*CustomExtractor)
	
	for _, extractor := range extractors {
		// Merge supported domains for this extractor
		domainMap := MergeSupportedDomains(extractor)
		
		// Add to result map
		for domain, ext := range domainMap {
			result[domain] = ext
		}
	}
	
	return result
}

// Default global registry instance
// JavaScript equivalent: The implicit global registry used throughout the codebase
var GlobalRegistryManager = NewRegistryManager()