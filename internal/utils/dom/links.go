package dom

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// MakeLinksAbsolute converts all relative URLs in the document to absolute URLs
// This exactly matches the JavaScript makeLinksAbsolute implementation
// JavaScript: export default function makeLinksAbsolute($content, $, url)
func MakeLinksAbsolute(doc *goquery.Document, rootURL string) *goquery.Document {
	// Check for base tag first (JavaScript behavior)
	// JavaScript: const baseUrl = $('base').attr('href');
	baseURL := rootURL
	baseTag := doc.Find("base").First()
	if baseTag.Length() > 0 {
		if baseHref, exists := baseTag.Attr("href"); exists && baseHref != "" {
			baseURL = baseHref
		}
	}
	
	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		return doc
	}

	// JavaScript: ['href', 'src'].forEach(attr => absolutize($, url, attr));
	absolutize(doc, parsedBase, "href")
	absolutize(doc, parsedBase, "src")
	
	// JavaScript: absolutizeSet($, url, $content);
	absolutizeSet(doc, parsedBase)

	return doc
}

// absolutize processes a specific attribute across all elements
// JavaScript: function absolutize($, rootUrl, attr)
func absolutize(doc *goquery.Document, baseURL *url.URL, attr string) {
	// JavaScript: $(`[${attr}]`).each((_, node) => {
	doc.Find("[" + attr + "]").Each(func(index int, element *goquery.Selection) {
		attrs := GetAttrs(element)
		urlValue, exists := attrs[attr]
		if !exists || urlValue == "" {
			return
		}
		
		// JavaScript: const absoluteUrl = URL.resolve(baseUrl || rootUrl, url);
		absoluteURL := makeAbsoluteURL(urlValue, baseURL)
		if absoluteURL != "" {
			element.SetAttr(attr, absoluteURL)
		}
	})
}

// absolutizeSet processes srcset attributes for responsive images
// JavaScript: function absolutizeSet($, rootUrl, $content)
func absolutizeSet(doc *goquery.Document, baseURL *url.URL) {
	// JavaScript: $('[srcset]', $content).each((_, node) => {
	doc.Find("[srcset]").Each(func(index int, element *goquery.Selection) {
		attrs := GetAttrs(element)
		urlSet, exists := attrs["srcset"]
		if !exists || urlSet == "" {
			return
		}
		
		// JavaScript regex: /(?:\s*)(\S+(?:\s*[\d.]+[wx])?)(?:\s*,\s*)?/g
		// a comma should be considered part of the candidate URL unless preceded by a descriptor
		// descriptors can only contain positive numbers followed immediately by either 'w' or 'x'
		candidateRegex := regexp.MustCompile(`(?:\s*)(\S+(?:\s*[\d.]+[wx])?)(?:\s*,\s*)?`)
		candidates := candidateRegex.FindAllString(urlSet, -1)
		
		if len(candidates) == 0 {
			return
		}
		
		// JavaScript: const absoluteCandidates = candidates.map(candidate => {
		var absoluteCandidates []string
		for _, candidate := range candidates {
			// a candidate URL cannot start or end with a comma
			// descriptors are separated from the URLs by unescaped whitespace
			trimmed := strings.TrimSpace(candidate)
			trimmed = strings.TrimSuffix(trimmed, ",")
			
			// JavaScript: .split(/\s+/)
			parts := strings.Fields(trimmed)
			if len(parts) > 0 {
				// JavaScript: parts[0] = URL.resolve(rootUrl, parts[0]);
				parts[0] = makeAbsoluteURL(parts[0], baseURL)
				// JavaScript: return parts.join(' ');
				absoluteCandidates = append(absoluteCandidates, strings.Join(parts, " "))
			}
		}
		
		// JavaScript: const absoluteUrlSet = [...new Set(absoluteCandidates)].join(', ');
		// Remove duplicates and join
		unique := make(map[string]bool)
		var finalCandidates []string
		for _, candidate := range absoluteCandidates {
			if !unique[candidate] {
				unique[candidate] = true
				finalCandidates = append(finalCandidates, candidate)
			}
		}
		
		absoluteURLSet := strings.Join(finalCandidates, ", ")
		element.SetAttr("srcset", absoluteURLSet)
	})
}

// makeAbsoluteURL converts a potentially relative URL to absolute using the base URL
func makeAbsoluteURL(href string, base *url.URL) string {
	// Skip if already absolute
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	// Skip javascript: and mailto: links
	if strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "mailto:") {
		return href
	}

	// Handle protocol-relative URLs
	if strings.HasPrefix(href, "//") {
		var urlBuilder strings.Builder
		urlBuilder.WriteString(base.Scheme)
		urlBuilder.WriteString(":")
		urlBuilder.WriteString(href)
		return urlBuilder.String()
	}

	// Parse the relative URL
	relativeURL, err := url.Parse(href)
	if err != nil {
		return ""
	}

	// Resolve against base URL
	absoluteURL := base.ResolveReference(relativeURL)
	return absoluteURL.String()
}

// ArticleBaseURL extracts the base URL for the article, removing fragments and query parameters
func ArticleBaseURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Remove fragment and query
	parsedURL.Fragment = ""
	parsedURL.RawQuery = ""

	return parsedURL.String()
}

// RemoveAnchor removes the anchor/fragment from a URL
func RemoveAnchor(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	parsedURL.Fragment = ""
	return parsedURL.String()
}

// ValidateURL checks if a URL is valid and well-formed
func ValidateURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Must have a scheme and host
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	// Must be http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	return true
}

// GetDomain extracts the domain from a URL
func GetDomain(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Remove www. prefix if present
	host := parsedURL.Host
	if strings.HasPrefix(host, "www.") {
		host = host[4:]
	}

	return host
}

// GetBaseDomain extracts the base domain (removing subdomains) from a URL
func GetBaseDomain(rawURL string) string {
	domain := GetDomain(rawURL)
	if domain == "" {
		return ""
	}

	// Simple logic: if there are more than 2 parts, take the last 2
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}

	return domain
}

// SanitizeURL cleans up a URL by removing tracking parameters and normalizing
func SanitizeURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Remove common tracking parameters
	trackingParams := []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "gclid", "ref", "source", "campaign",
	}

	query := parsedURL.Query()
	for _, param := range trackingParams {
		query.Del(param)
	}

	parsedURL.RawQuery = query.Encode()
	return parsedURL.String()
}