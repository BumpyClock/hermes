// ABOUTME: Helper functions for optimized DOM operations using the existing cache system.
// These functions provide convenient wrappers around the cache API for common DOM operations.
package cache

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// CachedElementOperations provides optimized DOM operations using the existing cache system
type CachedElementOperations struct {
	cache *DOMCache
}

// NewCachedElementOperations creates a new cached element operations helper
func NewCachedElementOperations() *CachedElementOperations {
	return &CachedElementOperations{
		cache: GlobalDOMCache,
	}
}

// Global cached operations instance
var GlobalCachedOps = NewCachedElementOperations()

// generateDocumentHash creates a hash for a document/element to use as cache key
func generateDocumentHash(element *goquery.Selection) string {
	if element.Length() == 0 {
		return "empty"
	}

	// Use element's HTML content to create a unique identifier
	html, err := element.Html()
	if err != nil {
		// Fallback to a simple identifier
		return fmt.Sprintf("element_%p", element)
	}

	// Create MD5 hash of content (limited for performance)
	content := html
	if len(content) > 500 {
		content = content[:500]
	}

	hasher := md5.New()
	hasher.Write([]byte(content))
	return fmt.Sprintf("%x", hasher.Sum(nil))[:16]
}

// CachedFind performs a cached selector query
func (ceo *CachedElementOperations) CachedFind(element *goquery.Selection, selector string) *goquery.Selection {
	if element.Length() == 0 {
		return element
	}

	documentHash := generateDocumentHash(element)
	cacheKey := SelectorCacheKey{
		DocumentHash: documentHash,
		Selector:     selector,
		Operation:    "find",
	}

	// Try to get from cache
	if cached, found := ceo.cache.GetSelectorResult(cacheKey); found {
		return cached
	}

	// Execute query and cache result
	result := element.Find(selector)
	ceo.cache.SetSelectorResult(cacheKey, result, time.Hour) // Cache for 1 hour

	return result
}

// CachedText gets cached text content for an element
func (ceo *CachedElementOperations) CachedText(element *goquery.Selection) string {
	if element.Length() == 0 {
		return ""
	}

	documentHash := generateDocumentHash(element)
	selector := "self" // Use "self" to indicate text of the element itself

	// Try to get from cache
	if cached, found := ceo.cache.GetTextContent(documentHash, selector); found {
		return cached
	}

	// Extract text and cache it
	text := strings.TrimSpace(element.Text())
	ceo.cache.SetTextContent(documentHash, selector, text, time.Hour)

	return text
}

// CachedAttr gets cached attribute value for an element
func (ceo *CachedElementOperations) CachedAttr(element *goquery.Selection, attrName string) (string, bool) {
	if element.Length() == 0 {
		return "", false
	}

	documentHash := generateDocumentHash(element)
	selector := "self" // Use "self" to indicate the element itself

	// Check cache - we store "exists|value" format
	cacheKey := fmt.Sprintf("%s_%s", attrName, "exists")
	if cached, found := ceo.cache.GetAttribute(documentHash, selector, cacheKey); found {
		parts := strings.SplitN(cached, "|", 2)
		if len(parts) == 2 {
			exists := parts[0] == "true"
			value := parts[1]
			return value, exists
		}
	}

	// Get attribute and cache result
	attrValue, exists := element.Attr(attrName)
	cacheValue := fmt.Sprintf("%t|%s", exists, attrValue)
	ceo.cache.SetAttribute(documentHash, selector, cacheKey, cacheValue, time.Hour)

	return attrValue, exists
}

// CachedHasClass checks if an element has a specific class using cached attributes
func (ceo *CachedElementOperations) CachedHasClass(element *goquery.Selection, className string) bool {
	class, exists := ceo.CachedAttr(element, "class")
	if !exists {
		return false
	}

	classes := strings.Fields(class)
	for _, cls := range classes {
		if cls == className {
			return true
		}
	}
	return false
}

// CachedChildren gets cached children for an element
func (ceo *CachedElementOperations) CachedChildren(element *goquery.Selection) *goquery.Selection {
	return ceo.CachedFind(element, "> *") // Direct children selector
}

// CachedParent gets cached parent for an element
func (ceo *CachedElementOperations) CachedParent(element *goquery.Selection) *goquery.Selection {
	if element.Length() == 0 {
		return element
	}

	// For parent, we can't really cache effectively since it depends on DOM structure
	// But we can optimize by avoiding repeated calls
	return element.Parent()
}

// BatchCachedFind performs multiple selector queries efficiently
func (ceo *CachedElementOperations) BatchCachedFind(element *goquery.Selection, selectors []string) map[string]*goquery.Selection {
	results := make(map[string]*goquery.Selection, len(selectors))
	
	if element.Length() == 0 {
		// Return empty selections for all selectors
		for _, selector := range selectors {
			results[selector] = element
		}
		return results
	}

	documentHash := generateDocumentHash(element)

	// Check cache for all selectors first
	uncachedSelectors := make([]string, 0, len(selectors))
	for _, selector := range selectors {
		cacheKey := SelectorCacheKey{
			DocumentHash: documentHash,
			Selector:     selector,
			Operation:    "find",
		}

		if cached, found := ceo.cache.GetSelectorResult(cacheKey); found {
			results[selector] = cached
		} else {
			uncachedSelectors = append(uncachedSelectors, selector)
		}
	}

	// Execute and cache uncached selectors
	for _, selector := range uncachedSelectors {
		result := element.Find(selector)
		results[selector] = result

		cacheKey := SelectorCacheKey{
			DocumentHash: documentHash,
			Selector:     selector,
			Operation:    "find",
		}
		ceo.cache.SetSelectorResult(cacheKey, result, time.Hour)
	}

	return results
}

// OptimizedLinkDensity calculates link density using cached operations
func (ceo *CachedElementOperations) OptimizedLinkDensity(element *goquery.Selection) float64 {
	totalText := ceo.CachedText(element)
	if len(totalText) == 0 {
		return 0
	}

	// Use cached find for links
	links := ceo.CachedFind(element, "a")
	if links.Length() == 0 {
		return 0
	}

	var linkTextLength int
	links.Each(func(index int, link *goquery.Selection) {
		linkText := ceo.CachedText(link)
		linkTextLength += len(strings.TrimSpace(linkText))
	})

	totalLength := len(totalText)
	if totalLength == 0 {
		return 0
	}

	return float64(linkTextLength) / float64(totalLength)
}

// ClearElementCache clears the cache for better memory management
func (ceo *CachedElementOperations) ClearElementCache() {
	ceo.cache.Clear()
}

// GetCacheStats returns cache performance statistics
func (ceo *CachedElementOperations) GetCacheStats() CacheStats {
	return ceo.cache.GetStats()
}

// Global helper functions for easy access
func CachedFind(element *goquery.Selection, selector string) *goquery.Selection {
	return GlobalCachedOps.CachedFind(element, selector)
}

func CachedText(element *goquery.Selection) string {
	return GlobalCachedOps.CachedText(element)
}

func CachedAttr(element *goquery.Selection, attrName string) (string, bool) {
	return GlobalCachedOps.CachedAttr(element, attrName)
}

func CachedHasClass(element *goquery.Selection, className string) bool {
	return GlobalCachedOps.CachedHasClass(element, className)
}

func OptimizedLinkDensity(element *goquery.Selection) float64 {
	return GlobalCachedOps.OptimizedLinkDensity(element)
}

func BatchCachedFind(element *goquery.Selection, selectors []string) map[string]*goquery.Selection {
	return GlobalCachedOps.BatchCachedFind(element, selectors)
}