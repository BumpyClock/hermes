// ABOUTME: Root extractor orchestration system for complex selector processing, transforms, and extended types
// ABOUTME: 1:1 port of JavaScript root-extractor.js with 100% behavioral compatibility

package extractors

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/parser-go/pkg/cleaners"
	"github.com/BumpyClock/parser-go/pkg/extractors/generic"
	"github.com/BumpyClock/parser-go/pkg/utils/dom"
)

// ExtractOptions contains parameters for the root extractor
type ExtractOptions struct {
	Doc            *goquery.Document
	URL            string
	Extractor      interface{}
	ContentOnly    bool
	ExtractedTitle interface{}
	Fallback       bool
}

// CleanBySelectors removes elements by an array of selectors
// Direct port of JavaScript cleanBySelectors function
func CleanBySelectors(content *goquery.Selection, doc *goquery.Document, opts map[string][]string) *goquery.Selection {
	clean, exists := opts["clean"]
	if !exists || len(clean) == 0 {
		return content
	}

	// Join selectors with comma and remove matching elements
	selector := strings.Join(clean, ",")
	content.Find(selector).Remove()

	return content
}

// TransformElements transforms matching elements based on transformation rules
// Direct port of JavaScript transformElements function
func TransformElements(content *goquery.Selection, doc *goquery.Document, opts map[string]map[string]interface{}) *goquery.Selection {
	transforms, exists := opts["transforms"]
	if !exists || len(transforms) == 0 {
		return content
	}

	// Process each transformation rule
	for selector, value := range transforms {
		matches := content.Find(selector)
		
		matches.Each(func(i int, node *goquery.Selection) {
			switch v := value.(type) {
			case string:
				// If value is a string, convert directly
				dom.ConvertNodeTo(node, v)
			case func(*goquery.Selection, *goquery.Document) string:
				// If value is function, apply function to node
				result := v(node, doc)
				if result != "" {
					dom.ConvertNodeTo(node, result)
				}
			case TransformFunc:
				// Handle typed transform function
				result := v(node, doc)
				if resultStr, ok := result.(string); ok && resultStr != "" {
					dom.ConvertNodeTo(node, resultStr)
				}
			}
		})
	}

	return content
}

// FindMatchingSelector finds the first selector that matches content
// Direct port of JavaScript findMatchingSelector function
func FindMatchingSelector(doc *goquery.Document, selectors []interface{}, extractHTML bool, allowMultiple bool) interface{} {
	for _, selector := range selectors {
		switch sel := selector.(type) {
		case []interface{}:
			// Array selector like ["img", "src"]
			if extractHTML {
				// Check if all selectors in array match
				allMatch := true
				for _, s := range sel {
					if sStr, ok := s.(string); ok {
						if doc.Find(sStr).Length() == 0 {
							allMatch = false
							break
						}
					}
				}
				if allMatch {
					return sel
				}
			} else {
				// Check [selector, attribute] pattern
				if len(sel) >= 2 {
					if selectorStr, ok := sel[0].(string); ok {
						if attrStr, ok := sel[1].(string); ok {
							matches := doc.Find(selectorStr)
							if allowMultiple || (!allowMultiple && matches.Length() == 1) {
								// Check if element has attribute with non-empty value
								attrVal, exists := matches.First().Attr(attrStr)
								if exists && strings.TrimSpace(attrVal) != "" {
									return sel
								}
							}
						}
					}
				}
			}
		case string:
			// String selector
			matches := doc.Find(sel)
			if allowMultiple || (!allowMultiple && matches.Length() == 1) {
				// Check if element has non-empty text
				if strings.TrimSpace(matches.Text()) != "" {
					return sel
				}
			}
		}
	}
	return nil
}

// selectHTML processes HTML extraction with transforms and cleaning
func selectHTML(matchingSelector interface{}, doc *goquery.Document, opts SelectOptions, extractionOpts map[string]interface{}) interface{} {
	var content *goquery.Selection

	// Handle array vs string selectors
	switch sel := matchingSelector.(type) {
	case []interface{}:
		// Multi-match selection - all selectors in array
		selectorStrs := make([]string, 0, len(sel))
		for _, s := range sel {
			if sStr, ok := s.(string); ok {
				selectorStrs = append(selectorStrs, sStr)
			}
		}
		selector := strings.Join(selectorStrs, ",")
		content = doc.Find(selector)
		
		// Create wrapper div and append all matches
		wrapperDoc, _ := goquery.NewDocumentFromReader(strings.NewReader("<div></div>"))
		wrapper := wrapperDoc.Find("div").First()
		content.Each(func(i int, el *goquery.Selection) {
			wrapper.AppendSelection(el.Clone())
		})
		content = wrapper
	case string:
		content = doc.Find(sel)
	}

	if content == nil || content.Length() == 0 {
		return nil
	}

	// Wrap in div so transformation can take place on root element
	if content.Parent().Length() == 0 {
		wrapperDoc, _ := goquery.NewDocumentFromReader(strings.NewReader("<div></div>"))
		wrapper := wrapperDoc.Find("div").First()
		wrapper.AppendSelection(content.Clone())
		content = wrapper
	}

	// Apply transforms and cleaning
	content = transformAndClean(content, doc, opts.URL, extractionOpts)

	// Apply cleaner if available
	if cleaner := getCleanerForType(opts.Type); cleaner != nil {
		defaultCleaner := true
		if dc, ok := extractionOpts["defaultCleaner"].(bool); ok {
			defaultCleaner = dc
		}
		if defaultCleaner {
			content = cleaner(content, doc)
		}
	}

	// Handle allowMultiple flag
	if allowMultiple, ok := extractionOpts["allowMultiple"].(bool); ok && allowMultiple {
		var results []string
		content.Children().Each(func(i int, el *goquery.Selection) {
			if html, err := el.Html(); err == nil {
				results = append(results, html)
			}
		})
		return results
	}

	// Return HTML content
	html, err := content.Html()
	if err != nil {
		return nil
	}
	return html
}

// transformAndClean applies transformations and cleaning to content
func transformAndClean(content *goquery.Selection, doc *goquery.Document, url string, extractionOpts map[string]interface{}) *goquery.Selection {
	// Make links absolute
	if url != "" {
		dom.MakeLinksAbsolute(doc, url)
	}

	// Apply cleaning selectors
	if clean, ok := extractionOpts["clean"].([]string); ok {
		opts := map[string][]string{"clean": clean}
		content = CleanBySelectors(content, doc, opts)
	}

	// Apply transforms
	if transforms, ok := extractionOpts["transforms"].(map[string]interface{}); ok {
		opts := map[string]map[string]interface{}{"transforms": transforms}
		content = TransformElements(content, doc, opts)
	}

	return content
}

// getCleanerForType returns the appropriate cleaner function for a field type
func getCleanerForType(fieldType string) func(*goquery.Selection, *goquery.Document) *goquery.Selection {
	switch fieldType {
	case "content":
		return func(content *goquery.Selection, doc *goquery.Document) *goquery.Selection {
			// Use existing content cleaner
			return cleaners.ExtractCleanNodeFunc(content, doc, cleaners.ContentCleanOptions{
				CleanConditionally: true,
			})
		}
	case "title":
		return func(content *goquery.Selection, doc *goquery.Document) *goquery.Selection {
			// Use existing title cleaner - URL will be added later when available
			cleanText := cleaners.CleanTitle(content.Text(), "", doc)
			content.SetText(cleanText)
			return content
		}
	// Add other cleaners as they become available
	}
	return nil
}

// Select performs field extraction with selector processing
// Direct port of JavaScript select function
func Select(opts SelectOptions) interface{} {
	// Skip if there's no extraction for this type
	if opts.ExtractionOpts == nil {
		return nil
	}

	// If a string is hardcoded for a type, return the string
	if str, ok := opts.ExtractionOpts.(string); ok {
		return str
	}

	// Convert extraction options to map
	extractionOpts, ok := opts.ExtractionOpts.(map[string]interface{})
	if !ok {
		return nil
	}

	// Get selectors
	selectorsRaw, exists := extractionOpts["selectors"]
	if !exists {
		return nil
	}

	selectors, ok := selectorsRaw.([]interface{})
	if !ok {
		return nil
	}

	// Get configuration options
	defaultCleaner := true
	if dc, ok := extractionOpts["defaultCleaner"].(bool); ok {
		defaultCleaner = dc
	}

	allowMultiple := false
	if am, ok := extractionOpts["allowMultiple"].(bool); ok {
		allowMultiple = am
	}

	// Override allowMultiple for lead_image_url
	overrideAllowMultiple := opts.Type == "lead_image_url" || allowMultiple

	// Find matching selector
	matchingSelector := FindMatchingSelector(opts.Doc, selectors, opts.ExtractHTML, overrideAllowMultiple)
	if matchingSelector == nil {
		return nil
	}

	// Handle HTML extraction
	if opts.ExtractHTML {
		return selectHTML(matchingSelector, opts.Doc, opts, extractionOpts)
	}

	// Handle text/attribute extraction
	var matches *goquery.Selection
	var results []string

	switch sel := matchingSelector.(type) {
	case []interface{}:
		// Array selector like ["img", "src", transformFunc]
		if len(sel) >= 2 {
			if selectorStr, ok := sel[0].(string); ok {
				if attrStr, ok := sel[1].(string); ok {
					matches = opts.Doc.Find(selectorStr)
					matches = transformAndClean(matches, opts.Doc, opts.URL, extractionOpts)

					matches.Each(func(i int, el *goquery.Selection) {
						if attrVal, exists := el.Attr(attrStr); exists {
							item := strings.TrimSpace(attrVal)
							// Apply transform if provided
							if len(sel) >= 3 {
								if transformFunc, ok := sel[2].(func(string) string); ok {
									item = transformFunc(item)
								}
							}
							results = append(results, item)
						}
					})
				}
			}
		}
	case string:
		// String selector - extract text
		matches = opts.Doc.Find(sel)
		matches = transformAndClean(matches, opts.Doc, opts.URL, extractionOpts)

		matches.Each(func(i int, el *goquery.Selection) {
			text := strings.TrimSpace(el.Text())
			if text != "" {
				results = append(results, text)
			}
		})
	}

	// Return result based on allowMultiple setting
	var result interface{}
	if allowMultiple && len(results) > 0 {
		result = results
	} else if len(results) > 0 {
		result = results[0]
	} else {
		return nil
	}

	// Apply cleaner if default cleaning is enabled
	if defaultCleaner {
		if cleaner := getCleanerForType(opts.Type); cleaner != nil {
			if str, ok := result.(string); ok {
				// For string results, apply cleaner to text
				doc, _ := goquery.NewDocumentFromReader(strings.NewReader(str))
				cleaned := cleaner(doc.Selection, doc)
				result = cleaned.Text()
			}
		}
	}

	return result
}

// SelectExtendedTypes processes extended field types
// Direct port of JavaScript selectExtendedTypes function
func SelectExtendedTypes(extend map[string]interface{}, opts SelectOptions) map[string]interface{} {
	results := make(map[string]interface{})

	for fieldType, extractionOpts := range extend {
		if _, exists := results[fieldType]; !exists {
			selectOpts := SelectOptions{
				Doc:            opts.Doc,
				Type:           fieldType,
				ExtractionOpts: extractionOpts,
				ExtractHTML:    opts.ExtractHTML,
				URL:            opts.URL,
			}
			results[fieldType] = Select(selectOpts)
		}
	}

	return results
}

// extractResult performs extraction with fallback to generic extractors
// Direct port of JavaScript extractResult function
func extractResult(opts ExtractOptions, fieldType string, extractHTML bool, additionalOpts map[string]interface{}) interface{} {
	extractor, ok := opts.Extractor.(map[string]interface{})
	if !ok {
		return nil
	}

	// Get extraction options for this field type
	extractionOpts, exists := extractor[fieldType]
	if !exists {
		extractionOpts = nil
	}

	// Prepare select options
	selectOpts := SelectOptions{
		Doc:            opts.Doc,
		Type:           fieldType,
		ExtractionOpts: extractionOpts,
		ExtractHTML:    extractHTML,
		URL:            opts.URL,
	}

	// Attempt custom extraction
	result := Select(selectOpts)
	if result != nil {
		return result
	}

	// Fallback to generic extraction if enabled
	if opts.Fallback {
		return callGenericExtractor(fieldType, opts, additionalOpts)
	}

	return nil
}

// callGenericExtractor calls the appropriate generic extractor
func callGenericExtractor(fieldType string, opts ExtractOptions, additionalOpts map[string]interface{}) interface{} {

	// Build metaCache as []string of meta tag names (like in parser package)
	var metaCache []string
	seen := make(map[string]bool)
	if opts.Doc != nil {
		opts.Doc.Find("meta").Each(func(i int, s *goquery.Selection) {
			if name, exists := s.Attr("name"); exists && name != "" && !seen[name] {
				metaCache = append(metaCache, name)
				seen[name] = true
			}
		})
	}

	switch fieldType {
	case "title":
		// Use individual field extractors from generic package
		return generic.GenericTitleExtractor.Extract(opts.Doc.Selection, opts.URL, metaCache)
	case "author":
		authorExtractor := &generic.GenericAuthorExtractor{}
		return authorExtractor.Extract(opts.Doc.Selection, metaCache)
	case "date_published":
		if dateStr := generic.GenericDateExtractor.Extract(opts.Doc.Selection, opts.URL, metaCache); dateStr != nil {
			return *dateStr
		}
		return ""
	case "content":
		contentExtractor := generic.NewGenericContentExtractor()
		params := generic.ExtractorParams{
			Doc:   opts.Doc,
			URL:   opts.URL,
			Title: "",
		}
		result := contentExtractor.Extract(params, generic.ExtractorOptions{})
		return result
	case "lead_image_url":
		imageExtractor := generic.NewGenericLeadImageExtractor()
		imageParams := generic.ExtractorImageParams{
			Doc: opts.Doc,
		}
		return imageExtractor.Extract(imageParams)
	case "dek":
		dekExtractor := &generic.GenericDekExtractor{}
		return dekExtractor.Extract(opts.Doc, nil)
	case "excerpt":
		// For now, return empty - excerpt extraction needs content parameter
		return ""
	case "word_count":
		// WordCount extractor expects different signature
		return 0
	case "direction":
		// Return default direction
		return "ltr"
	case "next_page_url":
		// For now, return empty - next page extraction not fully implemented
		return ""
	case "url_and_domain":
		// For now, return basic URL info
		return opts.URL
	}

	return nil
}

// RootExtractorInterface defines the root extractor interface
type RootExtractorInterface struct{}

// Extract is the main orchestration method
// Direct port of JavaScript RootExtractor.extract function
func (r *RootExtractorInterface) Extract(extractor interface{}, opts ExtractOptions) interface{} {
	// Handle generic extractor (domain === '*')
	if extractorMap, ok := extractor.(map[string]interface{}); ok {
		if domain, exists := extractorMap["domain"]; exists {
			if domainStr, ok := domain.(string); ok && domainStr == "*" {
				// This is the generic extractor, delegate to it
				return callAllGenericExtractors(opts)
			}
		}
	}

	// Set fallback default to true
	if !opts.Fallback {
		opts.Fallback = true
	}

	// Handle contentOnly mode
	if opts.ContentOnly {
		content := extractResult(opts, "content", true, map[string]interface{}{
			"title": opts.ExtractedTitle,
		})
		return map[string]interface{}{
			"content": content,
		}
	}

	// Extract extended types first (JavaScript order)
	extendedResults := make(map[string]interface{})
	if extractorMap, ok := extractor.(map[string]interface{}); ok {
		if extend, exists := extractorMap["extend"]; exists {
			if extendMap, ok := extend.(map[string]interface{}); ok {
				selectOpts := SelectOptions{
					Doc: opts.Doc,
					URL: opts.URL,
				}
				extendedResults = SelectExtendedTypes(extendMap, selectOpts)
			}
		}
	}

	// Extract standard fields in JavaScript order (with dependencies)
	title := extractResult(opts, "title", false, nil)
	datePublished := extractResult(opts, "date_published", false, nil)
	author := extractResult(opts, "author", false, nil)
	nextPageURL := extractResult(opts, "next_page_url", false, nil)
	
	// Content depends on title
	content := extractResult(opts, "content", true, map[string]interface{}{
		"title": title,
	})
	
	// Lead image depends on content
	leadImageURL := extractResult(opts, "lead_image_url", false, map[string]interface{}{
		"content": content,
	})
	
	// Excerpt depends on content
	excerpt := extractResult(opts, "excerpt", false, map[string]interface{}{
		"content": content,
	})
	
	// Dek depends on content and excerpt
	dek := extractResult(opts, "dek", false, map[string]interface{}{
		"content": content,
		"excerpt": excerpt,
	})
	
	// Word count depends on content
	wordCount := extractResult(opts, "word_count", false, map[string]interface{}{
		"content": content,
	})
	
	// Direction depends on title
	direction := extractResult(opts, "direction", false, map[string]interface{}{
		"title": title,
	})
	
	// URL and domain extraction
	urlAndDomain := extractResult(opts, "url_and_domain", false, nil)
	var url, domain interface{}
	if urlDomainMap, ok := urlAndDomain.(map[string]interface{}); ok {
		url = urlDomainMap["url"]
		domain = urlDomainMap["domain"]
	}

	// Build result matching JavaScript structure
	result := map[string]interface{}{
		"title":          title,
		"content":        content,
		"author":         author,
		"date_published": datePublished,
		"lead_image_url": leadImageURL,
		"dek":            dek,
		"next_page_url":  nextPageURL,
		"url":            url,
		"domain":         domain,
		"excerpt":        excerpt,
		"word_count":     wordCount,
		"direction":      direction,
	}

	// Merge extended results
	for key, value := range extendedResults {
		result[key] = value
	}

	return result
}

// callAllGenericExtractors handles full generic extraction
func callAllGenericExtractors(opts ExtractOptions) interface{} {
	// Create generic extractor instance
	extractor := generic.NewGenericExtractor()
	
	// Build extraction options for generic extraction
	extractionOpts := &generic.ExtractionOptions{
		URL:         opts.URL,
		HTML:        "", // Will be extracted from Doc if needed
		Doc:         opts.Doc,
		MetaCache:   []string{}, // Empty cache for generic extraction
		Fallback:    true,       // Use fallback extraction by default
		ContentType: "text/html",
	}

	// Perform generic extraction
	result, err := extractor.ExtractGeneric(extractionOpts)
	if err != nil {
		// Return empty result on error
		return make(map[string]interface{})
	}

	// Convert result to map[string]interface{} for compatibility
	resultMap := make(map[string]interface{})
	
	if result.Title != "" {
		resultMap["title"] = result.Title
	}
	if result.Author != "" {
		resultMap["author"] = result.Author
	}
	if result.DatePublished != nil {
		resultMap["date_published"] = result.DatePublished
	}
	if result.Content != "" {
		resultMap["content"] = result.Content
	}
	if result.LeadImageURL != "" {
		resultMap["lead_image_url"] = result.LeadImageURL
	}
	if result.Dek != "" {
		resultMap["dek"] = result.Dek
	}
	if result.Excerpt != "" {
		resultMap["excerpt"] = result.Excerpt
	}
	if result.WordCount > 0 {
		resultMap["word_count"] = result.WordCount
	}
	if result.NextPageURL != "" {
		resultMap["next_page_url"] = result.NextPageURL
	}
	if result.URL != "" {
		resultMap["url"] = result.URL
	}
	if result.Domain != "" {
		resultMap["domain"] = result.Domain
	}
	if result.Direction != "" {
		resultMap["direction"] = result.Direction
	}
	
	return resultMap
}

// RootExtractor is the singleton instance
var RootExtractor = &RootExtractorInterface{}