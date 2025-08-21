// ABOUTME: Core extraction orchestration that coordinates all field extractors with proper signatures and error handling
// ABOUTME: Handles the complete extraction pipeline from DOM to structured Result using all available extractors

package parser

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/parser-go/pkg/cleaners"
	"github.com/BumpyClock/parser-go/pkg/extractors/custom"
	"github.com/BumpyClock/parser-go/pkg/extractors/generic"
	"github.com/BumpyClock/parser-go/pkg/utils/security"
	"github.com/BumpyClock/parser-go/pkg/utils/text"
)

// extractAllFields orchestrates the complete extraction pipeline
func (m *Mercury) extractAllFields(doc *goquery.Document, targetURL string, parsedURL *url.URL, opts ParserOptions) (*Result, error) {
	// Merge provided options with defaults to ensure reasonable behavior
	if opts.ContentType == "" {
		opts.ContentType = "html"
	}
	// Enable fallback by default if no explicit preference is given
	// This is detected by checking if ALL options are zero values (empty struct)
	if opts.ContentType == "html" && !opts.Fallback && opts.Headers == nil && !opts.FetchAllPages {
		// Likely an empty ParserOptions{}, so enable fallback for better UX
		opts.Fallback = true
	}
	
	// Create base result
	result := &Result{
		URL:    targetURL,
		Domain: parsedURL.Host,
	}
	
	// Try to use custom extractor first if available
	if customResult := m.tryCustomExtractor(doc, targetURL, parsedURL, opts); customResult != nil {
		return customResult, nil
	}

	// Build meta cache by scanning all meta tags in the document
	metaCache := buildMetaCache(doc)

	// Extract title
	if title := generic.GenericTitleExtractor.Extract(doc.Selection, targetURL, metaCache); title != "" {
		// First apply basic title cleaning
		result.Title = cleaners.CleanTitle(title, targetURL, doc)
		// Then apply split title resolution to remove breadcrumbs and site names
		result.Title = cleaners.ResolveSplitTitle(result.Title, targetURL)
	}

	// Extract author
	authorExtractor := &generic.GenericAuthorExtractor{}
	if author := authorExtractor.Extract(doc.Selection, metaCache); author != nil && *author != "" {
		result.Author = cleaners.CleanAuthor(*author)
	}

	// Extract date published
	if dateStr := generic.GenericDateExtractor.Extract(doc.Selection, targetURL, metaCache); dateStr != nil && *dateStr != "" {
		if date, err := parseDate(*dateStr); err == nil {
			result.DatePublished = &date
		}
	}

	// Extract lead image URL
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

	// Extract dek (description/subtitle)
	dekExtractor := &generic.GenericDekExtractor{}
	dekOpts := map[string]interface{}{
		"$": doc.Selection,
	}
	if dek := dekExtractor.Extract(doc, dekOpts); dek != "" {
		result.Dek = dek
	}

	// Extract main content
	contentExtractor := generic.NewGenericContentExtractor()
	contentParams := generic.ExtractorParams{
		Doc:   doc,
		HTML:  "", // Could enhance with original HTML
		Title: result.Title,
		URL:   targetURL,
	}
	contentOpts := generic.ExtractorOptions{
		StripUnlikelyCandidates: true,
		WeightNodes:             true,
		CleanConditionally:      true,
	}
	if content := contentExtractor.Extract(contentParams, contentOpts); content != "" {
		// Apply content type conversion with security sanitization
		switch strings.ToLower(opts.ContentType) {
		case "text":
			result.Content = text.NormalizeSpaces(stripHTMLTags(content))
		case "markdown":
			result.Content = convertToMarkdown(content)
		default: // "html" or anything else
			// Sanitize HTML content to prevent XSS attacks
			result.Content = security.SanitizeHTML(content)
		}
		
		// Extract excerpt if content exists
		if result.Content != "" {
			result.Excerpt = text.ExcerptContent(result.Content, 160)
		}
		
		// Calculate word count
		result.WordCount = calculateWordCount(result.Content)

		// Update image extraction with content context
		imageParams.Content = result.Content
		if imageURL := imageExtractor.Extract(imageParams); imageURL != nil && *imageURL != "" && result.LeadImageURL == "" {
			result.LeadImageURL = cleaners.CleanLeadImageURL(*imageURL, targetURL)
		}

		// Update dek with excerpt context
		dekOpts["excerpt"] = result.Excerpt
		if dek := dekExtractor.Extract(doc, dekOpts); dek != "" && result.Dek == "" {
			result.Dek = dek
		}
	}

	// Set default values for fields not extracted
	if result.Title == "" && opts.Fallback {
		// Fallback title extraction
		if title := doc.Find("title").First().Text(); title != "" {
			result.Title = cleaners.CleanTitleSimple(strings.TrimSpace(title), targetURL)
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

// tryCustomExtractor attempts to use a custom extractor for the given domain
func (m *Mercury) tryCustomExtractor(doc *goquery.Document, targetURL string, parsedURL *url.URL, opts ParserOptions) *Result {
	// Look for custom extractor for this domain using the proper lookup function
	customExtractor, found := custom.GetCustomExtractorByDomain(parsedURL.Host)
	var usedDomain = parsedURL.Host
	
	if !found {
		// Try fallback - remove 'www.' prefix if present
		if strings.HasPrefix(parsedURL.Host, "www.") {
			baseDomain := strings.TrimPrefix(parsedURL.Host, "www.")
			customExtractor, found = custom.GetCustomExtractorByDomain(baseDomain)
			if found {
				usedDomain = baseDomain
			}
		} else {
			// Try adding 'www.' prefix
			wwwDomain := "www." + parsedURL.Host
			customExtractor, found = custom.GetCustomExtractorByDomain(wwwDomain)
			if found {
				usedDomain = wwwDomain
			}
		}
	}
	
	if !found || customExtractor == nil {
		// Log when no custom extractor is found for debugging
		if len(parsedURL.Host) > 0 {
			fmt.Printf("DEBUG: No custom extractor found for domain: %s\n", parsedURL.Host)
		}
		return nil // No custom extractor found
	}
	
	// Log successful custom extractor selection
	fmt.Printf("DEBUG: Using custom extractor for domain: %s (matched: %s)\n", parsedURL.Host, usedDomain)
	
	// Create result with custom extractor info
	result := &Result{
		URL:           targetURL,
		Domain:        parsedURL.Host,
		ExtractorUsed: "custom:" + customExtractor.Domain,
	}
	
	// Extract title using custom selectors
	if customExtractor.Title != nil && len(customExtractor.Title.Selectors) > 0 {
		for _, selector := range customExtractor.Title.Selectors {
			if selectorStr, ok := selector.(string); ok {
				if titleEl := doc.Find(selectorStr).First(); titleEl.Length() > 0 {
					if title := strings.TrimSpace(titleEl.Text()); title != "" {
						result.Title = cleaners.CleanTitle(title, targetURL, doc)
						break
					}
				}
			}
		}
	}
	
	// Extract author using custom selectors
	if customExtractor.Author != nil && len(customExtractor.Author.Selectors) > 0 {
		for _, selector := range customExtractor.Author.Selectors {
			if selectorStr, ok := selector.(string); ok {
				if authorEl := doc.Find(selectorStr).First(); authorEl.Length() > 0 {
					if author := strings.TrimSpace(authorEl.Text()); author != "" {
						result.Author = cleaners.CleanAuthor(author)
						break
					}
				}
			}
		}
	}
	
	// Extract content using custom selectors
	if customExtractor.Content != nil && len(customExtractor.Content.Selectors) > 0 {
		for _, selector := range customExtractor.Content.Selectors {
			if selectorStr, ok := selector.(string); ok {
				if contentEl := doc.Find(selectorStr).First(); contentEl.Length() > 0 {
					if contentHTML, err := contentEl.Html(); err == nil && strings.TrimSpace(contentHTML) != "" {
						// Apply content type conversion with security sanitization
						switch strings.ToLower(opts.ContentType) {
						case "text":
							result.Content = text.NormalizeSpaces(stripHTMLTags(contentHTML))
						case "markdown":
							result.Content = convertToMarkdown(contentHTML)
						default: // "html" or anything else
							result.Content = security.SanitizeHTML(contentHTML)
						}
						
						// Extract excerpt if content exists
						if result.Content != "" {
							result.Excerpt = text.ExcerptContent(result.Content, 160)
						}
						
						// Calculate word count
						result.WordCount = calculateWordCount(result.Content)
						break
					}
				}
			}
		}
	}
	
	// Extract date using custom selectors
	if customExtractor.DatePublished != nil && len(customExtractor.DatePublished.Selectors) > 0 {
		for _, selector := range customExtractor.DatePublished.Selectors {
			// Handle array selectors like [".dateblock time[datetime]", "datetime"]
			if selectorArray, ok := selector.([]string); ok && len(selectorArray) >= 2 {
				if dateEl := doc.Find(selectorArray[0]).First(); dateEl.Length() > 0 {
					if dateStr := strings.TrimSpace(dateEl.AttrOr(selectorArray[1], "")); dateStr != "" {
						if date, err := parseDate(dateStr); err == nil {
							result.DatePublished = &date
							break
						}
					}
				}
			} else if selectorStr, ok := selector.(string); ok {
				if dateEl := doc.Find(selectorStr).First(); dateEl.Length() > 0 {
					if dateStr := strings.TrimSpace(dateEl.Text()); dateStr != "" {
						if date, err := parseDate(dateStr); err == nil {
							result.DatePublished = &date
							break
						}
					}
				}
			}
		}
	}
	
	// Fall back to generic extractors for missing fields if fallback is enabled
	if opts.Fallback {
		metaCache := buildMetaCache(doc)
		
		// Fallback title extraction
		if result.Title == "" {
			if title := generic.GenericTitleExtractor.Extract(doc.Selection, targetURL, metaCache); title != "" {
				result.Title = cleaners.CleanTitle(title, targetURL, doc)
			}
		}
		
		// Fallback author extraction
		if result.Author == "" {
			authorExtractor := &generic.GenericAuthorExtractor{}
			if author := authorExtractor.Extract(doc.Selection, metaCache); author != nil && *author != "" {
				result.Author = cleaners.CleanAuthor(*author)
			}
		}
		
		// Fallback date extraction
		if result.DatePublished == nil {
			if dateStr := generic.GenericDateExtractor.Extract(doc.Selection, targetURL, metaCache); dateStr != nil && *dateStr != "" {
				if date, err := parseDate(*dateStr); err == nil {
					result.DatePublished = &date
				}
			}
		}
		
		// Fallback content extraction if no content was found
		if result.Content == "" {
			contentExtractor := generic.NewGenericContentExtractor()
			contentParams := generic.ExtractorParams{
				Doc:   doc,
				HTML:  "",
				Title: result.Title,
				URL:   targetURL,
			}
			contentOpts := generic.ExtractorOptions{
				StripUnlikelyCandidates: true,
				WeightNodes:             true,
				CleanConditionally:      true,
			}
			if content := contentExtractor.Extract(contentParams, contentOpts); content != "" {
				switch strings.ToLower(opts.ContentType) {
				case "text":
					result.Content = text.NormalizeSpaces(stripHTMLTags(content))
				case "markdown":
					result.Content = convertToMarkdown(content)
				default:
					result.Content = security.SanitizeHTML(content)
				}
				
				if result.Content != "" {
					result.Excerpt = text.ExcerptContent(result.Content, 160)
					result.WordCount = calculateWordCount(result.Content)
				}
			}
		}
	}
	
	return result
}

// parseDate parses a date string into a time.Time
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
func stripHTMLTags(content string) string {
	// Create a temporary document to extract text
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		// If parsing fails, return original content
		return content
	}
	return doc.Text()
}

// convertToMarkdown converts HTML content to Markdown using html-to-markdown library
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

// resolveImageTemplateURL resolves template placeholders in responsive image URLs
func resolveImageTemplateURL(src string, imgElement *goquery.Selection) string {
	// Check if URL contains template placeholders
	if !strings.Contains(src, "{width}") && !strings.Contains(src, "{quality}") && !strings.Contains(src, "{format}") {
		return src // No templates, return as-is
	}
	
	// Look for reasonable default values to replace templates
	// These are common web standards that should work for most images
	defaultWidth := "1200"   // Reasonable default width
	defaultQuality := "85"   // Good balance of quality vs size
	defaultFormat := "jpeg"  // Most compatible format
	
	// Try to get better values from the element's attributes
	if width, exists := imgElement.Attr("width"); exists && width != "" {
		defaultWidth = width
	}
	
	// Check for srcset or other attributes that might give us hints
	if srcset, exists := imgElement.Attr("srcset"); exists && srcset != "" {
		// Try to extract a reasonable width from srcset
		// Format: "url 400w, url 800w, url 1200w"
		if strings.Contains(srcset, "1200w") {
			defaultWidth = "1200"
		} else if strings.Contains(srcset, "800w") {
			defaultWidth = "800"
		} else if strings.Contains(srcset, "600w") {
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

// calculateWordCount calculates the number of words in text content
func calculateWordCount(content string) int {
	if content == "" {
		return 0
	}
	
	// Simple word count by splitting on whitespace
	words := strings.Fields(stripHTMLTags(content))
	return len(words)
}

// buildMetaCache builds a cache of all meta tag names present in the document
// This is used to optimize meta tag extraction by only searching for names that exist
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