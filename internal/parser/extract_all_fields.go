// ABOUTME: Core extraction orchestration that coordinates all field extractors with proper signatures and error handling
// ABOUTME: Handles the complete extraction pipeline from DOM to structured Result using all available extractors

package parser

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"

	"github.com/BumpyClock/hermes/internal/cleaners"
	"github.com/BumpyClock/hermes/internal/extractors/custom"
	"github.com/BumpyClock/hermes/internal/extractors/generic"
	"github.com/BumpyClock/hermes/internal/utils/dom"
	"github.com/BumpyClock/hermes/internal/utils/security"
	"github.com/BumpyClock/hermes/internal/utils/text"
)

// parserDebugEnabled controls whether debug logging is enabled
// Set via HERMES_PARSER_DEBUG=1 environment variable.
var parserDebugEnabled = os.Getenv("HERMES_PARSER_DEBUG") == "1"

// extractAllFieldsWithContext orchestrates the complete extraction pipeline with context support.
func (h *Hermes) extractAllFieldsWithContext(ctx context.Context, doc *goquery.Document, targetURL string, parsedURL *url.URL, opts ParserOptions) (*Result, error) {
	// Check context before starting
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("extraction cancelled: %w", ctx.Err())
	default:
	}
	// An explicit content type preserves the caller's fallback preference.
	if opts.ContentType == "" && !opts.Fallback && opts.Headers == nil {
		opts.Fallback = true
	}
	if opts.ContentType == "" {
		opts.ContentType = "html"
	}

	// Create base result
	result := &Result{
		URL:    targetURL,
		Domain: parsedURL.Host,
	}

	// Build meta cache first for use by both custom and generic extractors
	metaCache := buildMetaCache(doc)

	result.SiteName = (&generic.GenericSiteNameExtractor{}).Extract(doc.Selection, targetURL, metaCache)
	result.SiteTitle = (&generic.GenericSiteTitleExtractor{}).Extract(doc.Selection, targetURL, metaCache)
	result.SiteImage = (&generic.GenericSiteImageExtractor{}).Extract(doc.Selection, targetURL, metaCache)
	result.Favicon = (&generic.GenericFaviconExtractor{}).Extract(doc.Selection, targetURL, metaCache)
	result.Description = (&generic.GenericDescriptionExtractor{}).Extract(doc.Selection, targetURL, metaCache)
	result.Language = (&generic.GenericLanguageExtractor{}).Extract(doc.Selection, targetURL, metaCache)
	result.ThemeColor = (&generic.GenericThemeColorExtractor{}).Extract(doc.Selection, targetURL, metaCache)

	// Check context after metadata extraction
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("extraction cancelled after metadata: %w", ctx.Err())
	default:
	}

	// Try to use custom extractor, passing the result with site metadata
	if customResult := h.tryCustomExtractor(doc, targetURL, parsedURL, opts, result, metaCache); customResult != nil {
		return customResult, nil
	}

	if title := generic.GenericTitleExtractor.Extract(doc.Selection, targetURL, metaCache); title != "" {
		result.Title = cleaners.ResolveSplitTitle(cleaners.CleanTitle(title, targetURL, doc), targetURL)
	}
	extractGenericAuthorAndDate(doc, targetURL, metaCache, result)
	result.Dek = (&generic.GenericDekExtractor{}).Extract(doc, map[string]interface{}{"$": doc.Selection})

	// Preserve the cancellation checkpoint before content extraction.
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("extraction cancelled after parallel extraction: %w", ctx.Err())
	default:
	}

	// Extract the lead image before content extraction.
	imageExtractor := generic.NewGenericLeadImageExtractor()
	imageParams := generic.ExtractorImageParams{
		Doc:       doc,
		Content:   "", // Will be set after content extraction
		MetaCache: make(map[string]string),
		HTML:      "", // Could enhance with original HTML
	}
	if imageURL := imageExtractor.Extract(imageParams); imageURL != nil && *imageURL != "" {
		// Use the new cleaner that properly validates URLs
		if cleaned := cleaners.CleanLeadImageURLValidated(*imageURL); cleaned != nil {
			result.LeadImageURL = *cleaned
		}
	}

	if content := extractGenericContent(doc, result.Title, targetURL); content != "" {
		setFormattedContent(result, content, opts.ContentType)

		// Update image extraction with content context
		imageParams.Content = result.Content
		if imageURL := imageExtractor.Extract(imageParams); imageURL != nil && *imageURL != "" && result.LeadImageURL == "" {
			result.LeadImageURL = cleaners.CleanLeadImageURL(*imageURL, targetURL)
		}

		// Update dek with excerpt context
		dekExtractor := &generic.GenericDekExtractor{}
		dekOpts := map[string]interface{}{
			"$":       doc.Selection,
			"excerpt": result.Excerpt,
		}
		if dek := dekExtractor.Extract(doc, dekOpts); dek != "" && result.Dek == "" {
			result.Dek = dek
		}
	}

	extractVideoMetadata(doc, targetURL, metaCache, result)

	// Set default values for fields not extracted
	if result.Title == "" && opts.Fallback {
		// Fallback title extraction
		if title := doc.Find("title").First().Text(); title != "" {
			result.Title = cleaners.CleanTitle(strings.TrimSpace(title), targetURL, doc)
		} else if h1 := doc.Find("h1").First().Text(); h1 != "" {
			result.Title = strings.TrimSpace(h1)
		}
	}

	// Basic validation - content should not be empty for successful extraction
	if result.Content == "" && opts.Fallback {
		// Try progressively broader fallback selectors
		fallbackSelectors := []string{
			"article, .article, #article, .content, #content, .entry-content",
			"main",
			"[role=main]",
			"body",
		}

		for _, selector := range fallbackSelectors {
			if basicContent := doc.Find(selector).First().Text(); basicContent != "" {
				result.Content = strings.TrimSpace(basicContent)
				result.Excerpt = text.ExcerptContent(result.Content, 160)
				result.WordCount = calculateWordCount(result.Content)
				break
			}
		}
	}

	return result, nil
}

// tryCustomExtractor attempts to use a custom extractor for the given domain.
func (h *Hermes) tryCustomExtractor(doc *goquery.Document, targetURL string, parsedURL *url.URL, opts ParserOptions, baseResult *Result, metaCache []string) *Result {
	// Look for custom extractor for this domain using the proper lookup function
	customExtractor, found := custom.GetCustomExtractorByDomain(parsedURL.Host)

	if !found {
		// Try fallback - remove 'www.' prefix if present
		if strings.HasPrefix(parsedURL.Host, "www.") {
			baseDomain := strings.TrimPrefix(parsedURL.Host, "www.")
			customExtractor, found = custom.GetCustomExtractorByDomain(baseDomain)
		} else {
			// Try adding 'www.' prefix
			wwwDomain := "www." + parsedURL.Host
			customExtractor, found = custom.GetCustomExtractorByDomain(wwwDomain)
		}
	}

	if !found || customExtractor == nil {
		// No custom extractor found
		return nil // No custom extractor found
	}

	// Create result with custom extractor info, preserving site metadata from base result
	result := &Result{
		URL:           targetURL,
		Domain:        parsedURL.Host,
		ExtractorUsed: "custom:" + customExtractor.Domain,
		// Preserve site metadata
		SiteName:    baseResult.SiteName,
		SiteTitle:   baseResult.SiteTitle,
		SiteImage:   baseResult.SiteImage,
		Favicon:     baseResult.Favicon,
		Description: baseResult.Description,
		Language:    baseResult.Language,
		ThemeColor:  baseResult.ThemeColor,
	}

	if title := firstCustomField(doc, customExtractor.Title); title != "" {
		result.Title = cleaners.CleanTitle(title, targetURL, doc)
	}
	if author := firstCustomField(doc, customExtractor.Author); author != "" {
		result.Author = cleaners.CleanAuthor(author)
	}

	// Extract content using custom selectors
	if customExtractor.Content != nil && len(customExtractor.Content.Selectors) > 0 {
		for _, selector := range customExtractor.Content.Selectors {
			contentElements := contentElementsForSelector(doc, selector)

			// Process the first selector with non-empty raw content, preserving fallback order.
			if contentElements == nil || !hasCustomContent(contentElements) {
				continue
			}

			contentHTML, err := processCustomContent(contentElements, doc, customExtractor.Content, result.Title, targetURL)
			if err == nil && strings.TrimSpace(contentHTML) != "" {
				setFormattedContent(result, contentHTML, opts.ContentType)
			}
			break
		}
	}

	if customExtractor.DatePublished != nil {
		for _, selector := range customExtractor.DatePublished.Selectors {
			if date, err := parseDate(customFieldValue(doc, selector)); err == nil {
				result.DatePublished = &date
				break
			}
		}
	}
	if imageURL := firstCustomField(doc, customExtractor.LeadImageURL); imageURL != "" {
		result.LeadImageURL = cleaners.CleanLeadImageURL(imageURL, targetURL)
	}

	// Fall back to generic extractors for missing fields if fallback is enabled
	if opts.Fallback {
		fallbackMetaCache := buildMetaCache(doc)

		// Fallback title extraction
		if result.Title == "" {
			if title := generic.GenericTitleExtractor.Extract(doc.Selection, targetURL, fallbackMetaCache); title != "" {
				result.Title = cleaners.CleanTitle(title, targetURL, doc)
			}
		}

		extractGenericAuthorAndDate(doc, targetURL, fallbackMetaCache, result)

		// Fallback content extraction if no content was found
		if result.Content == "" {
			if content := extractGenericContent(doc, result.Title, targetURL); content != "" {
				setFormattedContent(result, content, opts.ContentType)
			}
		}
	}

	extractVideoMetadata(doc, targetURL, metaCache, result)

	return result
}

func extractGenericContent(doc *goquery.Document, title, targetURL string) string {
	return generic.NewGenericContentExtractor().Extract(generic.ExtractorParams{
		Doc: doc, Title: title, URL: targetURL,
	}, generic.ExtractorOptions{
		StripUnlikelyCandidates: true,
		WeightNodes:             true,
		CleanConditionally:      true,
	})
}

func setFormattedContent(result *Result, content, contentType string) {
	result.Content = formatContent(content, contentType)
	if result.Content != "" {
		result.Excerpt = text.ExcerptContent(result.Content, 160)
	}
	result.WordCount = calculateWordCount(result.Content)
}

func extractVideoMetadata(doc *goquery.Document, targetURL string, metaCache []string, result *Result) {
	// Video metadata extraction
	videoExtractor := &generic.GenericVideoExtractor{}
	if videoData := videoExtractor.Extract(doc.Selection, targetURL, metaCache); videoData != nil {
		// Set the primary video URL
		if videoData.SecureURL != "" {
			result.VideoURL = videoData.SecureURL
		} else if videoData.URL != "" {
			result.VideoURL = videoData.URL
		}

		// Store full video metadata if we have any data
		if metadata := buildVideoMetadata(videoData); metadata != nil {
			result.VideoMetadata = metadata
		}
	}

}

func customFieldValue(doc *goquery.Document, selector custom.SelectorEntry) string {
	element := doc.Find(selector.Selector).First()
	if selector.Attribute != "" {
		return strings.TrimSpace(element.AttrOr(selector.Attribute, ""))
	}
	return strings.TrimSpace(element.Text())
}

func firstCustomField(doc *goquery.Document, extractor *custom.FieldExtractor) string {
	if extractor != nil {
		for _, selector := range extractor.Selectors {
			if value := customFieldValue(doc, selector); value != "" {
				return value
			}
		}
	}
	return ""
}

func extractGenericAuthorAndDate(doc *goquery.Document, targetURL string, metaCache []string, result *Result) {
	if result.Author == "" {
		if author := (&generic.GenericAuthorExtractor{}).Extract(doc.Selection, metaCache); author != nil && *author != "" {
			result.Author = cleaners.CleanAuthor(*author)
		}
	}
	if result.DatePublished == nil {
		if dateStr := generic.GenericDateExtractor.Extract(doc.Selection, targetURL, metaCache); dateStr != nil && *dateStr != "" {
			if date, err := parseDate(*dateStr); err == nil {
				result.DatePublished = &date
			}
		}
	}
}

func contentElementsForSelector(doc *goquery.Document, selectors custom.ContentSelectorGroup) *goquery.Selection {
	if len(selectors) == 0 {
		return nil
	}
	return doc.Find(strings.Join(selectors, ","))
}

func hasCustomContent(contentElements *goquery.Selection) bool {
	for _, element := range contentElements.Nodes {
		for child := element.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.TextNode && strings.TrimSpace(child.Data) == "" {
				continue
			}
			return true
		}
	}
	return false
}

func processCustomContent(contentElements *goquery.Selection, doc *goquery.Document, extractor *custom.ContentExtractor, title, targetURL string) (string, error) {
	var combinedContent strings.Builder
	var processErr error

	contentElements = outermostContentElements(contentElements)
	contentElements.EachWithBreak(func(_ int, element *goquery.Selection) bool {
		contentDoc, err := goquery.NewDocumentFromReader(strings.NewReader("<div></div>"))
		if err != nil {
			processErr = fmt.Errorf("create custom content wrapper: %w", err)
			return false
		}
		wrapper := contentDoc.Find("div").First()
		wrapper.AppendSelection(element.Clone())

		type transformMatch struct {
			selector  string
			transform custom.TransformFunction
			match     *goquery.Selection
			depth     int
			index     int
		}

		selectors := make([]string, 0, len(extractor.Transforms))
		for selector := range extractor.Transforms {
			selectors = append(selectors, selector)
		}
		sort.Strings(selectors)

		matches := make([]transformMatch, 0)
		seen := make(map[*html.Node]struct{})
		for _, selector := range selectors {
			transform := extractor.Transforms[selector]
			wrapper.Find(selector).Each(func(index int, match *goquery.Selection) {
				node := match.Get(0)
				if _, ok := seen[node]; ok {
					return
				}
				seen[node] = struct{}{}

				depth := 0
				for ancestor := node; ancestor != nil; ancestor = ancestor.Parent {
					depth++
				}
				matches = append(matches, transformMatch{
					selector:  selector,
					transform: transform,
					match:     match,
					depth:     depth,
					index:     index,
				})
			})
		}
		sort.Slice(matches, func(i, j int) bool {
			if matches[i].depth != matches[j].depth {
				return matches[i].depth > matches[j].depth
			}
			if matches[i].selector != matches[j].selector {
				return matches[i].selector < matches[j].selector
			}
			return matches[i].index < matches[j].index
		})

		for _, match := range matches {
			if transformErr := match.transform.Transform(match.match); transformErr != nil {
				processErr = fmt.Errorf("transform custom content selector %q: %w", match.selector, transformErr)
				return false
			}
		}

		for _, selector := range extractor.Clean {
			wrapper.Find(selector).Remove()
		}

		content := wrapper.Children().First()
		if !extractor.DisableDefaultCleaner {
			content = cleaners.ExtractCleanNode(content, doc, cleaners.ContentCleanOptions{
				CleanConditionally: true,
				Title:              title,
				URL:                targetURL,
			})
		}

		html, err := content.Html()
		if err != nil {
			processErr = fmt.Errorf("serialize custom content: %w", err)
			return false
		}
		if strings.TrimSpace(html) != "" {
			combinedContent.WriteString(html)
			combinedContent.WriteByte('\n')
		}
		return true
	})

	if processErr != nil {
		return "", processErr
	}
	return strings.TrimSpace(combinedContent.String()), nil
}

func outermostContentElements(contentElements *goquery.Selection) *goquery.Selection {
	selected := make(map[*html.Node]struct{}, contentElements.Length())
	for _, node := range contentElements.Nodes {
		selected[node] = struct{}{}
	}

	return contentElements.FilterFunction(func(_ int, element *goquery.Selection) bool {
		for ancestor := element.Get(0).Parent; ancestor != nil; ancestor = ancestor.Parent {
			if _, ok := selected[ancestor]; ok {
				return false
			}
		}
		return true
	})
}

// parseDate parses a date string into a time.Time.
func parseDate(dateStr string) (time.Time, error) {
	// Try common date formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"January 2, 2006",
		"Jan 2, 2006",
		"2006/01/02",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// stripHTMLTags removes HTML tags from content for text output
// Delegates to dom.StripTags for consistent behavior.
func stripHTMLTags(content string) string {
	return dom.StripTags(content)
}

// convertToMarkdown converts HTML content to Markdown using html-to-markdown library.
func convertToMarkdown(content string) string {
	// Create converter with options similar to TurndownService
	converter := md.NewConverter("", true, nil)

	// Configure options to match TurndownService behavior
	converter.Use(md.Plugin(func(c *md.Converter) []md.Rule {
		return []md.Rule{
			// Handle images properly with template URL resolution
			{
				Filter: []string{"img"},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					alt := selec.AttrOr("alt", "")
					src := selec.AttrOr("src", "")
					if src == "" {
						return md.String("")
					}

					// Resolve template placeholders in image URLs
					src = resolveImageTemplateURL(src, selec)

					result := fmt.Sprintf("![%s](%s)", alt, src)
					return &result
				},
			},
			// Handle links properly
			{
				Filter: []string{"a"},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					href := selec.AttrOr("href", "")
					if href == "" {
						return md.String(content)
					}
					result := fmt.Sprintf("[%s](%s)", content, href)
					return &result
				},
			},
		}
	}))

	// Convert HTML to Markdown
	markdown, err := converter.ConvertString(content)
	if err != nil {
		// Fallback to text extraction if conversion fails
		return stripHTMLTags(content)
	}

	return markdown
}

// formatContent applies the specified content type transformation and security sanitization
// Trims inputs, supports common content type aliases, and returns sanitized output.
func formatContent(content string, contentType string) string {
	// Trim both inputs
	content = strings.TrimSpace(content)
	contentType = strings.TrimSpace(contentType)

	// Normalize content type and handle common aliases
	normalized := strings.ToLower(contentType)

	// Map aliases to canonical types
	switch normalized {
	case "text", "text/plain", "txt":
		return text.NormalizeSpaces(stripHTMLTags(content))
	case "markdown", "md", "text/markdown":
		return convertToMarkdown(content)
	case "html", "text/html", "":
		// Empty string defaults to HTML (expected behavior)
		return security.SanitizeHTML(content)
	default:
		// Log unexpected content type only when debug is enabled
		if parserDebugEnabled {
			log.Printf("WARNING: Unexpected contentType '%s', defaulting to HTML sanitization", contentType)
		}
		return security.SanitizeHTML(content)
	}
}

// resolveImageTemplateURL resolves template placeholders in responsive image URLs.
func resolveImageTemplateURL(src string, imgElement *goquery.Selection) string {
	// Check if URL contains template placeholders
	if !strings.Contains(src, "{width}") && !strings.Contains(src, "{quality}") && !strings.Contains(src, "{format}") {
		return src // No templates, return as-is
	}

	// Look for reasonable default values to replace templates
	// These are common web standards that should work for most images
	defaultWidth := "1200"  // Reasonable default width
	defaultQuality := "85"  // Good balance of quality vs size
	defaultFormat := "jpeg" // Most compatible format

	// Try to get better values from the element's attributes
	if width, exists := imgElement.Attr("width"); exists && width != "" {
		defaultWidth = width
	}

	// Check for srcset or other attributes that might give us hints
	if srcset, exists := imgElement.Attr("srcset"); exists && srcset != "" {
		// Try to extract a reasonable width from srcset
		// Format: "url 400w, url 800w, url 1200w"
		switch {
		case strings.Contains(srcset, "1200w"):
			defaultWidth = "1200"
		case strings.Contains(srcset, "800w"):
			defaultWidth = "800"
		case strings.Contains(srcset, "600w"):
			defaultWidth = "600"
		}
	}

	// Replace template placeholders with defaults
	resolved := src
	resolved = strings.ReplaceAll(resolved, "{width}", defaultWidth)
	resolved = strings.ReplaceAll(resolved, "{quality}", defaultQuality)
	resolved = strings.ReplaceAll(resolved, "{format}", defaultFormat)

	return resolved
}

// buildVideoMetadata creates a metadata map from VideoMetadata struct
// Centralizes the logic for converting video data to the result format.
func buildVideoMetadata(videoData *generic.VideoMetadata) map[string]interface{} {
	if videoData == nil {
		return nil
	}

	// Only build metadata if we have at least one URL
	if videoData.URL == "" && videoData.SecureURL == "" {
		return nil
	}

	metadata := make(map[string]interface{})

	if videoData.URL != "" {
		metadata["url"] = videoData.URL
	}
	if videoData.SecureURL != "" {
		metadata["secure_url"] = videoData.SecureURL
	}
	if videoData.Type != "" {
		metadata["type"] = videoData.Type
	}
	if videoData.Width > 0 {
		metadata["width"] = videoData.Width
	}
	if videoData.Height > 0 {
		metadata["height"] = videoData.Height
	}
	if videoData.Duration > 0 {
		metadata["duration"] = videoData.Duration
	}

	return metadata
}

// calculateWordCount calculates the number of words in text content.
func calculateWordCount(content string) int {
	if content == "" {
		return 0
	}

	// Simple word count by splitting on whitespace
	words := strings.Fields(stripHTMLTags(content))
	return len(words)
}

// buildMetaCache builds a cache of all meta tag names present in the document
// This is used to optimize meta tag extraction by only searching for names that exist.
func buildMetaCache(doc *goquery.Document) []string {
	var metaNames []string
	seen := make(map[string]bool)

	// Find all meta tags and collect their name and property attributes
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		// Check name attribute
		if name, exists := s.Attr("name"); exists && name != "" && !seen[name] {
			metaNames = append(metaNames, name)
			seen[name] = true
		}

		// Note: ExtractFromMeta only searches meta[name="..."] not meta[property="..."]
		// The property attributes (like og:title) are handled differently
		// We could enhance this to support property attributes in the future
	})

	return metaNames
}
