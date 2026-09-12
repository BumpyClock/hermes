package dom

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var srcsetCandidateRE = regexp.MustCompile(`(?:\s*)(\S+(?:\s*[\d.]+[wx])?)(?:\s*,\s*)?`)

// MakeLinksAbsolute converts all relative URLs in the document to absolute URLs
// This exactly matches the JavaScript makeLinksAbsolute implementation
// JavaScript: export default function makeLinksAbsolute($content, $, url).
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
// JavaScript: function absolutize($, rootUrl, attr).
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
// JavaScript: function absolutizeSet($, rootUrl, $content).
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
		candidates := srcsetCandidateRE.FindAllString(urlSet, -1)

		if len(candidates) == 0 {
			return
		}

		// JavaScript: const absoluteCandidates = candidates.map(candidate => {
		unique := make(map[string]bool)
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
				absoluteCandidate := strings.Join(parts, " ")
				if !unique[absoluteCandidate] {
					unique[absoluteCandidate] = true
					absoluteCandidates = append(absoluteCandidates, absoluteCandidate)
				}
			}
		}

		// JavaScript: const absoluteUrlSet = [...new Set(absoluteCandidates)].join(', ');
		absoluteURLSet := strings.Join(absoluteCandidates, ", ")
		element.SetAttr("srcset", absoluteURLSet)
	})
}

// makeAbsoluteURL converts a potentially relative URL to absolute using the base URL.
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
