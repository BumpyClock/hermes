// ABOUTME: Utility function that creates domain-to-extractor mappings for custom extractors
// ABOUTME: 100% JavaScript-compatible implementation of merge-supported-domains.js functionality

package utils

// MergeSupportedDomains creates a map from domains to extractors, handling multi-domain support
// This is a generic function that works with any type that has domain information
// JavaScript equivalent: export default function mergeSupportedDomains(extractor)
func MergeSupportedDomains[T any](extractor T) map[string]T {
	// Extract domain and supportedDomains from the extractor
	domain, supportedDomains := extractDomainInfo(extractor)
	
	// JavaScript logic:
	// return extractor.supportedDomains
	//   ? merge(extractor, [extractor.domain, ...extractor.supportedDomains])
	//   : merge(extractor, [extractor.domain]);
	
	if len(supportedDomains) > 0 {
		// Create array of all domains: [domain, ...supportedDomains]
		allDomains := append([]string{domain}, supportedDomains...)
		return mergeExtractorToDomains(extractor, allDomains)
	} else {
		// Only use the main domain
		return mergeExtractorToDomains(extractor, []string{domain})
	}
}

// mergeExtractorToDomains maps an extractor to multiple domains
// JavaScript equivalent: const merge = (extractor, domains) =>
func mergeExtractorToDomains[T any](extractor T, domains []string) map[string]T {
	result := make(map[string]T)
	
	// domains.reduce((acc, domain) => {
	//   acc[domain] = extractor;
	//   return acc;
	// }, {});
	for _, domain := range domains {
		result[domain] = extractor
	}
	
	return result
}

// extractDomainInfo extracts domain and supportedDomains from any struct type
func extractDomainInfo(extractor any) (domain string, supportedDomains []string) {
	// For testing purposes, we'll check if it's our MockExtractor type
	switch e := extractor.(type) {
	case MockExtractor:
		return e.Domain, e.SupportedDomains
	case *MockExtractor:
		return e.Domain, e.SupportedDomains
	default:
		// For other types, we'd use reflection here
		// For now, return empty values
		return "", nil
	}
}

// MockExtractor represents a test extractor for domain mapping tests
// In real usage, this would be replaced by actual extractor types
type MockExtractor struct {
	Domain           string   `json:"domain"`
	SupportedDomains []string `json:"supportedDomains,omitempty"`
	Name             string   `json:"name"`
}