// ABOUTME: Pointer optimization utilities to reduce struct copying in performance-critical paths.
// This file provides optimized versions of commonly used functions that use pointers instead of value copying.
package parser

import (
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/parser-go/pkg/resource"
)

// PointerOptimizedMercury provides optimized methods using pointers for large structs
type PointerOptimizedMercury struct {
	*Mercury
}

// NewPointerOptimized creates a new pointer-optimized Mercury parser
func NewPointerOptimized(opts ...*ParserOptions) *PointerOptimizedMercury {
	var defaultOpts ParserOptions
	if len(opts) > 0 && opts[0] != nil {
		defaultOpts = *opts[0]
	} else {
		defaultOpts = ParserOptions{
			FetchAllPages: true,
			Fallback:      true,
			ContentType:   "html",
		}
	}

	return &PointerOptimizedMercury{
		Mercury: &Mercury{options: defaultOpts},
	}
}

// ParseOptimized extracts content from a URL using pointer-optimized parameters
func (pom *PointerOptimizedMercury) ParseOptimized(targetURL string, opts *ParserOptions) (*Result, error) {
	// Parse and validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, &ParseError{
			URL:   targetURL,
			Phase: "url_parsing",
			Err:   err,
		}
	}

	if !validateURL(parsedURL) {
		return &Result{
			Error:   true,
			Message: "The url parameter passed does not look like a valid URL. Please check your URL and try again.",
		}, nil
	}

	// Use default options if nil provided
	var finalOpts ParserOptions
	if opts != nil {
		finalOpts = *opts
	} else {
		finalOpts = pom.Mercury.options
	}

	// Create resource from URL (with fetching)
	r := createResourceOptimized()
	doc, err := r.CreateOptimized(targetURL, "", parsedURL, finalOpts.Headers)
	if err != nil {
		return nil, &ParseError{
			URL:   targetURL,
			Phase: "resource_creation",
			Err:   err,
		}
	}

	// Extract all fields using optimized method
	result, err := pom.extractAllFieldsOptimized(doc, targetURL, parsedURL, &finalOpts)
	if err != nil {
		return nil, &ParseError{
			URL:   targetURL,
			Phase: "extraction",
			Err:   err,
		}
	}

	return result, nil
}

// ParseHTMLOptimized extracts content from provided HTML using pointer optimization
func (pom *PointerOptimizedMercury) ParseHTMLOptimized(html string, targetURL string, opts *ParserOptions) (*Result, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, &ParseError{
			URL:   targetURL,
			Phase: "url_parsing",
			Err:   err,
		}
	}

	if !validateURL(parsedURL) {
		return nil, &ParseError{
			URL:   targetURL,
			Phase: "url_validation",
			Err:   err,
		}
	}

	// Use default options if nil provided
	var finalOpts ParserOptions
	if opts != nil {
		finalOpts = *opts
	} else {
		finalOpts = pom.Mercury.options
	}

	// Create resource from provided HTML
	r := createResourceOptimized()
	doc, err := r.CreateOptimized(targetURL, html, parsedURL, finalOpts.Headers)
	if err != nil {
		return nil, &ParseError{
			URL:   targetURL,
			Phase: "resource_creation",
			Err:   err,
		}
	}

	// Extract all fields using optimized method
	result, err := pom.extractAllFieldsOptimized(doc, targetURL, parsedURL, &finalOpts)
	if err != nil {
		return nil, &ParseError{
			URL:   targetURL,
			Phase: "extraction",
			Err:   err,
		}
	}

	return result, nil
}

// extractAllFieldsOptimized performs extraction using pointer-optimized parameters
func (pom *PointerOptimizedMercury) extractAllFieldsOptimized(doc *goquery.Document, targetURL string, parsedURL *url.URL, opts *ParserOptions) (*Result, error) {
	// Convert to legacy format for now - in a full optimization, we'd update all extractors
	optsValue := *opts
	result, err := pom.Mercury.extractAllFields(doc, targetURL, parsedURL, optsValue)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// OptimizedExtractorOptions provides pointer-optimized version of ExtractorOptions
type OptimizedExtractorOptions struct {
	URL         *string
	HTML        *string
	MetaCache   *map[string]string
	Fallback    *bool
	ContentType *string
}

// ToValue converts pointer-optimized options to value type for compatibility
func (oeo *OptimizedExtractorOptions) ToValue() ExtractorOptions {
	opts := ExtractorOptions{}

	if oeo.URL != nil {
		opts.URL = *oeo.URL
	}
	if oeo.HTML != nil {
		opts.HTML = *oeo.HTML
	}
	if oeo.MetaCache != nil {
		opts.MetaCache = *oeo.MetaCache
	}
	if oeo.Fallback != nil {
		opts.Fallback = *oeo.Fallback
	}
	if oeo.ContentType != nil {
		opts.ContentType = *oeo.ContentType
	}

	return opts
}

// OptimizedResource provides pointer-optimized resource operations
type OptimizedResource struct{}

// createResourceOptimized creates a new optimized resource instance
func createResourceOptimized() *OptimizedResource {
	return &OptimizedResource{}
}

// CreateOptimized creates a Resource by fetching from URL or using provided HTML with pointer optimization
func (or *OptimizedResource) CreateOptimized(rawURL string, preparedResponse string, parsedURL *url.URL, headers map[string]string) (*goquery.Document, error) {
	// For now, delegate to the existing implementation
	// In a full optimization, this would be rewritten to use pointers throughout
	r := &resource.Resource{}
	return r.Create(rawURL, preparedResponse, parsedURL, headers)
}

// BatchOptionsOptimized provides optimized batch processing with pointer parameters
type BatchOptionsOptimized struct {
	Options     *ParserOptions
	URLs        *[]string
	Concurrency *int
	Timeout     *int // seconds
}

// ProcessBatchOptimized processes multiple URLs with optimized parameter passing
func (pom *PointerOptimizedMercury) ProcessBatchOptimized(batchOpts *BatchOptionsOptimized) ([]*Result, error) {
	if batchOpts == nil || batchOpts.URLs == nil || len(*batchOpts.URLs) == 0 {
		return nil, &ParseError{
			URL:   "",
			Phase: "batch_validation",
			Err:   fmt.Errorf("URLs list cannot be empty"),
		}
	}

	urls := *batchOpts.URLs
	results := make([]*Result, len(urls))
	errors := make([]error, 0)

	// Use default options if not provided
	var opts *ParserOptions
	if batchOpts.Options != nil {
		opts = batchOpts.Options
	} else {
		defaultOpts := pom.Mercury.options
		opts = &defaultOpts
	}

	// Process URLs sequentially for now - in full optimization, use worker pool
	for i, url := range urls {
		result, err := pom.ParseOptimized(url, opts)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		results[i] = result
	}

	if len(errors) > 0 {
		// Return first error for simplicity
		return results, errors[0]
	}

	return results, nil
}

// ResultOptimizer provides utilities for optimizing Result struct operations
type ResultOptimizer struct{}

// MergeResultsOptimized efficiently merges multiple results using pointers
func (ro *ResultOptimizer) MergeResultsOptimized(primary *Result, additional ...*Result) *Result {
	if primary == nil {
		if len(additional) > 0 && additional[0] != nil {
			return additional[0]
		}
		return &Result{}
	}

	// Create a copy to avoid modifying the original
	merged := *primary

	// Merge content from additional results
	for _, result := range additional {
		if result == nil {
			continue
		}

		// Merge non-empty fields
		if merged.Title == "" && result.Title != "" {
			merged.Title = result.Title
		}
		if merged.Author == "" && result.Author != "" {
			merged.Author = result.Author
		}
		if merged.Content == "" && result.Content != "" {
			merged.Content = result.Content
		}
		if merged.DatePublished == nil && result.DatePublished != nil {
			merged.DatePublished = result.DatePublished
		}
		if merged.LeadImageURL == "" && result.LeadImageURL != "" {
			merged.LeadImageURL = result.LeadImageURL
		}
		if merged.Dek == "" && result.Dek != "" {
			merged.Dek = result.Dek
		}
		if merged.Excerpt == "" && result.Excerpt != "" {
			merged.Excerpt = result.Excerpt
		}
		
		// Merge numeric and other fields
		if merged.WordCount == 0 && result.WordCount != 0 {
			merged.WordCount = result.WordCount
		}
		if merged.Direction == "" && result.Direction != "" {
			merged.Direction = result.Direction
		}
		if merged.Domain == "" && result.Domain != "" {
			merged.Domain = result.Domain
		}
		if merged.NextPageURL == "" && result.NextPageURL != "" {
			merged.NextPageURL = result.NextPageURL
		}
		if merged.TotalPages == 0 && result.TotalPages != 0 {
			merged.TotalPages = result.TotalPages
		}
		if merged.RenderedPages == 0 && result.RenderedPages != 0 {
			merged.RenderedPages = result.RenderedPages
		}

		// Merge extended fields
		if result.Extended != nil {
			if merged.Extended == nil {
				merged.Extended = make(map[string]interface{})
			}
			for key, value := range result.Extended {
				if _, exists := merged.Extended[key]; !exists {
					merged.Extended[key] = value
				}
			}
		}
	}

	return &merged
}

// CloneResultOptimized creates a deep copy of a Result struct efficiently
func (ro *ResultOptimizer) CloneResultOptimized(result *Result) *Result {
	if result == nil {
		return nil
	}

	cloned := *result

	// Deep copy DatePublished if it exists
	if result.DatePublished != nil {
		dateClone := *result.DatePublished
		cloned.DatePublished = &dateClone
	}

	// Deep copy Extended map if it exists
	if result.Extended != nil {
		cloned.Extended = make(map[string]interface{}, len(result.Extended))
		for key, value := range result.Extended {
			cloned.Extended[key] = value
		}
	}

	return &cloned
}

// Global optimizer instances
var (
	GlobalResultOptimizer = &ResultOptimizer{}
)

// Convenience functions for easy access to optimized operations
func NewOptimizedParser(opts ...*ParserOptions) *PointerOptimizedMercury {
	return NewPointerOptimized(opts...)
}

func MergeResults(primary *Result, additional ...*Result) *Result {
	return GlobalResultOptimizer.MergeResultsOptimized(primary, additional...)
}

func CloneResult(result *Result) *Result {
	return GlobalResultOptimizer.CloneResultOptimized(result)
}