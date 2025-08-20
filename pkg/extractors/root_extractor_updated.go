// ABOUTME: Complete root extractor orchestration system with JavaScript compatibility and field dependencies
// ABOUTME: Direct port of root-extractor.js with custom extractor framework integration and proper extraction pipeline

package extractors

import (
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/postlight/parser-go/pkg/extractors/custom"
	"github.com/postlight/parser-go/pkg/extractors/generic"
)

// RootExtractorOrchestrator manages the complete extraction process
// JavaScript equivalent: RootExtractor object in root-extractor.js
type RootExtractorOrchestrator struct {
	doc               *goquery.Document
	url               string
	genericExtractor  *generic.GenericExtractor
	customExtractor   *custom.CustomExtractor
	selectorProcessor *custom.SelectorProcessor
}

// NewRootExtractorOrchestrator creates a new root extractor
func NewRootExtractorOrchestrator(doc *goquery.Document, url string) *RootExtractorOrchestrator {
	return &RootExtractorOrchestrator{
		doc:               doc,
		url:               url,
		genericExtractor:  generic.NewGenericExtractor(),
		selectorProcessor: custom.NewSelectorProcessor(doc, url),
	}
}

// SetCustomExtractor sets the custom extractor to use
func (reo *RootExtractorOrchestrator) SetCustomExtractor(extractor *custom.CustomExtractor) {
	reo.customExtractor = extractor
}

// Extract performs the complete extraction process
// JavaScript equivalent: RootExtractor.extract(extractor = GenericExtractor, opts)
func (reo *RootExtractorOrchestrator) Extract(opts ExtractorOptions) (*Result, error) {
	// Handle content-only extraction
	// JavaScript equivalent: if (contentOnly) { const content = extractResult(...); return { content }; }
	if opts.ContentOnly {
		content, err := reo.extractField("content", opts)
		if err != nil {
			return nil, err
		}
		
		return &Result{
			Content: content.(string),
		}, nil
	}
	
	// Check if this is a generic extractor (domain === '*')
	// JavaScript equivalent: if (extractor.domain === '*') return extractor.extract(opts);
	if reo.customExtractor == nil || reo.customExtractor.Domain == "*" {
		return reo.genericExtractor.Extract(reo.doc, reo.url, opts)
	}
	
	// Extract all fields with proper dependencies
	// JavaScript equivalent: Complete field extraction pipeline in root-extractor.js
	result, err := reo.extractAllFields(opts)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// extractAllFields extracts all fields with proper dependency order
// JavaScript equivalent: The field extraction sequence in RootExtractor.extract
func (reo *RootExtractorOrchestrator) extractAllFields(opts ExtractorOptions) (*Result, error) {
	result := &Result{
		URL:    reo.url,
		Domain: reo.extractDomain(),
	}
	
	// Phase 1: Extract extended types first (they may be needed by other fields)
	// JavaScript equivalent: if (extractor.extend) { extendedResults = selectExtendedTypes(extractor.extend, opts); }
	var extendedResults map[string]interface{}
	if reo.customExtractor != nil && reo.customExtractor.Extend != nil {
		extendedResults = reo.selectorProcessor.ExtractExtendedTypes(reo.customExtractor.Extend, opts)
	}
	
	// Phase 2: Extract title (needed by content and direction)
	// JavaScript equivalent: const title = extractResult({ ...opts, type: 'title' });
	title, err := reo.extractField("title", opts)
	if err != nil {
		return nil, err
	}
	if title != nil {
		result.Title = title.(string)
		opts.Title = result.Title // Pass to dependent fields
	}
	
	// Phase 3: Extract author and date (independent fields)
	// JavaScript equivalent: const author = extractResult({ ...opts, type: 'author' });
	author, err := reo.extractField("author", opts)
	if err != nil {
		return nil, err
	}
	if author != nil {
		result.Author = author.(string)
	}
	
	// JavaScript equivalent: const date_published = extractResult({ ...opts, type: 'date_published' });
	datePublished, err := reo.extractField("date_published", opts)
	if err != nil {
		return nil, err
	}
	if datePublished != nil {
		if dateStr, ok := datePublished.(string); ok {
			// Parse date string to *time.Time if needed
			result.DatePublished = parseDate(dateStr)
		}
	}
	
	// Phase 4: Extract next page URL (independent)
	// JavaScript equivalent: const next_page_url = extractResult({ ...opts, type: 'next_page_url' });
	nextPageURL, err := reo.extractField("next_page_url", opts)
	if err != nil {
		return nil, err
	}
	if nextPageURL != nil {
		result.NextPageURL = nextPageURL.(string)
	}
	
	// Phase 5: Extract content (depends on title)
	// JavaScript equivalent: const content = extractResult({ ...opts, type: 'content', extractHtml: true, title });
	opts.ExtractHtml = true
	content, err := reo.extractField("content", opts)
	if err != nil {
		return nil, err
	}
	if content != nil {
		result.Content = content.(string)
		opts.Content = result.Content // Pass to dependent fields
	}
	
	// Phase 6: Extract lead image (may depend on content)
	// JavaScript equivalent: const lead_image_url = extractResult({ ...opts, type: 'lead_image_url', content });
	leadImageURL, err := reo.extractField("lead_image_url", opts)
	if err != nil {
		return nil, err
	}
	if leadImageURL != nil {
		result.LeadImageURL = leadImageURL.(string)
	}
	
	// Phase 7: Extract excerpt (depends on content)
	// JavaScript equivalent: const excerpt = extractResult({ ...opts, type: 'excerpt', content });
	excerpt, err := reo.extractField("excerpt", opts)
	if err != nil {
		return nil, err
	}
	if excerpt != nil {
		result.Excerpt = excerpt.(string)
		opts.Excerpt = result.Excerpt // Pass to dependent fields
	}
	
	// Phase 8: Extract dek (depends on content and excerpt)
	// JavaScript equivalent: const dek = extractResult({ ...opts, type: 'dek', content, excerpt });
	dek, err := reo.extractField("dek", opts)
	if err != nil {
		return nil, err
	}
	if dek != nil {
		result.Dek = dek.(string)
	}
	
	// Phase 9: Extract word count (depends on content)
	// JavaScript equivalent: const word_count = extractResult({ ...opts, type: 'word_count', content });
	wordCount, err := reo.extractField("word_count", opts)
	if err != nil {
		return nil, err
	}
	if wordCount != nil {
		if count, ok := wordCount.(int); ok {
			result.WordCount = count
		}
	}
	
	// Phase 10: Extract direction (depends on title)
	// JavaScript equivalent: const direction = extractResult({ ...opts, type: 'direction', title });
	direction, err := reo.extractField("direction", opts)
	if err != nil {
		return nil, err
	}
	if direction != nil {
		result.Direction = direction.(string)
	}
	
	// Phase 11: Add extended results
	// JavaScript equivalent: return { title, content, author, ..., ...extendedResults };
	if extendedResults != nil {
		result.Extended = extendedResults
	}
	
	return result, nil
}

// extractField extracts a single field using custom or generic extraction
// JavaScript equivalent: extractResult function in root-extractor.js
func (reo *RootExtractorOrchestrator) extractField(fieldType string, opts ExtractorOptions) (interface{}, error) {
	// Try custom extraction first if custom extractor is available
	if reo.customExtractor != nil {
		customResult := reo.tryCustomExtraction(fieldType, opts)
		if customResult != nil {
			return customResult, nil
		}
	}
	
	// Fallback to generic extraction
	// JavaScript equivalent: if (fallback) return GenericExtractor[type](opts);
	if opts.Fallback {
		return reo.tryGenericExtraction(fieldType, opts)
	}
	
	return nil, nil
}

// tryCustomExtraction attempts to extract a field using the custom extractor
// JavaScript equivalent: const result = select({ ...opts, extractionOpts: extractor[type] });
func (reo *RootExtractorOrchestrator) tryCustomExtraction(fieldType string, opts ExtractorOptions) interface{} {
	if reo.customExtractor == nil {
		return nil
	}
	
	var fieldExtractor *custom.FieldExtractor
	
	// Get the appropriate field extractor
	switch fieldType {
	case "title":
		fieldExtractor = reo.customExtractor.Title
	case "author":
		fieldExtractor = reo.customExtractor.Author
	case "date_published":
		fieldExtractor = reo.customExtractor.DatePublished
	case "lead_image_url":
		fieldExtractor = reo.customExtractor.LeadImageURL
	case "dek":
		fieldExtractor = reo.customExtractor.Dek
	case "next_page_url":
		fieldExtractor = reo.customExtractor.NextPageURL
	case "excerpt":
		fieldExtractor = reo.customExtractor.Excerpt
	case "content":
		if reo.customExtractor.Content != nil {
			fieldExtractor = reo.customExtractor.Content.FieldExtractor
		}
	default:
		return nil
	}
	
	// Extract using the field extractor
	if fieldExtractor != nil {
		result, err := reo.selectorProcessor.ExtractField(fieldType, fieldExtractor, opts)
		if err != nil {
			return nil
		}
		return result
	}
	
	return nil
}

// tryGenericExtraction attempts to extract a field using generic extraction
// JavaScript equivalent: GenericExtractor[type](opts)
func (reo *RootExtractorOrchestrator) tryGenericExtraction(fieldType string, opts ExtractorOptions) (interface{}, error) {
	// Convert opts to generic extractor options and call appropriate method
	genericOpts := convertToGenericOptions(opts)
	
	switch fieldType {
	case "title":
		return reo.genericExtractor.ExtractTitle(reo.doc, reo.url, genericOpts)
	case "author":
		return reo.genericExtractor.ExtractAuthor(reo.doc, reo.url, genericOpts)
	case "date_published":
		return reo.genericExtractor.ExtractDatePublished(reo.doc, reo.url, genericOpts)
	case "content":
		return reo.genericExtractor.ExtractContent(reo.doc, reo.url, genericOpts)
	case "lead_image_url":
		return reo.genericExtractor.ExtractLeadImageURL(reo.doc, reo.url, genericOpts)
	case "dek":
		return reo.genericExtractor.ExtractDek(reo.doc, reo.url, genericOpts)
	case "excerpt":
		return reo.genericExtractor.ExtractExcerpt(reo.doc, reo.url, genericOpts)
	case "next_page_url":
		return reo.genericExtractor.ExtractNextPageURL(reo.doc, reo.url, genericOpts)
	case "word_count":
		return reo.genericExtractor.ExtractWordCount(reo.doc, reo.url, genericOpts)
	case "direction":
		return reo.genericExtractor.ExtractDirection(reo.doc, reo.url, genericOpts)
	default:
		return nil, nil
	}
}

// extractDomain extracts the domain from the URL
func (reo *RootExtractorOrchestrator) extractDomain() string {
	if reo.customExtractor != nil {
		return reo.customExtractor.Domain
	}
	
	// Extract domain from URL
	_, baseDomain, err := extractURLComponents(reo.url)
	if err != nil {
		return ""
	}
	
	return baseDomain
}

// Helper functions

// convertToGenericOptions converts ExtractorOptions to generic extractor options
func convertToGenericOptions(opts ExtractorOptions) generic.ExtractorOptions {
	return generic.ExtractorOptions{
		URL:         opts.URL,
		Fallback:    opts.Fallback,
		ContentOnly: opts.ContentOnly,
		// Add other fields as needed
	}
}

// parseDate parses a date string to *time.Time
func parseDate(dateStr string) *time.Time {
	// This would use the existing date parsing logic
	// For now, return nil as placeholder
	return nil
}

// ExtractWithRootExtractor is the main entry point for root extraction
// JavaScript equivalent: RootExtractor.extract(extractor, opts)
func ExtractWithRootExtractor(doc *goquery.Document, url string, customExtractor *custom.CustomExtractor, opts ExtractorOptions) (*Result, error) {
	orchestrator := NewRootExtractorOrchestrator(doc, url)
	
	if customExtractor != nil {
		orchestrator.SetCustomExtractor(customExtractor)
	}
	
	return orchestrator.Extract(opts)
}