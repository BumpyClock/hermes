// ABOUTME: Generic extractor orchestration matching JavaScript GenericExtractor functionality
// ABOUTME: Coordinates all individual field extractors with proper dependencies and JavaScript-compatible behavior

package generic

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	
	"github.com/PuerkitoBio/goquery"
)

// GenericExtractor orchestrates all individual field extractors
// Direct port of JavaScript src/extractors/generic/index.js
type GenericExtractor struct {
	Domain string
}

// NewGenericExtractor creates a new generic extractor instance
func NewGenericExtractor() *GenericExtractor {
	return &GenericExtractor{
		Domain: "*", // Matches JavaScript GenericExtractor domain
	}
}

// GetDomain returns the domain this extractor handles
func (ge *GenericExtractor) GetDomain() string {
	return ge.Domain
}

// ExtractionResult represents the complete result from generic extraction
// Matches JavaScript extraction result structure exactly
type ExtractionResult struct {
	Title         string     `json:"title"`
	Author        string     `json:"author"`
	DatePublished *time.Time `json:"date_published"` // null if not found
	Dek           string     `json:"dek"`
	LeadImageURL  string     `json:"lead_image_url"`
	Content       string     `json:"content"`
	NextPageURL   string     `json:"next_page_url"`
	URL           string     `json:"url"`
	Domain        string     `json:"domain"`
	Excerpt       string     `json:"excerpt"`
	WordCount     int        `json:"word_count"`
	Direction     string     `json:"direction"`
}

// ExtractionOptions contains all parameters needed for extraction
// Matches JavaScript options object structure
type ExtractionOptions struct {
	URL         string
	HTML        string
	Doc         *goquery.Document
	MetaCache   []string
	Fallback    bool
	ContentType string
}

// This method has been moved to ExtractGeneric to avoid conflicts with the Extractor interface
// See ExtractGeneric method below for the full implementation

// Individual extractor methods - delegate to specialized extractors

func (ge *GenericExtractor) extractTitle(options *ExtractionOptions) string {
	// Use the existing GenericTitleExtractor with Selection interface
	if options.Doc == nil {
		return ""
	}
	selection := options.Doc.Selection
	return GenericTitleExtractor.Extract(selection, options.URL, options.MetaCache)
}

func (ge *GenericExtractor) extractAuthor(options *ExtractionOptions) string {
	// Use the existing GenericAuthorExtractor
	if options.Doc == nil {
		return ""
	}
	extractor := &GenericAuthorExtractor{}
	result := extractor.Extract(options.Doc.Selection, options.MetaCache)
	if result != nil {
		return *result
	}
	return ""
}

func (ge *GenericExtractor) extractDatePublished(options *ExtractionOptions) *time.Time {
	// Use the existing GenericDateExtractor
	if options.Doc == nil {
		return nil
	}
	result := GenericDateExtractor.Extract(options.Doc.Selection, options.URL, options.MetaCache)
	if result != nil && *result != "" {
		// Parse the date string to time.Time
		if parsedTime, err := time.Parse("2006-01-02T15:04:05Z07:00", *result); err == nil {
			return &parsedTime
		}
	}
	return nil
}

func (ge *GenericExtractor) extractContent(options *ExtractionOptions, title string) string {
	// Use the existing GenericContentExtractor
	extractor := NewGenericContentExtractor()
	
	extractorOptions := ExtractorOptions{
		StripUnlikelyCandidates: options.Fallback,
		WeightNodes:            options.Fallback,
		CleanConditionally:     options.Fallback,
	}
	
	params := ExtractorParams{
		Doc:     options.Doc,
		URL:     options.URL,
		Title:   title,
	}
	
	return extractor.Extract(params, extractorOptions)
}

func (ge *GenericExtractor) extractLeadImageURL(options *ExtractionOptions, content string) string {
	// Use the existing image extractor
	if options.Doc == nil {
		return ""
	}
	
	extractor := NewGenericLeadImageExtractor()
	params := ExtractorImageParams{
		Doc:     options.Doc,
		Content: content,
	}
	
	result := extractor.Extract(params)
	if result != nil {
		return *result
	}
	return ""
}

func (ge *GenericExtractor) extractDek(options *ExtractionOptions, content string) string {
	// Use the existing dek extractor
	if options.Doc == nil {
		return ""
	}
	
	extractor := &GenericDekExtractor{}
	opts := map[string]interface{}{
		"excerpt": content, // Use content as excerpt for comparison
	}
	
	return extractor.Extract(options.Doc, opts)
}

func (ge *GenericExtractor) extractNextPageURL(options *ExtractionOptions) string {
	// Use the existing next page URL extractor
	if options.Doc == nil {
		return ""
	}
	
	extractor := NewGenericNextPageUrlExtractor()
	parsedURL, _ := url.Parse(options.URL)
	previousUrls := []string{} // Empty for generic extraction
	
	return extractor.Extract(options.Doc, options.URL, parsedURL, previousUrls)
}

func (ge *GenericExtractor) extractExcerpt(options *ExtractionOptions, content string) string {
	// Use the existing excerpt extractor
	if options.Doc == nil {
		return ""
	}
	
	extractor := NewGenericExcerptExtractor()
	return extractor.Extract(options.Doc, content, options.MetaCache)
}

func (ge *GenericExtractor) extractWordCount(options *ExtractionOptions, content string) int {
	// Use the existing word count extractor
	opts := map[string]interface{}{
		"content": content,
	}
	
	return GenericWordCountExtractor.Extract(opts)
}

func (ge *GenericExtractor) extractDirection(title string) string {
	// Use the existing direction extractor
	params := ExtractorParams{
		Title: title,
	}
	
	direction, _ := DirectionExtractor(params)
	return direction
}

func (ge *GenericExtractor) extractURLAndDomain(options *ExtractionOptions) (string, string) {
	// Simple URL and domain extraction
	if options.URL == "" {
		return "", ""
	}
	
	// Parse the URL to get domain
	parsedURL, err := url.Parse(options.URL)
	if err != nil {
		return options.URL, ""
	}
	
	domain := parsedURL.Hostname()
	return options.URL, domain
}

// ExtractGeneric performs the main generic extraction with full options
func (ge *GenericExtractor) ExtractGeneric(options *ExtractionOptions) (*ExtractionResult, error) {
	// Ensure we have a document to work with
	if options.Doc == nil && options.HTML != "" {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(options.HTML))
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML: %w", err)
		}
		options.Doc = doc
	}
	
	if options.Doc == nil {
		return nil, fmt.Errorf("no document or HTML provided for extraction")
	}

	// Extract fields in JavaScript-compatible order with proper dependencies
	// Phase 1: Independent extractions
	title := ge.extractTitle(options)
	datePublished := ge.extractDatePublished(options)
	author := ge.extractAuthor(options)
	nextPageURL := ge.extractNextPageURL(options)
	
	// Phase 2: Content extraction (depends on title)
	content := ge.extractContent(options, title)
	
	// Phase 3: Content-dependent extractions
	leadImageURL := ge.extractLeadImageURL(options, content)
	dek := ge.extractDek(options, content)
	excerpt := ge.extractExcerpt(options, content)
	wordCount := ge.extractWordCount(options, content)
	
	// Phase 4: Final extractions
	direction := ge.extractDirection(title)
	url, domain := ge.extractURLAndDomain(options)

	return &ExtractionResult{
		Title:         title,
		Author:        author,
		DatePublished: datePublished,
		Dek:           dek,
		LeadImageURL:  leadImageURL,
		Content:       content,
		NextPageURL:   nextPageURL,
		URL:           url,
		Domain:        domain,
		Excerpt:       excerpt,
		WordCount:     wordCount,
		Direction:     direction,
	}, nil
}

// Extract implements the Extractor interface for compatibility with extractor selection
func (ge *GenericExtractor) Extract(doc *goquery.Document) (interface{}, error) {
	options := &ExtractionOptions{
		Doc:         doc,
		URL:         "",
		HTML:        "",
		MetaCache:   []string{},
		Fallback:    true,
		ContentType: "html",
	}
	return ge.ExtractGeneric(options)
}