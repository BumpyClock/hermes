package resource

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// CreateDocument fetches HTML or prepares supplied, already-decoded HTML.
func CreateDocument(ctx context.Context, rawURL, html string, parsedURL *url.URL, headers map[string]string, client *HTTPClient) (*goquery.Document, error) {
	if html != "" {
		return PrepareDocument(ctx, []byte(html), "text/html", true)
	}
	response, err := Fetch(ctx, rawURL, parsedURL, headers, client)
	if err != nil {
		return nil, fmt.Errorf("resource fetch failed: %s", err)
	}
	return PrepareDocument(ctx, response.Body, response.GetContentType(), false)
}

// PrepareDocument decodes HTML, checks resource limits, and prepares the DOM.
func PrepareDocument(ctx context.Context, body []byte, contentType string, alreadyDecoded bool) (*goquery.Document, error) {
	// Check if content appears to be HTML/text
	if !IsTextContent(contentType) {
		return nil, fmt.Errorf("content does not appear to be text, got: %s", contentType)
	}

	// Validate resource limits before processing
	if err := ValidateResourceLimits(body); err != nil {
		return nil, fmt.Errorf("resource limits exceeded: %w", err)
	}

	// Handle encoding and create initial document
	doc, err := EncodeDoc(body, contentType, alreadyDecoded)
	if err != nil {
		return nil, fmt.Errorf("failed to encode document: %w", err)
	}

	// Check if document parsed correctly
	if doc.Find("*").Length() == 0 {
		return nil, fmt.Errorf("no children found, likely a bad parse")
	}

	// Validate DOM complexity
	if err := ValidateDOMComplexity(doc); err != nil {
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

// ValidateResourceLimits checks if the resource is within safe processing limits.
func ValidateResourceLimits(body []byte) error {
	bodySize := len(body)

	if bodySize > MAX_DOCUMENT_SIZE {
		return fmt.Errorf("document size %d bytes exceeds maximum %d bytes", bodySize, MAX_DOCUMENT_SIZE)
	}

	return nil
}

// ValidateDOMComplexity checks if the DOM has too many elements.
func ValidateDOMComplexity(doc *goquery.Document) error {
	elementCount := doc.Find("*").Length()

	if elementCount > MAX_DOM_ELEMENTS {
		return fmt.Errorf("DOM has %d elements, exceeds maximum %d", elementCount, MAX_DOM_ELEMENTS)
	}

	return nil
}

// EncodeDoc handles character encoding detection and document creation.
func EncodeDoc(content []byte, contentType string, alreadyDecoded bool) (*goquery.Document, error) {
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
		doc, err = recheckEncoding(content, doc, contentType)
		if err != nil {
			return nil, err
		}
	}

	return doc, nil
}

// recheckEncoding checks if encoding in header matches encoding in HTML meta tags
// and re-encodes if necessary (matches JavaScript behavior).
func recheckEncoding(content []byte, doc *goquery.Document, headerContentType string) (*goquery.Document, error) {
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
