// ABOUTME: Main extractor selection logic that maps URLs to appropriate extractors with 100% JavaScript compatibility
// ABOUTME: Implements priority-based extractor lookup: API extractors → static extractors → HTML detection → generic fallback

package extractors

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/parser-go/pkg/extractors/custom"
	"github.com/BumpyClock/parser-go/pkg/extractors/generic"
	"github.com/BumpyClock/parser-go/pkg/parser"
)

// DetectByHTMLFunc type for HTML-based extractor detection
type DetectByHTMLFunc func(*goquery.Document) Extractor

// Registry variables for extractor storage
// These are now properly integrated with the custom extractor framework
var (
	// All contains all static custom extractors, populated from custom registry
	// JavaScript equivalent: The result of all.js processing
	All = make(map[string]Extractor)
	
	// CustomRegistry manages all custom extractors with domain mapping
	CustomRegistry = custom.GlobalRegistryManager
)

// GenericExtractor creates a generic fallback extractor
// Returns a basic implementation that satisfies the Extractor interface
func GenericExtractor() Extractor {
	return &BasicGenericExtractor{domain: "*"}
}

// BasicGenericExtractor implements the Extractor interface for generic extraction
type BasicGenericExtractor struct {
	domain string
}

// GetDomain returns the domain this extractor handles
func (bge *BasicGenericExtractor) GetDomain() string {
	return bge.domain
}

// Extract performs generic extraction by delegating to the generic package
func (bge *BasicGenericExtractor) Extract(doc *goquery.Document, url string, opts parser.ExtractorOptions) (*parser.Result, error) {
	// Create generic extractor from the generic package
	genericExtractor := generic.NewGenericExtractor()
	
	// Convert to generic extraction options
	genericOpts := &generic.ExtractionOptions{
		URL:      url,
		Doc:      doc,
		HTML:     opts.HTML,
		Fallback: opts.Fallback,
	}
	
	// Perform generic extraction
	result, err := genericExtractor.ExtractGeneric(genericOpts)
	if err != nil {
		return nil, err
	}
	
	// Convert to parser.Result
	parserResult := &parser.Result{
		URL:           result.URL,
		Domain:        result.Domain,
		Title:         result.Title,
		Author:        result.Author,
		Content:       result.Content,
		DatePublished: result.DatePublished,
		LeadImageURL:  result.LeadImageURL,
		Dek:           result.Dek,
		NextPageURL:   result.NextPageURL,
		Excerpt:       result.Excerpt,
		WordCount:     result.WordCount,
		Direction:     result.Direction,
	}
	
	return parserResult, nil
}

// GetAPIExtractors returns all runtime-registered extractors as Extractor interface
func GetAPIExtractors() map[string]Extractor {
	impl := GetAPIExtractorsImpl()
	result := make(map[string]Extractor)
	for domain, extractor := range impl {
		result[domain] = extractor
	}
	return result
}

// GetExtractor returns the appropriate extractor for a given URL
// Direct 1:1 port of JavaScript getExtractor function with identical behavior
// 
// JavaScript signature: getExtractor(url, parsedUrl, $)
// Go signature: GetExtractor(url, parsedUrl, doc) - $ becomes doc for goquery compatibility
func GetExtractor(urlStr string, parsedURL *url.URL, doc *goquery.Document) (Extractor, error) {
	extractor := getExtractorWithRegistries(urlStr, parsedURL, doc, GetAPIExtractors(), All, DetectByHTML)
	return extractor, nil
}

// getExtractorWithRegistries allows dependency injection for testing
// Internal function that implements the core extractor selection logic
func getExtractorWithRegistries(
	urlStr string, 
	parsedURL *url.URL, 
	doc *goquery.Document,
	apiExtractors map[string]Extractor,
	staticExtractors map[string]Extractor,
	detectByHtml DetectByHTMLFunc,
) Extractor {
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
	// 3. Extractors[hostname]
	// 4. Extractors[baseDomain]
	// 5. detectByHtml($)
	// 6. GenericExtractor
	
	// Priority 1: API extractor by hostname
	if extractor, found := apiExtractors[hostname]; found {
		return extractor
	}
	
	// Priority 2: API extractor by base domain
	if extractor, found := apiExtractors[baseDomain]; found {
		return extractor
	}
	
	// Priority 3: Static extractor by hostname
	if extractor, found := staticExtractors[hostname]; found {
		return extractor
	}
	
	// Priority 4: Static extractor by base domain
	if extractor, found := staticExtractors[baseDomain]; found {
		return extractor
	}
	
	// Priority 5: HTML-based detection
	if doc != nil && detectByHtml != nil {
		if extractor := detectByHtml(doc); extractor != nil {
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