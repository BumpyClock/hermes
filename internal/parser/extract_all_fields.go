// ABOUTME: Core extraction orchestration that coordinates all field extractors with proper signatures and error handling
// ABOUTME: Handles the complete extraction pipeline from DOM to structured Result using all available extractors

package parser

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/hermes/internal/cleaners"
	"github.com/BumpyClock/hermes/internal/extractors/custom"
	"github.com/BumpyClock/hermes/internal/extractors/generic"
	"github.com/BumpyClock/hermes/internal/utils/security"
	"github.com/BumpyClock/hermes/internal/utils/text"
)

// extractAllFields orchestrates the complete extraction pipeline
func (m *Mercury) extractAllFields(doc *goquery.Document, targetURL string, parsedURL *url.URL, opts ParserOptions) (*Result, error) {
	// Use background context for backward compatibility
	// Callers should use extractAllFieldsWithContext for proper context handling
	return m.extractAllFieldsWithContext(context.Background(), doc, targetURL, parsedURL, opts)
}

// extractAllFieldsWithContext orchestrates the complete extraction pipeline with context support
func (m *Mercury) extractAllFieldsWithContext(ctx context.Context, doc *goquery.Document, targetURL string, parsedURL *url.URL, opts ParserOptions) (*Result, error) {
	// Check context before starting
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("extraction cancelled: %w", ctx.Err())
	default:
	}
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
	
	// Build meta cache first for use by both custom and generic extractors
	metaCache := buildMetaCache(doc)
	
	// Extract site metadata first (independent of custom/generic extractor choice)
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	// Start parallel site metadata extractions
	wg.Add(6)
	
	// Extract site name
	go func() {
		defer wg.Done()
		siteNameExtractor := &generic.GenericSiteNameExtractor{}
		if siteName := siteNameExtractor.Extract(doc.Selection, targetURL, metaCache); siteName != "" {
			mu.Lock()
			result.SiteName = siteName
			mu.Unlock()
		}
	}()
	
	// Extract site title  
	go func() {
		defer wg.Done()
		siteTitleExtractor := &generic.GenericSiteTitleExtractor{}
		if siteTitle := siteTitleExtractor.Extract(doc.Selection, targetURL, metaCache); siteTitle != "" {
			mu.Lock()
			result.SiteTitle = siteTitle
			mu.Unlock()
		}
	}()
	
	// Extract site image
	go func() {
		defer wg.Done()
		siteImageExtractor := &generic.GenericSiteImageExtractor{}
		if siteImage := siteImageExtractor.Extract(doc.Selection, targetURL, metaCache); siteImage != "" {
			mu.Lock()
			result.SiteImage = siteImage
			mu.Unlock()
		}
	}()
	
	// Extract favicon
	go func() {
		defer wg.Done()
		faviconExtractor := &generic.GenericFaviconExtractor{}
		if favicon := faviconExtractor.Extract(doc.Selection, targetURL, metaCache); favicon != "" {
			mu.Lock()
			result.Favicon = favicon
			mu.Unlock()
		}
	}()
	
	// Extract description
	go func() {
		defer wg.Done()
		descriptionExtractor := &generic.GenericDescriptionExtractor{}
		if description := descriptionExtractor.Extract(doc.Selection, targetURL, metaCache); description != "" {
			mu.Lock()
			result.Description = description
			mu.Unlock()
		}
	}()
	
	// Extract language
	go func() {
		defer wg.Done()
		languageExtractor := &generic.GenericLanguageExtractor{}
		if language := languageExtractor.Extract(doc.Selection, targetURL, metaCache); language != "" {
			mu.Lock()
			result.Language = language
			mu.Unlock()
		}
	}()
	
	// Wait for site metadata extraction to complete
	wg.Wait()
	
	// Check context after metadata extraction
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("extraction cancelled after metadata: %w", ctx.Err())
	default:
	}
	
	// Try to use custom extractor, passing the result with site metadata
	if customResult := m.tryCustomExtractor(doc, targetURL, parsedURL, opts, result); customResult != nil {
		return customResult, nil
	}

	// Parallel extraction for independent fields (meta cache already built)
	wg.Add(4) // Reset for generic extraction

	// Extract title in parallel
	go func() {
		defer wg.Done()
		if title := generic.GenericTitleExtractor.Extract(doc.Selection, targetURL, metaCache); title != "" {
			// First apply basic title cleaning
			cleanedTitle := cleaners.CleanTitle(title, targetURL, doc)
			// Then apply split title resolution to remove breadcrumbs and site names
			cleanedTitle = cleaners.ResolveSplitTitle(cleanedTitle, targetURL)
			mu.Lock()
			result.Title = cleanedTitle
			mu.Unlock()
		}
	}()

	// Extract author in parallel
	go func() {
		defer wg.Done()
		authorExtractor := &generic.GenericAuthorExtractor{}
		if author := authorExtractor.Extract(doc.Selection, metaCache); author != nil && *author != "" {
			cleanedAuthor := cleaners.CleanAuthor(*author)
			mu.Lock()
			result.Author = cleanedAuthor
			mu.Unlock()
		}
	}()

	// Extract date published in parallel
	go func() {
		defer wg.Done()
		if dateStr := generic.GenericDateExtractor.Extract(doc.Selection, targetURL, metaCache); dateStr != nil && *dateStr != "" {
			if date, err := parseDate(*dateStr); err == nil {
				mu.Lock()
				result.DatePublished = &date
				mu.Unlock()
			}
		}
	}()

	// Extract initial dek (description/subtitle) in parallel
	go func() {
		defer wg.Done()
		dekExtractor := &generic.GenericDekExtractor{}
		dekOpts := map[string]interface{}{
			"$": doc.Selection,
		}
		if dek := dekExtractor.Extract(doc, dekOpts); dek != "" {
			mu.Lock()
			result.Dek = dek
			mu.Unlock()
		}
	}()

	// Wait for all parallel extractions to complete
	wg.Wait()
	
	// Check context after parallel extraction
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("extraction cancelled after parallel extraction: %w", ctx.Err())
	default:
	}

	// Extract lead image URL (needs to be done after parallel extraction for content dependency)
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
		dekExtractor := &generic.GenericDekExtractor{}
		dekOpts := map[string]interface{}{
			"$":       doc.Selection,
			"excerpt": result.Excerpt,
		}
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
func (m *Mercury) tryCustomExtractor(doc *goquery.Document, targetURL string, parsedURL *url.URL, opts ParserOptions, baseResult *Result) *Result {
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
		// No custom extractor found
		return nil // No custom extractor found
	}
	
	// Log successful custom extractor selection (optional debug)
	_ = usedDomain // Suppress unused variable warning
	
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
			} else if selectorArray, ok := selector.([]string); ok && len(selectorArray) >= 2 {
				// Handle array selectors like ["meta[name='author']", "content"]
				if authorEl := doc.Find(selectorArray[0]).First(); authorEl.Length() > 0 {
					if author := strings.TrimSpace(authorEl.AttrOr(selectorArray[1], "")); author != "" {
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
			var contentHTML string
			
			// Handle array selectors (multi-match like [".c-entry-hero .e-image", ".c-entry-intro", ".c-entry-content"])
			if selectorArray, ok := selector.([]interface{}); ok {
				var combinedContent strings.Builder
				for _, selectorItem := range selectorArray {
					if selectorStr, ok := selectorItem.(string); ok {
						contentElements := doc.Find(selectorStr)
						if contentElements.Length() > 0 {
							contentElements.Each(func(i int, el *goquery.Selection) {
								if html, err := el.Html(); err == nil && strings.TrimSpace(html) != "" {
									combinedContent.WriteString(html)
									combinedContent.WriteString("\n")
								}
							})
						}
					}
				}
				contentHTML = strings.TrimSpace(combinedContent.String())
			} else if selectorStr, ok := selector.(string); ok {
				// Handle single string selectors - get ALL matching elements
				contentElements := doc.Find(selectorStr)
				if contentElements.Length() > 0 {
					var combinedContent strings.Builder
					contentElements.Each(func(i int, el *goquery.Selection) {
						if html, err := el.Html(); err == nil && strings.TrimSpace(html) != "" {
							combinedContent.WriteString(html)
							combinedContent.WriteString("\n")
						}
					})
					contentHTML = strings.TrimSpace(combinedContent.String())
				}
			}
			
			// If we found content, process it and break
			if contentHTML != "" && strings.TrimSpace(contentHTML) != "" {
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
	
	// Extract lead image URL using custom selectors
	if customExtractor.LeadImageURL != nil && len(customExtractor.LeadImageURL.Selectors) > 0 {
		for _, selector := range customExtractor.LeadImageURL.Selectors {
			if selectorStr, ok := selector.(string); ok {
				if imageEl := doc.Find(selectorStr).First(); imageEl.Length() > 0 {
					if imageURL := strings.TrimSpace(imageEl.Text()); imageURL != "" {
						result.LeadImageURL = cleaners.CleanLeadImageURL(imageURL, targetURL)
						break
					}
				}
			} else if selectorArray, ok := selector.([]string); ok && len(selectorArray) >= 2 {
				// Handle array selectors like ["meta[property='og:image']", "content"]
				if imageEl := doc.Find(selectorArray[0]).First(); imageEl.Length() > 0 {
					if imageURL := strings.TrimSpace(imageEl.AttrOr(selectorArray[1], "")); imageURL != "" {
						result.LeadImageURL = cleaners.CleanLeadImageURL(imageURL, targetURL)
						break
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
	
	// Extract site metadata for custom extractors too (independent of content extraction)
	metaCache := buildMetaCache(doc)
	
	// Site name extraction
	siteNameExtractor := &generic.GenericSiteNameExtractor{}
	if siteName := siteNameExtractor.Extract(doc.Selection, targetURL, metaCache); siteName != "" {
		result.SiteName = siteName
	}
	
	// Site title extraction  
	siteTitleExtractor := &generic.GenericSiteTitleExtractor{}
	if siteTitle := siteTitleExtractor.Extract(doc.Selection, targetURL, metaCache); siteTitle != "" {
		result.SiteTitle = siteTitle
	}
	
	// Site image extraction
	siteImageExtractor := &generic.GenericSiteImageExtractor{}
	if siteImage := siteImageExtractor.Extract(doc.Selection, targetURL, metaCache); siteImage != "" {
		result.SiteImage = siteImage
	}
	
	// Favicon extraction
	faviconExtractor := &generic.GenericFaviconExtractor{}
	if favicon := faviconExtractor.Extract(doc.Selection, targetURL, metaCache); favicon != "" {
		result.Favicon = favicon
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