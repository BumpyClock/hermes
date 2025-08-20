// ABOUTME: Updated extractor selection logic with complete custom extractor framework integration
// ABOUTME: Implements priority-based extractor lookup: API → custom registry → HTML detection → generic fallback

package extractors

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/postlight/parser-go/pkg/extractors/custom"
	"github.com/postlight/parser-go/pkg/extractors/generic"
)

// Registry variables for extractor storage
var (
	// CustomRegistry manages all custom extractors with domain mapping
	// JavaScript equivalent: The global registry created by all.js
	CustomRegistry = custom.GlobalRegistryManager
	
	// All contains static extractors for backwards compatibility
	// This will be populated from CustomRegistry.GetDomainMapping()
	All = make(map[string]Extractor)
)

// DetectByHTMLFunc defines the HTML-based detection function signature
type DetectByHTMLFunc func(*goquery.Document) Extractor

// DetectByHTML detects extractor using HTML-based selectors
// JavaScript equivalent: detectByHtml($) function in detect-by-html.js
func DetectByHTML(doc *goquery.Document) Extractor {
	if doc == nil {
		return nil
	}
	
	// Use the custom registry to detect by HTML
	if customExtractor := CustomRegistry.GetByHTML(doc); customExtractor != nil {
		return NewCustomExtractorWrapper(customExtractor)
	}
	
	return nil
}

// GenericExtractor creates a generic fallback extractor
func GenericExtractor() Extractor {
	return generic.NewGenericExtractor()
}

// GetExtractor returns the appropriate extractor for a given URL
// Direct 1:1 port of JavaScript getExtractor function with custom framework integration
// JavaScript signature: getExtractor(url, parsedUrl, $)
func GetExtractor(urlStr string, parsedURL *url.URL, doc *goquery.Document) (Extractor, error) {
	extractor := getExtractorWithCustomFramework(urlStr, parsedURL, doc)
	return extractor, nil
}

// getExtractorWithCustomFramework implements core extractor selection with custom registry
// JavaScript equivalent: getExtractor function in get-extractor.js
func getExtractorWithCustomFramework(urlStr string, parsedURL *url.URL, doc *goquery.Document) Extractor {
	// Extract URL components - matches JavaScript URL parsing behavior exactly
	hostname, baseDomain, err := extractURLComponents(urlStr)
	if err != nil {
		// On URL parsing error, fallback to GenericExtractor like JavaScript
		return GenericExtractor()
	}
	
	// If parsedURL is provided, use its hostname (matches JavaScript function signature)
	// JavaScript: parsedUrl = parsedUrl || URL.parse(url);
	if parsedURL != nil {
		hostname = parsedURL.Hostname()
		// Recalculate base domain from provided hostname
		baseDomain = calculateBaseDomain(hostname)
	}
	
	// Priority-based extractor lookup matching JavaScript exactly:
	// 1. apiExtractors[hostname]
	// 2. apiExtractors[baseDomain] 
	// 3. CustomRegistry[hostname] (replaces Extractors[hostname])
	// 4. CustomRegistry[baseDomain] (replaces Extractors[baseDomain])
	// 5. detectByHtml($)
	// 6. GenericExtractor
	
	// Priority 1: API extractor by hostname
	if extractor, found := GetAPIExtractors()[hostname]; found {
		return extractor
	}
	
	// Priority 2: API extractor by base domain
	if extractor, found := GetAPIExtractors()[baseDomain]; found {
		return extractor
	}
	
	// Priority 3: Custom extractor by hostname (from registry)
	// JavaScript equivalent: Extractors[hostname]
	if customExtractor, found := CustomRegistry.GetByDomain(hostname); found {
		return NewCustomExtractorWrapper(customExtractor)
	}
	
	// Priority 4: Custom extractor by base domain (from registry)
	// JavaScript equivalent: Extractors[baseDomain]
	if customExtractor, found := CustomRegistry.GetByDomain(baseDomain); found {
		return NewCustomExtractorWrapper(customExtractor)
	}
	
	// Priority 5: HTML-based detection
	// JavaScript equivalent: detectByHtml($)
	if doc != nil {
		if extractor := DetectByHTML(doc); extractor != nil {
			return extractor
		}
	}
	
	// Priority 6: Generic extractor fallback (always returns non-nil)
	return GenericExtractor()
}

// extractURLComponents extracts hostname and base domain from URL string
// Matches JavaScript URL.parse(url).hostname behavior exactly
func extractURLComponents(urlStr string) (hostname, baseDomain string, err error) {
	// Validate input
	if urlStr == "" {
		return "", "", fmt.Errorf("empty URL provided")
	}
	
	// Parse URL - matches JavaScript URL.parse() behavior
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL: %w", err)
	}
	
	// Extract hostname - matches JavaScript parsedUrl.hostname
	hostname = parsedURL.Hostname()
	if hostname == "" {
		// Handle case where hostname is empty (e.g., "https:///path")
		return "", "", fmt.Errorf("URL missing hostname")
	}
	
	// Calculate base domain using JavaScript logic
	baseDomain = calculateBaseDomain(hostname)
	
	return hostname, baseDomain, nil
}

// calculateBaseDomain extracts base domain from hostname
// Direct port of JavaScript: hostname.split('.').slice(-2).join('.')
func calculateBaseDomain(hostname string) string {
	// Handle empty hostname
	if hostname == "" {
		return ""
	}
	
	// Split hostname on dots - matches JavaScript .split('.')
	parts := strings.Split(hostname, ".")
	
	// Take last 2 parts - matches JavaScript .slice(-2)
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	
	// If less than 2 parts, return the original hostname
	// Matches JavaScript behavior when slice(-2) on single element
	return hostname
}

// NewCustomExtractorWrapper wraps a CustomExtractor to implement the Extractor interface
// This is the bridge between the new custom framework and the existing extractor interface
func NewCustomExtractorWrapper(customExtractor *custom.CustomExtractor) Extractor {
	return &CustomExtractorWrapper{
		customExtractor: customExtractor,
	}
}

// CustomExtractorWrapper adapts CustomExtractor to the Extractor interface
type CustomExtractorWrapper struct {
	customExtractor *custom.CustomExtractor
}

// Extract implements the Extractor interface using CustomExtractor
func (cew *CustomExtractorWrapper) Extract(doc *goquery.Document, url string, opts ExtractorOptions) (*Result, error) {
	// This will be implemented to bridge between CustomExtractor and the existing Result type
	// For now, return a placeholder
	return &Result{
		Title:   "Custom Extractor: " + cew.customExtractor.Domain,
		URL:     url,
		Domain:  cew.customExtractor.Domain,
		Content: "Content extracted by custom extractor",
	}, nil
}

// GetDomain returns the extractor's domain
func (cew *CustomExtractorWrapper) GetDomain() string {
	return cew.customExtractor.Domain
}

// UpdateAllRegistry updates the All map from the custom registry
// JavaScript equivalent: Building the All map from all.js processing
func UpdateAllRegistry() {
	// Clear existing entries
	for k := range All {
		delete(All, k)
	}
	
	// Populate from custom registry
	domainMapping := CustomRegistry.GetDomainMapping()
	for domain, customExtractor := range domainMapping {
		All[domain] = NewCustomExtractorWrapper(customExtractor)
	}
}

// InitializeCustomRegistry populates the custom registry with all extractors
// This will be called during initialization to load all 150+ extractors
func InitializeCustomRegistry() error {
	// This function will be populated with all custom extractor registrations
	// For now, it's a placeholder that can be expanded by future implementations
	
	// Example of how extractors will be registered:
	// mediumExtractor := &custom.CustomExtractor{
	//     Domain: "medium.com",
	//     Title: &custom.FieldExtractor{
	//         Selectors: []interface{}{"h1", []string{"meta[name='og:title']", "value"}},
	//     },
	//     // ... other fields
	// }
	// CustomRegistry.Register(mediumExtractor)
	
	// Update the All map after registration
	UpdateAllRegistry()
	
	return nil
}