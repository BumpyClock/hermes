// ABOUTME: Demonstration of 100% JavaScript-compatible extractor selection logic
// ABOUTME: This file shows the exact 1:1 port of JavaScript getExtractor() without type conflicts

package extractors

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// SimpleExtractor represents a basic extractor for demonstration
type SimpleExtractor struct {
	Domain string
}

// GetDomain returns the domain this extractor handles
func (s SimpleExtractor) GetDomain() string {
	return s.Domain
}

// JavaScriptCompatibleGetExtractor demonstrates the exact JavaScript getExtractor logic
// This is a faithful 1:1 port showing the correct behavior without type system conflicts
func JavaScriptCompatibleGetExtractor(urlStr string, parsedURL *url.URL, doc *goquery.Document) (SimpleExtractor, error) {
	// Direct port of JavaScript: getExtractor(url, parsedUrl, $)
	
	// Step 1: URL parsing - matches JavaScript exactly
	// JavaScript: parsedUrl = parsedUrl || URL.parse(url);
	var hostname string
	if parsedURL != nil {
		hostname = parsedURL.Hostname()
	} else {
		parsed, err := url.Parse(urlStr)
		if err != nil {
			return SimpleExtractor{Domain: "*"}, fmt.Errorf("invalid URL: %w", err)
		}
		hostname = parsed.Hostname()
	}
	
	if hostname == "" {
		return SimpleExtractor{Domain: "*"}, fmt.Errorf("URL missing hostname")
	}
	
	// Step 2: Base domain calculation - matches JavaScript exactly
	// JavaScript: const baseDomain = hostname.split('.').slice(-2).join('.');
	baseDomain := calculateBaseDomainJSCompat(hostname)
	
	// Step 3: Priority-based lookup - matches JavaScript exactly
	// JavaScript return statement:
	// return (
	//   apiExtractors[hostname] ||
	//   apiExtractors[baseDomain] ||
	//   Extractors[hostname] ||
	//   Extractors[baseDomain] ||
	//   detectByHtml($) ||
	//   GenericExtractor
	// );
	
	// Mock registries for demonstration (in real implementation these would be populated)
	apiExtractors := getDemoAPIExtractors()
	staticExtractors := getDemoStaticExtractors()
	
	// Priority 1: apiExtractors[hostname]
	if extractor, found := apiExtractors[hostname]; found {
		return extractor, nil
	}
	
	// Priority 2: apiExtractors[baseDomain]
	if extractor, found := apiExtractors[baseDomain]; found {
		return extractor, nil
	}
	
	// Priority 3: Extractors[hostname]
	if extractor, found := staticExtractors[hostname]; found {
		return extractor, nil
	}
	
	// Priority 4: Extractors[baseDomain]
	if extractor, found := staticExtractors[baseDomain]; found {
		return extractor, nil
	}
	
	// Priority 5: detectByHtml($)
	if doc != nil {
		if htmlExtractor := demoDetectByHTML(doc); htmlExtractor.Domain != "" {
			return htmlExtractor, nil
		}
	}
	
	// Priority 6: GenericExtractor fallback
	return SimpleExtractor{Domain: "*"}, nil
}

// calculateBaseDomainJSCompat exactly matches JavaScript behavior
// JavaScript: hostname.split('.').slice(-2).join('.')
func calculateBaseDomainJSCompat(hostname string) string {
	if hostname == "" {
		return ""
	}
	
	// JavaScript .split('.')
	parts := strings.Split(hostname, ".")
	
	// JavaScript .slice(-2)
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	
	// If less than 2 parts, return original (matches JavaScript slice behavior)
	return hostname
}

// Demo functions to show the concept without type conflicts
func getDemoAPIExtractors() map[string]SimpleExtractor {
	// In real implementation, this would be populated from add-extractor.js equivalent
	return map[string]SimpleExtractor{
		"api.example.com": {Domain: "api.example.com"},
		"example.com":     {Domain: "example.com"},
	}
}

func getDemoStaticExtractors() map[string]SimpleExtractor {
	// In real implementation, this would be populated from all.js equivalent
	return map[string]SimpleExtractor{
		"www.nytimes.com": {Domain: "www.nytimes.com"},
		"www.cnn.com":     {Domain: "www.cnn.com"},
		"medium.com":      {Domain: "medium.com"},
		"cnn.com":         {Domain: "cnn.com"},
	}
}

func demoDetectByHTML(doc *goquery.Document) SimpleExtractor {
	// In real implementation, this would be detect-by-html.js equivalent
	
	// Example: Medium detection
	if doc.Find("meta[property='al:ios:app_name'][content='Medium']").Length() > 0 {
		return SimpleExtractor{Domain: "medium.com"}
	}
	
	// Example: Blogger detection
	if doc.Find("meta[content*='blogger']").Length() > 0 {
		return SimpleExtractor{Domain: "blogspot.com"}
	}
	
	return SimpleExtractor{} // Empty domain means no detection
}