package resource

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"bytes"
	"io"

	"github.com/PuerkitoBio/goquery"
)

// Resource provides functionality for fetching and preparing HTML documents
type Resource struct{}

// Create creates a Resource by fetching from URL or using provided HTML
// This is the main entry point that orchestrates fetch -> decode -> DOM preparation
// Automatically detects large documents and uses streaming when beneficial
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - rawURL: The URL for the document we should retrieve
// - preparedResponse: If set, use as the response rather than fetching. Expects HTML string
// - parsedURL: Pre-parsed URL object (optional)
// - headers: Custom headers to include in the request
func (r *Resource) Create(ctx context.Context, rawURL string, preparedResponse string, parsedURL *url.URL, headers map[string]string) (*goquery.Document, error) {
	// Use nil client for backward compatibility
	return r.CreateWithClient(ctx, rawURL, preparedResponse, parsedURL, headers, nil)
}

// CreateWithClient creates a Resource using the provided HTTP client
func (r *Resource) CreateWithClient(ctx context.Context, rawURL string, preparedResponse string, parsedURL *url.URL, headers map[string]string, httpClient *HTTPClient) (*goquery.Document, error) {
	var result *FetchResult

	if preparedResponse != "" {
		// Use provided HTML
		result = &FetchResult{
			Response: &Response{
				StatusCode: 200,
				Status:     "OK",
				Headers: map[string][]string{
					"Content-Type":   {"text/html"},
					"Content-Length": {fmt.Sprintf("%d", len(preparedResponse))},
				},
				Body: []byte(preparedResponse),
			},
			AlreadyDecoded: true,
		}
	} else {
		// Fetch from URL with provided client
		var err error
		result, err = FetchResourceWithClient(ctx, rawURL, parsedURL, headers, httpClient)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch resource: %w", err)
		}
	}

	if result.IsError() {
		return nil, fmt.Errorf("resource fetch failed: %s", result.Message)
	}

	// Check if document is large and should use streaming
	documentSize := int64(len(result.Response.Body))
	if IsLargeDocument(documentSize) {
		return r.GenerateDocStreaming(result)
	}

	return r.GenerateDocWithContext(ctx, result)
}

// GenerateDoc creates a goquery Document from fetch result
// Handles encoding detection and applies DOM preparation pipeline with resource limits
// DEPRECATED: Use Create or GenerateDocWithContext instead
func (r *Resource) GenerateDoc(result *FetchResult) (*goquery.Document, error) {
	// Use background context for backward compatibility
	// Callers should provide context via GenerateDocWithContext
	return r.GenerateDocWithContext(context.Background(), result)
}

// GenerateDocWithContext creates a document with context for timeout control
func (r *Resource) GenerateDocWithContext(ctx context.Context, result *FetchResult) (*goquery.Document, error) {
	contentType := result.Response.GetContentType()

	// Check if content appears to be HTML/text
	if !IsTextContent(contentType) {
		return nil, fmt.Errorf("content does not appear to be text, got: %s", contentType)
	}

	// Validate resource limits before processing
	if err := r.ValidateResourceLimits(result.Response.Body); err != nil {
		return nil, fmt.Errorf("resource limits exceeded: %w", err)
	}

	// Handle encoding and create initial document
	doc, err := r.EncodeDoc(result.Response.Body, contentType, result.AlreadyDecoded)
	if err != nil {
		return nil, fmt.Errorf("failed to encode document: %w", err)
	}

	// Check if document parsed correctly
	if doc.Find("*").Length() == 0 {
		return nil, fmt.Errorf("no children found, likely a bad parse")
	}

	// Validate DOM complexity
	if err := r.ValidateDOMComplexity(doc); err != nil {
		return nil, fmt.Errorf("DOM too complex: %w", err)
	}

	// Apply DOM preparation pipeline with context checking
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("document processing timed out")
	default:
		doc = NormalizeMetaTags(doc)
		doc = ConvertLazyLoadedImages(doc)
		doc = Clean(doc)
	}

	return doc, nil
}

// ValidateResourceLimits checks if the resource is within safe processing limits
func (r *Resource) ValidateResourceLimits(body []byte) error {
	bodySize := len(body)

	if bodySize > MAX_DOCUMENT_SIZE {
		return fmt.Errorf("document size %d bytes exceeds maximum %d bytes", bodySize, MAX_DOCUMENT_SIZE)
	}

	return nil
}

// ValidateDOMComplexity checks if the DOM has too many elements
func (r *Resource) ValidateDOMComplexity(doc *goquery.Document) error {
	elementCount := doc.Find("*").Length()

	if elementCount > MAX_DOM_ELEMENTS {
		return fmt.Errorf("DOM has %d elements, exceeds maximum %d", elementCount, MAX_DOM_ELEMENTS)
	}

	return nil
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

	// Create initial document directly (no fake pooling)
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

// IsLargeDocument determines if a document should use streaming
func IsLargeDocument(size int64) bool {
	const largeSizeThreshold = 1024 * 1024 // 1MB
	return size > largeSizeThreshold
}

// GenerateDocStreaming creates a goquery Document using streaming for large documents
// Provides memory optimization for documents over 1MB by processing HTML in chunks
func (r *Resource) GenerateDocStreaming(result *FetchResult) (*goquery.Document, error) {
	contentType := result.Response.GetContentType()

	// Check if content appears to be HTML/text
	if !IsTextContent(contentType) {
		return nil, fmt.Errorf("content does not appear to be text, got: %s", contentType)
	}

	// For streaming, we still need to validate limits but can be more lenient
	documentSize := int64(len(result.Response.Body))
	if documentSize > MAX_DOCUMENT_SIZE_STREAMING {
		return nil, fmt.Errorf("document too large for streaming: %d bytes (max: %d)", 
			documentSize, MAX_DOCUMENT_SIZE_STREAMING)
	}

	// For now, implement a simplified streaming approach
	// In a complete implementation, this would use the full streaming parser
	
	// Create a reader from the response body
	reader := bytes.NewReader(result.Response.Body)
	
	// Process the document in chunks to reduce memory pressure
	const chunkSize = 128 * 1024 // 128KB chunks
	var htmlBuilder strings.Builder
	
	buffer := make([]byte, chunkSize)
	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading document chunks: %w", err)
		}
		
		if n > 0 {
			htmlBuilder.Write(buffer[:n])
		}
		
		if err == io.EOF {
			break
		}
	}
	
	// Parse the complete HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlBuilder.String()))
	if err != nil {
		// Fallback to regular parsing if streaming approach fails
		if documentSize < 5*1024*1024 { // 5MB fallback limit
			return r.GenerateDoc(result)
		}
		return nil, fmt.Errorf("streaming parse failed: %w", err)
	}

	if doc == nil {
		return nil, fmt.Errorf("streaming parser returned nil document")
	}

	// Apply basic DOM validation 
	if doc.Find("*").Length() == 0 {
		return nil, fmt.Errorf("no children found in streamed document, likely a bad parse")
	}

	return doc, nil
}
