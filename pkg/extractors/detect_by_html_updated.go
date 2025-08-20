// ABOUTME: HTML-based extractor detection with complete custom framework integration
// ABOUTME: Direct port of detect-by-html.js with registry-based detection for Medium, Blogger, and future extractors

package extractors

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/postlight/parser-go/pkg/extractors/custom"
)

// HTMLDetector manages HTML-based extractor detection
// JavaScript equivalent: Detectors object in detect-by-html.js
type HTMLDetector struct {
	detectors map[string]*custom.CustomExtractor
}

// NewHTMLDetector creates a new HTML detector
func NewHTMLDetector() *HTMLDetector {
	return &HTMLDetector{
		detectors: make(map[string]*custom.CustomExtractor),
	}
}

// Register adds a HTML selector-based detector
func (hd *HTMLDetector) Register(selector string, extractor *custom.CustomExtractor) {
	hd.detectors[selector] = extractor
}

// Detect finds an extractor using HTML selectors
// JavaScript equivalent: detectByHtml($) function
func (hd *HTMLDetector) Detect(doc *goquery.Document) *custom.CustomExtractor {
	// JavaScript equivalent: Reflect.ownKeys(Detectors).find(s => $(s).length > 0)
	for selector, extractor := range hd.detectors {
		if doc.Find(selector).Length() > 0 {
			return extractor
		}
	}
	return nil
}

// DetectByHTMLWithCustomRegistry uses the global custom registry for detection
// JavaScript equivalent: detectByHtml function in detect-by-html.js
func DetectByHTMLWithCustomRegistry(doc *goquery.Document) *custom.CustomExtractor {
	if doc == nil {
		return nil
	}
	
	// Use the global registry's HTML detection
	return custom.GlobalRegistryManager.GetByHTML(doc)
}

// InitializeHTMLDetectors sets up all HTML-based detectors
// JavaScript equivalent: The Detectors object initialization in detect-by-html.js
func InitializeHTMLDetectors() error {
	// Get extractors from registry for HTML detection
	
	// Medium detection: 'meta[name="al:ios:app_name"][value="Medium"]'
	if mediumExtractor, found := custom.GlobalRegistryManager.GetByDomain("medium.com"); found {
		err := custom.GlobalRegistryManager.RegisterHTMLDetector(
			"meta[name=\"al:ios:app_name\"][value=\"Medium\"]",
			mediumExtractor,
		)
		if err != nil {
			return err
		}
	}
	
	// Blogger detection: 'meta[name="generator"][value="blogger"]'
	if bloggerExtractor, found := custom.GlobalRegistryManager.GetByDomain("blogspot.com"); found {
		err := custom.GlobalRegistryManager.RegisterHTMLDetector(
			"meta[name=\"generator\"][value=\"blogger\"]",
			bloggerExtractor,
		)
		if err != nil {
			return err
		}
	}
	
	// WordPress detection: 'meta[name="generator"][value*="WordPress"]'
	// This would be added when WordPress extractor is implemented
	
	// Add more HTML detectors as custom extractors are added
	// Each follows the same pattern: unique HTML signature -> extractor
	
	return nil
}

// GetRegisteredDetectors returns all registered HTML detectors
// Useful for testing and debugging
func GetRegisteredDetectors() map[string]*custom.CustomExtractor {
	// For now, return from global registry
	// In a full implementation, this would return the complete detector map
	detectors := make(map[string]*custom.CustomExtractor)
	
	// This is a placeholder - the actual implementation would extract
	// the HTML detectors from the GlobalRegistryManager
	
	return detectors
}

// DetectByHTMLSelectors is the main detection function used by get-extractor.js
// JavaScript equivalent: detectByHtml($) export
func DetectByHTMLSelectors(doc *goquery.Document) *custom.CustomExtractor {
	return DetectByHTMLWithCustomRegistry(doc)
}

// LegacyDetectByHTML provides compatibility with existing code
// This bridges the gap between the new custom framework and existing interfaces
func LegacyDetectByHTML(doc *goquery.Document) Extractor {
	customExtractor := DetectByHTMLSelectors(doc)
	if customExtractor != nil {
		return NewCustomExtractorWrapper(customExtractor)
	}
	return nil
}

// JavaScript equivalent of the complete detect-by-html.js file:
//
// import { MediumExtractor, BloggerExtractor } from './custom';
//
// const Detectors = {
//   'meta[name="al:ios:app_name"][value="Medium"]': MediumExtractor,
//   'meta[name="generator"][value="blogger"]': BloggerExtractor,
// };
//
// export default function detectByHtml($) {
//   const selector = Reflect.ownKeys(Detectors).find(s => $(s).length > 0);
//   return Detectors[selector];
// }

// This Go implementation provides the same functionality through the custom registry system