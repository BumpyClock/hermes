// ABOUTME: Core extraction orchestration that coordinates all field extractors with proper signatures and error handling
// ABOUTME: Handles the complete extraction pipeline from DOM to structured Result using all available extractors

package parser

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/escape"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"

	"github.com/BumpyClock/hermes/internal/cleaners"
	"github.com/BumpyClock/hermes/internal/extractors"
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

	// Try an explicitly configured definition rule, preserving site metadata.
	definitionResult, err := h.tryDefinitionExtractor(ctx, doc, targetURL, parsedURL, opts, result, metaCache)
	if err != nil {
		return nil, err
	}
	if definitionResult != nil {
		return definitionResult, nil
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
		if opts.DefinitionsConfigured {
			content = security.SanitizeHTML(content)
		}
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
				// DOM text can contain literal markup from raw-text elements.
				result.Content = formatPlainText(basicContent, opts.ContentType)
				result.Excerpt = text.ExcerptContent(result.Content, 160)
				result.WordCount = calculateWordCount(result.Content)
				break
			}
		}
	}

	return result, nil
}

// tryDefinitionExtractor applies an explicitly configured external definition snapshot.
func (h *Hermes) tryDefinitionExtractor(ctx context.Context, doc *goquery.Document, targetURL string, parsedURL *url.URL, opts ParserOptions, baseResult *Result, metaCache []string) (*Result, error) {
	if !opts.DefinitionsConfigured || opts.Definitions == nil {
		return nil, nil
	}
	definitionExtractor := opts.Definitions.Match(parsedURL.Hostname())
	if definitionExtractor == nil {
		return nil, nil
	}

	// Create result with definition-rule metadata, preserving generic site metadata.
	result := &Result{
		URL:           targetURL,
		Domain:        parsedURL.Host,
		ExtractorUsed: "definition:" + definitionExtractor.Domain,
		// Preserve site metadata
		SiteName:    baseResult.SiteName,
		SiteTitle:   baseResult.SiteTitle,
		SiteImage:   baseResult.SiteImage,
		Favicon:     baseResult.Favicon,
		Description: baseResult.Description,
		Language:    baseResult.Language,
		ThemeColor:  baseResult.ThemeColor,
	}

	if title, fieldErr := firstDefinitionField(ctx, doc, definitionExtractor.Title); fieldErr != nil {
		return nil, fieldErr
	} else if title != "" {
		result.Title = cleaners.CleanTitle(title, targetURL, doc)
	}
	if author, fieldErr := firstDefinitionField(ctx, doc, definitionExtractor.Author); fieldErr != nil {
		return nil, fieldErr
	} else if author != "" {
		result.Author = cleaners.CleanAuthor(author)
	}

	outputBaseURL := ""
	if opts.DefinitionsConfigured {
		outputBaseURL = targetURL
		if baseHref := doc.Find("base[href]").First().AttrOr("href", ""); baseHref != "" {
			if resolvedBase, err := parsedURL.Parse(baseHref); err == nil {
				outputBaseURL = resolvedBase.String()
			}
		}
	}

	// Extract content using custom selectors
	if definitionExtractor.Content != nil && len(definitionExtractor.Content.Selectors) > 0 {
		for _, selector := range definitionExtractor.Content.Selectors {
			contentElements := contentElementsForSelector(doc, selector)

			// Process the first selector with non-empty raw content, preserving fallback order.
			if contentElements == nil || !hasCustomContent(contentElements) {
				continue
			}

			var contentHTML string
			var err error
			if definitionExtractor.Content.OrderedTransforms != nil {
				contentHTML, err = processOrderedContent(ctx, contentElements, doc, definitionExtractor.Content, result.Title, targetURL, outputBaseURL)
			} else {
				contentHTML, err = processDefinitionContentWithBase(contentElements, doc, definitionExtractor.Content, result.Title, targetURL, outputBaseURL)
			}
			if err != nil && opts.DefinitionsConfigured {
				return nil, fmt.Errorf("extract definition site %q content: %w", definitionExtractor.Domain, err)
			}
			if err == nil && strings.TrimSpace(contentHTML) != "" {
				if opts.DefinitionsConfigured {
					contentHTML = security.SanitizeHTML(contentHTML)
				}
				setFormattedContent(result, contentHTML, opts.ContentType)
			}
			break
		}
	}

	if definitionExtractor.DatePublished != nil {
		for _, selector := range definitionExtractor.DatePublished.Selectors {
			value, fieldErr := definitionFieldValue(ctx, doc, selector)
			if fieldErr != nil {
				return nil, fieldErr
			}
			if date, err := parseDate(value); err == nil {
				result.DatePublished = &date
				break
			}
		}
	}
	if imageURL, fieldErr := firstDefinitionField(ctx, doc, definitionExtractor.LeadImageURL); fieldErr != nil {
		return nil, fieldErr
	} else if imageURL != "" {
		imageBaseURL := targetURL
		if opts.DefinitionsConfigured {
			imageBaseURL = outputBaseURL
		}
		result.LeadImageURL = cleaners.CleanLeadImageURL(imageURL, imageBaseURL)
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
				if opts.DefinitionsConfigured {
					content = security.SanitizeHTML(content)
				}
				setFormattedContent(result, content, opts.ContentType)
			}
		}
	}

	extractVideoMetadata(doc, targetURL, metaCache, result)

	return result, nil
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

func formatPlainText(content, contentType string) string {
	content = strings.TrimSpace(content)
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "text", "text/plain", "txt":
		return text.NormalizeSpaces(content)
	case "markdown", "md", "text/markdown":
		// Escape CommonMark punctuation directly: HTML conversion decodes entities
		// and can turn literal text back into raw HTML or Markdown syntax.
		var escaped strings.Builder
		for _, char := range content {
			if char >= '!' && char <= '/' || char >= ':' && char <= '@' ||
				char >= '[' && char <= '`' || char >= '{' && char <= '~' {
				escaped.WriteByte('\\')
			}
			escaped.WriteRune(char)
		}
		return escaped.String()
	default:
		return formatContent(html.EscapeString(content), contentType)
	}
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

func definitionFieldValue(ctx context.Context, doc *goquery.Document, selector extractors.SelectorEntry) (string, error) {
	element := doc.Find(selector.Selector).First()
	if element.Length() == 0 {
		return "", nil
	}
	if selector.Attribute != "" {
		return strings.TrimSpace(element.AttrOr(selector.Attribute, "")), nil
	}
	if selector.Capture == nil {
		return strings.TrimSpace(element.Text()), nil
	}
	return capturedFieldText(ctx, element.Get(0), selector.Selector, selector.Capture)
}

func firstDefinitionField(ctx context.Context, doc *goquery.Document, extractor *extractors.FieldExtractor) (string, error) {
	if extractor != nil {
		for _, selector := range extractor.Selectors {
			value, err := definitionFieldValue(ctx, doc, selector)
			if err != nil {
				return "", err
			}
			if value != "" {
				return value, nil
			}
		}
	}
	return "", nil
}

func capturedFieldText(ctx context.Context, root *html.Node, selector string, capture *extractors.TextCapture) (string, error) {
	type nodeDepth struct {
		node  *html.Node
		depth int
	}
	stack := []nodeDepth{{node: root}}
	var text strings.Builder
	nodes := 0
	for len(stack) > 0 {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		nodes++
		if nodes > capture.MaxNodes || current.depth > capture.MaxDepth {
			return "", &extractors.MetadataCaptureError{Selector: selector, Err: fmt.Errorf("%w: exceeds %d nodes or depth %d", extractors.ErrMetadataCaptureLimit, capture.MaxNodes, capture.MaxDepth)}
		}
		if current.node.Type == html.TextNode {
			if text.Len()+len(current.node.Data) > capture.MaxBytes {
				return "", &extractors.MetadataCaptureError{Selector: selector, Err: fmt.Errorf("%w: exceeds %d bytes", extractors.ErrMetadataCaptureLimit, capture.MaxBytes)}
			}
			text.WriteString(current.node.Data)
		}
		for child := current.node.LastChild; child != nil; child = child.PrevSibling {
			stack = append(stack, nodeDepth{node: child, depth: current.depth + 1})
		}
	}
	match := capture.Pattern.FindStringSubmatchIndex(text.String())
	if match == nil || match[capture.Group*2] < 0 {
		return "", nil
	}
	return strings.TrimSpace(text.String()[match[capture.Group*2]:match[capture.Group*2+1]]), nil
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

func contentElementsForSelector(doc *goquery.Document, selectors extractors.ContentSelectorGroup) *goquery.Selection {
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

func processDefinitionContent(contentElements *goquery.Selection, doc *goquery.Document, extractor *extractors.ContentExtractor, title, targetURL string) (string, error) {
	return processDefinitionContentWithBase(contentElements, doc, extractor, title, targetURL, "")
}

func processDefinitionContentWithBase(contentElements *goquery.Selection, doc *goquery.Document, extractor *extractors.ContentExtractor, title, targetURL, outputBaseURL string) (string, error) {
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

		for _, selector := range extractor.Clean {
			wrapper.Find(selector).Remove()
		}
		if outputBaseURL != "" {
			// Rules see source attributes; only the output clone receives absolute URLs.
			contentDoc.Find("base").Remove()
			dom.MakeLinksAbsolute(contentDoc, outputBaseURL)
		}

		content := wrapper.Children().First()
		if !extractor.DisableDefaultCleaner {
			content = cleaners.ExtractCleanNode(content, doc, cleaners.ContentCleanOptions{
				CleanConditionally: true,
				Title:              title,
				URL:                targetURL,
				Preserve:           extractor.Preserve,
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
		"2006-01-02 15:04:05 UTC",
		"2006-01-02",
		"January 2, 2006",
		"Jan 2, 2006",
		"2 January 2006",
		"2 Jan 2006",
		"2006/01/02",
		"2006/1/2",
		"2006年1月2日",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

var (
	markdownTextSpaces   = regexp.MustCompile(`[\t ]+`)
	markdownTextEntities = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
)

// convertToMarkdown converts HTML content to Markdown using html-to-markdown library.
func convertToMarkdown(content string) string {
	// Create converter with options similar to TurndownService
	converter := md.NewConverter("", true, nil)

	// Configure options to match TurndownService behavior
	converter.Use(md.Plugin(func(c *md.Converter) []md.Rule {
		return []md.Rule{
			{
				Filter: []string{"#text"},
				Replacement: func(_ string, selection *goquery.Selection, _ *md.Options) *string {
					value := selection.Text()
					if strings.ContainsAny(value, "<>&") {
						// Encode literal HTML/entity syntax before Markdown escaping.
						// Code rules read the original DOM instead of this serialized text.
						value = markdownTextSpaces.ReplaceAllString(value, " ")
						return md.String(escape.MarkdownCharacters(markdownTextEntities.Replace(value)))
					}
					if !preservesInlineWhitespace(selection, value) {
						return nil
					}
					value = strings.NewReplacer("\t", " ", "\r", " ", "\n", " ").Replace(value)
					return md.String(escape.MarkdownCharacters(markdownTextSpaces.ReplaceAllString(value, " ")))
				},
			},
			{
				Filter: []string{"strong", "b"},
				Replacement: func(content string, selection *goquery.Selection, _ *md.Options) *string {
					if wrapped := renderInlineWrapperWhitespace(content, selection, "**"); wrapped != nil {
						return wrapped
					}
					if rendered := renderFlankedStrong(selection, content); rendered != nil {
						return rendered
					}
					return nil
				},
			},
			{
				Filter: []string{"i", "em"},
				Replacement: func(content string, selection *goquery.Selection, _ *md.Options) *string {
					if wrapped := renderInlineWrapperWhitespace(content, selection, "_"); wrapped != nil {
						return wrapped
					}
					if !canRenderAdjacentEmphasis(selection, content) {
						return nil
					}
					// Asterisks preserve valid adjacent emphasis without source-absent padding.
					return md.String("*" + content + "*")
				},
			},
			{
				Filter: []string{"caption", "figcaption"},
				Replacement: func(content string, _ *goquery.Selection, _ *md.Options) *string {
					content = markdownBlockContent(content)
					if content == "" {
						return md.String("")
					}
					return md.String("\n\n" + content + "\n\n")
				},
			},
			{
				Filter: []string{"tr"},
				Replacement: func(content string, _ *goquery.Selection, _ *md.Options) *string {
					content = markdownBlockContent(content)
					if content == "" {
						return md.String("")
					}
					return md.String("\n\n" + content + "\n\n")
				},
			},
			{
				Filter: []string{"th", "td"},
				Replacement: func(content string, _ *goquery.Selection, _ *md.Options) *string {
					content = markdownBlockContent(content)
					if content == "" {
						return md.String("")
					}
					return md.String(content + "\n")
				},
			},
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
			{
				Filter: []string{"a"},
				Replacement: func(content string, selec *goquery.Selection, _ *md.Options) *string {
					// Preserve source spacing; the upstream link rule can add a separator absent from the DOM.
					return markdownLink(content, selec)
				},
			},
		}
	}))

	// Convert HTML to Markdown
	markdown, err := converter.ConvertString(content)
	if err != nil {
		// Fallback to text extraction if conversion fails
		return dom.StripTags(content)
	}

	return markdown
}

func markdownBlockContent(content string) string {
	return preserveMarkdownUnicodeEdges(strings.TrimFunc(content, isHTMLWhitespace))
}

func preserveMarkdownUnicodeEdges(content string) string {
	characters := []rune(content)
	start, end := 0, len(characters)
	for start < end && unicode.IsSpace(characters[start]) {
		start++
	}
	for end > start && unicode.IsSpace(characters[end-1]) {
		end--
	}
	if start == 0 && end == len(characters) {
		return content
	}
	// The converter trims Unicode space at its outer boundary; references retain visible source spacing.
	var output strings.Builder
	writeBoundary := func(characters []rune) {
		for _, character := range characters {
			if isHTMLWhitespace(character) {
				output.WriteRune(character)
			} else {
				fmt.Fprintf(&output, "&#%d;", character)
			}
		}
	}
	writeBoundary(characters[:start])
	output.WriteString(string(characters[start:end]))
	writeBoundary(characters[end:])
	return output.String()
}

func markdownLink(content string, selection *goquery.Selection) *string {
	href := selection.AttrOr("href", "")
	if href == "" {
		return md.String(content)
	}
	leadingWhitespace, content, trailingWhitespace := splitHTMLWhitespace(content)
	if content == "" {
		if leadingWhitespace != "" || trailingWhitespace != "" {
			return md.String(" ")
		}
		content = selection.AttrOr("title", selection.AttrOr("aria-label", ""))
	}
	if content == "" {
		return md.String("")
	}
	content = escapeMarkdownLinkLabel(content)

	title := ""
	if value, ok := selection.Attr("title"); ok {
		title = ` "` + escapeMarkdownLinkTitle(value) + `"`
	}
	result := fmt.Sprintf("[%s](%s%s)", content, href, title)
	if leadingWhitespace != "" {
		result = leadingWhitespace + result
	}
	if trailingWhitespace != "" {
		result += trailingWhitespace
	}
	return &result
}

func splitHTMLWhitespace(value string) (leading string, content string, trailing string) {
	runes := []rune(value)
	start := 0
	for start < len(runes) && isHTMLWhitespace(runes[start]) {
		start++
	}
	end := len(runes)
	for end > start && isHTMLWhitespace(runes[end-1]) {
		end--
	}
	if start > 0 {
		leading = " "
	}
	if end < len(runes) {
		trailing = " "
	}
	return leading, string(runes[start:end]), trailing
}

func isHTMLWhitespace(character rune) bool {
	switch character {
	case '\t', '\n', '\f', '\r', ' ':
		return true
	default:
		return false
	}
}

func escapeMarkdownLinkLabel(value string) string {
	value = strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(value)
	return strings.ReplaceAll(value, "\n", "\\\n")
}

func escapeMarkdownLinkTitle(value string) string {
	value = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ").Replace(value)
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}

func preservesInlineWhitespace(selection *goquery.Selection, value string) bool {
	node := selection.Get(0)
	if node == nil {
		return false
	}
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Type == html.ElementNode && md.IsInlineElement(parent.Data) {
			return true
		}
	}
	if strings.TrimSpace(value) != "" || node.PrevSibling == nil || node.NextSibling == nil {
		return false
	}
	return isInlineOrText(node.PrevSibling) && isInlineOrText(node.NextSibling)
}

func isInlineOrText(node *html.Node) bool {
	return node.Type == html.TextNode && strings.TrimSpace(node.Data) != "" ||
		node.Type == html.ElementNode && md.IsInlineElement(node.Data)
}

func renderInlineWrapperWhitespace(content string, selection *goquery.Selection, delimiter string) *string {
	if strings.ContainsAny(content, "\r\n") || selection.Find("strong,b,em,i,code").Length() != 0 {
		return nil
	}
	leading, inner, trailing := splitInlineWrapperWhitespace(content)
	if leading == "" && trailing == "" {
		return nil
	}
	if inner == "" {
		return md.String(preserveMarkdownUnicodeEdges(leading + trailing))
	}
	return md.String(preserveMarkdownUnicodeEdges(leading + delimiter + inner + delimiter + trailing))
}

func splitInlineWrapperWhitespace(value string) (leading, inner, trailing string) {
	runes := []rune(value)
	start := 0
	for start < len(runes) && unicode.IsSpace(runes[start]) {
		start++
	}
	end := len(runes)
	for end > start && unicode.IsSpace(runes[end-1]) {
		end--
	}
	return normalizeInlineWrapperWhitespace(runes[:start]), string(runes[start:end]), normalizeInlineWrapperWhitespace(runes[end:])
}

func normalizeInlineWrapperWhitespace(value []rune) string {
	var output strings.Builder
	previousASCIIWhitespace := false
	for _, character := range value {
		if isHTMLWhitespace(character) {
			if !previousASCIIWhitespace {
				output.WriteByte(' ')
			}
			previousASCIIWhitespace = true
			continue
		}
		output.WriteRune(character)
		previousASCIIWhitespace = false
	}
	return output.String()
}

func canRenderAdjacentEmphasis(selection *goquery.Selection, content string) bool {
	if parent := selection.Parent(); parent.Is("i,em") {
		return false
	}
	return canRenderAdjacentAsterisk(selection, content, "strong,b,em,i,code")
}

func canRenderAdjacentAsterisk(selection *goquery.Selection, content, nestedSelector string) bool {
	node := selection.Get(0)
	if node == nil || content == "" || content != strings.TrimSpace(content) ||
		strings.ContainsAny(content, "\r\n") || selection.Find(nestedSelector).Length() != 0 {
		return false
	}

	if parent := selection.Parent(); parent.Is("strong,b") {
		return false
	}
	before, hasBefore := lastTextRune(node.PrevSibling)
	after, hasAfter := firstTextRune(node.NextSibling)
	if node.PrevSibling == nil {
		before, hasBefore = ' ', true
	}
	if node.NextSibling == nil {
		after, hasAfter = ' ', true
	}
	if !hasBefore || !hasAfter {
		return false
	}
	if unicode.IsSpace(before) && unicode.IsSpace(after) {
		return false
	}
	characters := []rune(content)
	return leftFlankingAsterisk(before, characters[0]) &&
		rightFlankingAsterisk(characters[len(characters)-1], after)
}

func renderFlankedStrong(selection *goquery.Selection, content string) *string {
	node := selection.Get(0)
	if node == nil || content == "" || content != strings.TrimSpace(content) ||
		strings.ContainsAny(content, "\r\n") || selection.Find("strong,b,em,i,code").Length() != 0 {
		return nil
	}
	if parent := selection.Parent(); parent.Is("strong,b") {
		return nil
	}
	before, hasBefore := lastTextRune(node.PrevSibling)
	after, hasAfter := firstTextRune(node.NextSibling)
	if node.PrevSibling == nil {
		before, hasBefore = ' ', true
	}
	if node.NextSibling == nil {
		after, hasAfter = ' ', true
	}
	if !hasBefore || !hasAfter {
		return nil
	}
	characters := []rune(content)
	original := []rune(selection.Text())
	if len(original) == 0 || isCommonMarkSymbol(original[0]) || isCommonMarkSymbol(original[len(original)-1]) {
		return nil
	}
	leading, trailing := "", ""
	if !leftFlankingAsterisk(before, characters[0]) {
		leading = " "
	}
	if !rightFlankingAsterisk(characters[len(characters)-1], after) {
		trailing = " "
	}
	return md.String(leading + "**" + content + "**" + trailing)
}

func leftFlankingAsterisk(before, after rune) bool {
	return !unicode.IsSpace(after) &&
		(!isCommonMarkPunctuation(after) || unicode.IsSpace(before) || isCommonMarkPunctuation(before))
}

func rightFlankingAsterisk(before, after rune) bool {
	return !unicode.IsSpace(before) &&
		(!isCommonMarkPunctuation(before) || unicode.IsSpace(after) || isCommonMarkPunctuation(after))
}

func isCommonMarkPunctuation(character rune) bool {
	return unicode.IsPunct(character) || isCommonMarkSymbol(character)
}

func isCommonMarkSymbol(character rune) bool {
	return strings.ContainsRune("$+<=>^`|~", character)
}

func lastTextRune(node *html.Node) (rune, bool) {
	if node == nil || node.Type != html.TextNode || node.Data == "" {
		return 0, false
	}
	characters := []rune(node.Data)
	return characters[len(characters)-1], true
}

func firstTextRune(node *html.Node) (rune, bool) {
	if node == nil || node.Type != html.TextNode || node.Data == "" {
		return 0, false
	}
	return []rune(node.Data)[0], true
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
		return text.NormalizeSpaces(dom.StripTagsWithBlockBoundaries(content))
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
	words := strings.Fields(dom.StripTags(content))
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
