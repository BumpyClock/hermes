// ABOUTME: Generic URL extractor with JavaScript-compatible canonical URL detection and domain extraction
// ABOUTME: Faithfully ports URL extraction logic from JavaScript with same priority order and behavior

package generic

import (
	"net"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/hermes/pkg/utils/dom"
)

// URL extraction constants matching JavaScript behavior exactly
var (
	// CANONICAL_META_SELECTORS - meta tag names for canonical URL extraction
	// From JavaScript: export const CANONICAL_META_SELECTORS = ['og:url'];
	CANONICAL_META_SELECTORS = []string{
		"og:url",
	}
)

// URLResult represents the extracted URL and domain information
type URLResult struct {
	URL    string `json:"url"`
	Domain string `json:"domain"`
}

// parseDomain extracts the domain from a URL string
// This is a faithful port of the JavaScript parseDomain function
func parseDomain(urlStr string) string {
	if urlStr == "" {
		return ""
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	// Return the hostname (domain) part
	// Note: url.Parse handles port numbers automatically by excluding them from Hostname()
	hostname := parsedURL.Hostname()
	if hostname == "" {
		// For some edge cases, try the Host field (includes port)
		if parsedURL.Host != "" {
			// Extract just the hostname part if there's a port
			if host, _, err := net.SplitHostPort(parsedURL.Host); err == nil {
				return host
			}
			return parsedURL.Host
		}
		return ""
	}

	return hostname
}

// result creates a URLResult with url and domain
// This mirrors the JavaScript result() helper function
func result(urlStr string) URLResult {
	return URLResult{
		URL:    urlStr,
		Domain: parseDomain(urlStr),
	}
}

// GenericUrlExtractor provides URL extraction functionality matching JavaScript exactly
var GenericUrlExtractor = struct {
	Extract func(doc *goquery.Selection, url string, metaCache []string) URLResult
}{
	Extract: func(doc *goquery.Selection, url string, metaCache []string) URLResult {
		// First, check for canonical link tag
		// JavaScript: const $canonical = $('link[rel=canonical]');
		canonical := doc.Find("link[rel=canonical]")
		if canonical.Length() != 0 {
			href, exists := canonical.Attr("href")
			if exists && href != "" {
				return result(href)
			}
		}

		// Second, check for canonical URL in meta tags
		// Need to convert selection to document for meta tag extraction
		var document *goquery.Document
		
		// Check if we already have a document
		if doc.Is("html") {
			// We might already have the document root
			if docNode := doc.Get(0); docNode != nil {
				document = goquery.NewDocumentFromNode(docNode)
			}
		}
		
		if document == nil {
			// Create document from HTML content
			if html, err := doc.Html(); err == nil {
				fullHTML := html
				if !containsHTML(html) {
					fullHTML = "<html>" + html + "</html>"
				}
				
				if tempDoc, err := goquery.NewDocumentFromReader(strings.NewReader(fullHTML)); err == nil {
					document = tempDoc
				}
			}
		}

		if document != nil {
			// JavaScript: const metaUrl = extractFromMeta($, CANONICAL_META_SELECTORS, metaCache);
			metaURL := dom.ExtractFromMeta(document, CANONICAL_META_SELECTORS, metaCache, false)
			if metaURL != nil && *metaURL != "" {
				return result(*metaURL)
			}
		}

		// Finally, return the original URL
		// JavaScript: return result(url);
		return result(url)
	},
}

// Helper function to check if string contains HTML tags
func containsHTML(s string) bool {
	return strings.Contains(s, "<html") || strings.Contains(s, "<!DOCTYPE")
}