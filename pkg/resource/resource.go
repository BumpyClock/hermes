package resource

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Resource provides functionality for fetching and preparing HTML documents
type Resource struct{}

// Create creates a Resource by fetching from URL or using provided HTML
// This is the main entry point that orchestrates fetch -> decode -> DOM preparation
//
// Parameters:
// - rawURL: The URL for the document we should retrieve
// - preparedResponse: If set, use as the response rather than fetching. Expects HTML string
// - parsedURL: Pre-parsed URL object (optional)
// - headers: Custom headers to include in the request
func (r *Resource) Create(rawURL string, preparedResponse string, parsedURL *url.URL, headers map[string]string) (*goquery.Document, error) {
	var result *FetchResult
	
	if preparedResponse != "" {
		// Use provided HTML
		result = &FetchResult{
			Body: []byte(preparedResponse),
			Response: &Response{
				StatusCode: 200,
				Status:     "OK",
				Headers:    map[string][]string{
					"Content-Type": {"text/html"},
					"Content-Length": {fmt.Sprintf("%d", len(preparedResponse))},
				},
			},
			AlreadyDecoded: true,
		}
	} else {
		// Fetch from URL
		var err error
		result, err = FetchResource(rawURL, parsedURL, headers)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch resource: %w", err)
		}
	}
	
	if result.IsError() {
		return nil, fmt.Errorf("resource fetch failed: %s", result.Message)
	}
	
	return r.GenerateDoc(result)
}

// GenerateDoc creates a goquery Document from fetch result
// Handles encoding detection and applies DOM preparation pipeline
func (r *Resource) GenerateDoc(result *FetchResult) (*goquery.Document, error) {
	contentType := result.Response.GetContentType()
	
	// Check if content appears to be HTML/text
	if !IsTextContent(contentType) {
		return nil, fmt.Errorf("content does not appear to be text, got: %s", contentType)
	}
	
	// Handle encoding and create initial document
	doc, err := r.EncodeDoc(result.Body, contentType, result.AlreadyDecoded)
	if err != nil {
		return nil, fmt.Errorf("failed to encode document: %w", err)
	}
	
	// Check if document parsed correctly
	if doc.Find("*").Length() == 0 {
		return nil, fmt.Errorf("no children found, likely a bad parse")
	}
	
	// Apply DOM preparation pipeline
	doc = NormalizeMetaTags(doc)
	doc = ConvertLazyLoadedImages(doc)
	doc = Clean(doc)
	
	return doc, nil
}

// EncodeDoc handles character encoding detection and document creation
func (r *Resource) EncodeDoc(content []byte, contentType string, alreadyDecoded bool) (*goquery.Document, error) {
	var htmlContent string
	var err error
	
	if alreadyDecoded {
		htmlContent = string(content)
	} else {
		// Detect and convert encoding
		htmlContent, err = DetectAndDecodeText(content, contentType)
		if err != nil {
			return nil, fmt.Errorf("encoding detection failed: %w", err)
		}
	}
	
	// Create initial document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	
	// After first parse, check for encoding mismatch in meta tags
	if !alreadyDecoded {
		doc, err = r.recheckEncoding(content, doc, contentType)
		if err != nil {
			return nil, err
		}
	}
	
	return doc, nil
}

// recheckEncoding checks if encoding in header matches encoding in HTML meta tags
// and re-encodes if necessary (matches JavaScript behavior)
func (r *Resource) recheckEncoding(content []byte, doc *goquery.Document, headerContentType string) (*goquery.Document, error) {
	// Get encoding from Content-Type header
	headerEncoding := getEncodingFromContentType(headerContentType)
	
	// Check for meta charset in document
	var metaContentType string
	
	// Look for <meta http-equiv="content-type" content="...">
	doc.Find("meta[http-equiv]").Each(func(i int, s *goquery.Selection) {
		httpEquiv, _ := s.Attr("http-equiv")
		if strings.ToLower(httpEquiv) == "content-type" {
			if content, exists := s.Attr("value"); exists { // We normalized content -> value
				metaContentType = content
			}
		}
	})
	
	// Also check for <meta charset="...">
	if metaContentType == "" {
		if charset, exists := doc.Find("meta[charset]").Attr("charset"); exists {
			metaContentType = "charset=" + charset
		}
	}
	
	// If we found meta charset, check if it differs from header
	if metaContentType != "" {
		metaEncoding := getEncodingFromContentType(metaContentType)
		
		// If encodings differ, re-decode with the correct one
		if metaEncoding != nil && headerEncoding != nil && 
		   metaEncoding != headerEncoding {
			
			htmlContent, err := DetectAndDecodeText(content, metaContentType)
			if err != nil {
				return doc, nil // Return original doc if re-encoding fails
			}
			
			// Re-parse with correct encoding
			newDoc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
			if err != nil {
				return doc, nil // Return original doc if re-parsing fails
			}
			
			return newDoc, nil
		}
	}
	
	return doc, nil
}

// NewResource creates a new Resource instance
func NewResource() *Resource {
	return &Resource{}
}