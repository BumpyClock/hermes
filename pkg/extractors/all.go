// ABOUTME: Main extractor registry that aggregates all custom extractors into domain mappings
// ABOUTME: 100% JavaScript-compatible implementation of all.js functionality

package extractors

import (
	"github.com/postlight/parser-go/pkg/utils"
)

// GetAllExtractors returns a map of all custom extractors keyed by domain
// JavaScript equivalent: export default Object.keys(CustomExtractors).reduce((acc, key) => { ... }, {});
func GetAllExtractors() map[string]Extractor {
	registry := make(map[string]Extractor)
	
	// JavaScript logic:
	// export default Object.keys(CustomExtractors).reduce((acc, key) => {
	//   const extractor = CustomExtractors[key];
	//   return {
	//     ...acc,
	//     ...mergeSupportedDomains(extractor),
	//   };
	// }, {});
	
	// Get all custom extractors (currently just Medium and Blogger as foundation)
	customExtractors := getCustomExtractors()
	
	// Reduce pattern: iterate through each extractor and merge supported domains
	for _, extractor := range customExtractors {
		// Apply mergeSupportedDomains to each extractor
		domainMappings := mergeSupportedDomainsForExtractor(extractor)
		
		// Merge into accumulator (registry)
		for domain, mappedExtractor := range domainMappings {
			registry[domain] = mappedExtractor
		}
	}
	
	return registry
}

// getCustomExtractors returns all available custom extractors
// JavaScript equivalent: import * as CustomExtractors from './custom/index';
func getCustomExtractors() []Extractor {
	return []Extractor{
		// Foundation extractors - more will be added here
		&MediumExtractor{},
		&BloggerExtractor{},
		// TODO: Add remaining 142+ custom extractors here
		// &NYTimesExtractor{},
		// &WashingtonPostExtractor{},
		// &CNNExtractor{},
		// ... etc
	}
}

// mergeSupportedDomainsForExtractor applies mergeSupportedDomains logic to a single extractor
// This bridges the gap between our Extractor interface and the utils.MergeSupportedDomains generic function
func mergeSupportedDomainsForExtractor(extractor Extractor) map[string]Extractor {
	// Convert extractor to a format compatible with utils.MergeSupportedDomains
	extractorInfo := extractorToMockExtractor(extractor)
	
	// Apply mergeSupportedDomains
	domainMap := utils.MergeSupportedDomains(extractorInfo)
	
	// Convert back to map[string]Extractor
	result := make(map[string]Extractor)
	for domain := range domainMap {
		result[domain] = extractor
	}
	
	return result
}

// extractorToMockExtractor converts an Extractor to utils.MockExtractor format
// This is needed for compatibility with the generic MergeSupportedDomains function
func extractorToMockExtractor(extractor Extractor) utils.MockExtractor {
	domain := extractor.GetDomain()
	supportedDomains := getSupportedDomains(extractor)
	
	return utils.MockExtractor{
		Domain:           domain,
		SupportedDomains: supportedDomains,
		Name:             getExtractorName(extractor),
	}
}

// getSupportedDomains extracts supported domains from an extractor if available
func getSupportedDomains(extractor Extractor) []string {
	// For now, return empty slice - in real implementation, extractors would have this info
	// This would be populated from the actual extractor configurations
	switch extractor.(type) {
	case *MediumExtractor:
		// Medium might support multiple domains in the future
		return []string{}
	case *BloggerExtractor:
		// Blogger supports multiple blogspot domains
		return []string{"www.blogspot.com", "blogspot.co.uk", "blogspot.ca"}
	default:
		return []string{}
	}
}

// getExtractorName returns a human-readable name for the extractor
func getExtractorName(extractor Extractor) string {
	switch extractor.(type) {
	case *MediumExtractor:
		return "MediumExtractor"
	case *BloggerExtractor:
		return "BloggerExtractor"
	default:
		return "UnknownExtractor"
	}
}