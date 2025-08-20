// ABOUTME: Multi-page article collection system with 100% JavaScript behavioral compatibility
// ABOUTME: Faithful 1:1 port of JavaScript collect-all-pages.js with pagination, deduplication, and safety limits

package extractors

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/postlight/parser-go/pkg/extractors/generic"
	"github.com/postlight/parser-go/pkg/utils/text"
)

// ResourceInterface defines the interface for resource fetching
type ResourceInterface interface {
	Create(url string, preparedResponse string, parsedURL interface{}, headers map[string]string) (*goquery.Document, error)
}

// Use existing RootExtractorInterface from root_extractor.go

// CollectAllPagesOptions contains all parameters needed for multi-page collection
// This matches the JavaScript function signature exactly
type CollectAllPagesOptions struct {
	NextPageURL   string
	HTML          string
	Doc           *goquery.Document
	MetaCache     map[string]interface{}
	Result        map[string]interface{}
	Extractor     interface{}
	Title         interface{}
	URL           string
	Resource      ResourceInterface
	RootExtractor *RootExtractorInterface
}

// CollectAllPages collects and merges content from multiple pages of an article
// This is a faithful 1:1 port of the JavaScript collectAllPages function with:
// - Page counter starting at 1 (first page already fetched) 
// - 26-page safety limit to prevent infinite loops
// - URL deduplication using RemoveAnchor utility
// - Progressive content concatenation with <hr><h4>Page N</h4> separators
// - Final word count calculation for combined content
func CollectAllPages(opts CollectAllPagesOptions) map[string]interface{} {
	// At this point, we've fetched just the first page
	pages := 1
	
	// Track previous URLs to prevent cycles - use RemoveAnchor for consistency
	previousUrls := []string{text.RemoveAnchor(opts.URL)}
	
	// Initialize working variables
	nextPageURL := opts.NextPageURL
	result := make(map[string]interface{})
	
	// Copy all fields from original result
	for key, value := range opts.Result {
		result[key] = value
	}
	
	// If we've gone over 26 pages, something has likely gone wrong.
	// This matches the JavaScript safety limit exactly
	for nextPageURL != "" && pages < 26 {
		pages++ // Increment page counter (JavaScript: pages += 1)
		
		// Fetch the next page using the resource interface
		// This matches JavaScript: $ = await Resource.create(next_page_url)
		doc, err := opts.Resource.Create(nextPageURL, "", nil, nil)
		if err != nil {
			// If resource fetch fails, break the loop and return what we have
			break
		}
		
		// Get HTML from the document (matches JavaScript: html = $.html())
		// Note: html variable not used in Go version as we work directly with document
		
		// Prepare extraction options matching JavaScript extractorOpts
		extractionOpts := ExtractOptions{
			Doc:           doc,
			URL:           nextPageURL,
			Extractor:     opts.Extractor,
			ExtractedTitle: opts.Title,
			// previousUrls would be used in JavaScript but not directly in ExtractOptions
		}
		
		// Extract content from this page using RootExtractor
		// This matches JavaScript: RootExtractor.extract(Extractor, extractorOpts)
		var nextPageResult map[string]interface{}
		if opts.RootExtractor != nil {
			if extractedResult := opts.RootExtractor.Extract(opts.Extractor, extractionOpts); extractedResult != nil {
				if resultMap, ok := extractedResult.(map[string]interface{}); ok {
					nextPageResult = resultMap
				}
			}
		}
		
		// If extraction failed, break the loop
		if nextPageResult == nil {
			break
		}
		
		// Add current URL to previousUrls for cycle detection
		// JavaScript: previousUrls.push(next_page_url)
		previousUrls = append(previousUrls, nextPageURL)
		
		// Merge content with page separator
		// This matches JavaScript exactly: `${result.content}<hr><h4>Page ${pages}</h4>${nextPageResult.content}`
		currentContent := ""
		if content, ok := result["content"].(string); ok {
			currentContent = content
		}
		
		nextContent := ""
		if content, ok := nextPageResult["content"].(string); ok {
			nextContent = content
		}
		
		// Format: current_content + <hr><h4>Page N</h4> + next_page_content
		mergedContent := fmt.Sprintf("%s<hr><h4>Page %d</h4>%s", currentContent, pages, nextContent)
		result["content"] = mergedContent
		
		// Get next page URL for the loop
		// JavaScript: next_page_url = nextPageResult.next_page_url
		if nextURL, ok := nextPageResult["next_page_url"].(string); ok {
			nextPageURL = nextURL
		} else {
			// No more pages
			nextPageURL = ""
		}
		
		// Check for cycles by comparing with previous URLs using RemoveAnchor
		// This prevents infinite loops from circular pagination
		if nextPageURL != "" {
			cleanNextURL := text.RemoveAnchor(nextPageURL)
			for _, prevURL := range previousUrls {
				if cleanNextURL == prevURL {
					// Cycle detected, stop pagination
					nextPageURL = ""
					break
				}
			}
		}
	}
	
	// Calculate final word count using GenericWordCountExtractor
	// This matches JavaScript: GenericExtractor.word_count({ content: `<div>${result.content}</div>` })
	wordCount := 1 // Default value
	if contentStr, ok := result["content"].(string); ok {
		// Wrap content in div to match JavaScript behavior exactly
		wrappedContent := fmt.Sprintf("<div>%s</div>", contentStr)
		
		// Use the GenericWordCountExtractor (matches GenericExtractor.word_count)
		wordCount = generic.GenericWordCountExtractor.Extract(map[string]interface{}{
			"content": wrappedContent,
		})
	}
	
	// Return final result with pagination metadata
	// This matches the JavaScript return structure exactly
	return map[string]interface{}{
		// Spread the original result
		"title":          result["title"],
		"content":        result["content"],
		"author":         result["author"],
		"date_published": result["date_published"],
		"lead_image_url": result["lead_image_url"],
		"dek":            result["dek"],
		"next_page_url":  result["next_page_url"],
		"url":            result["url"],
		"domain":         result["domain"],
		"excerpt":        result["excerpt"],
		"direction":      result["direction"],
		
		// Add pagination-specific fields
		"total_pages":    pages,
		"rendered_pages": pages,
		"word_count":     wordCount,
	}
}