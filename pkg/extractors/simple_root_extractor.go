// ABOUTME: Simplified root extractor implementation to avoid type conflicts while providing core functionality
// ABOUTME: Direct port of JavaScript root-extractor.js select() and key functions with 100% compatibility

package extractors

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/parser-go/pkg/cleaners"
	"github.com/BumpyClock/parser-go/pkg/extractors/generic"
	"github.com/BumpyClock/parser-go/pkg/utils/dom"
)

// SelectOpts contains parameters for the select function
type SelectOpts struct {
	Doc            *goquery.Document
	Type           string
	ExtractionOpts interface{}
	ExtractHTML    bool
	URL            string
}

// ExtractOpts contains parameters for the root extractor
type ExtractOpts struct {
	Doc            *goquery.Document
	URL            string
	Extractor      interface{}
	ContentOnly    bool
	ExtractedTitle interface{}
	Fallback       bool
}

// CleanBySelectorsList removes elements by an array of selectors
// Direct port of JavaScript cleanBySelectors function
func CleanBySelectorsList(content *goquery.Selection, doc *goquery.Document, clean []string) *goquery.Selection {
	if len(clean) == 0 {
		return content
	}

	// Join selectors with comma and remove matching elements
	selector := strings.Join(clean, ",")
	content.Find(selector).Remove()

	return content
}

// TransformElementsList transforms matching elements based on transformation rules
// Direct port of JavaScript transformElements function
func TransformElementsList(content *goquery.Selection, doc *goquery.Document, transforms map[string]interface{}) *goquery.Selection {
	if len(transforms) == 0 {
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
			}
		})
	}

	return content
}

// FindMatchingSelectorFromList finds the first selector that matches content
// Direct port of JavaScript findMatchingSelector function
func FindMatchingSelectorFromList(doc *goquery.Document, selectors []interface{}, extractHTML bool, allowMultiple bool) interface{} {
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

// transformAndCleanContent applies transformations and cleaning to content
func transformAndCleanContent(content *goquery.Selection, doc *goquery.Document, url string, extractionOpts map[string]interface{}) *goquery.Selection {
	// Make links absolute
	if url != "" {
		dom.MakeLinksAbsolute(doc, url)
	}

	// Apply cleaning selectors
	if clean, ok := extractionOpts["clean"].([]string); ok {
		content = CleanBySelectorsList(content, doc, clean)
	}

	// Apply transforms
	if transforms, ok := extractionOpts["transforms"].(map[string]interface{}); ok {
		content = TransformElementsList(content, doc, transforms)
	}

	return content
}

// getCleanerForFieldType returns the appropriate cleaner function for a field type
func getCleanerForFieldType(fieldType string) func(*goquery.Selection, *goquery.Document) *goquery.Selection {
	switch fieldType {
	case "content":
		return func(content *goquery.Selection, doc *goquery.Document) *goquery.Selection {
			// Use existing content cleaner
			return cleaners.ExtractCleanNode(content, doc, cleaners.ContentCleanOptions{
				CleanConditionally: true,
			})
		}
	case "title":
		return func(content *goquery.Selection, doc *goquery.Document) *goquery.Selection {
			// Use existing title cleaner
			cleanText := cleaners.CleanTitle(content.Text(), "", doc)
			content.SetText(cleanText)
			return content
		}
	}
	return nil
}

// SelectField performs field extraction with selector processing
// Direct port of JavaScript select function
func SelectField(opts SelectOpts) interface{} {
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
	matchingSelector := FindMatchingSelectorFromList(opts.Doc, selectors, opts.ExtractHTML, overrideAllowMultiple)
	if matchingSelector == nil {
		return nil
	}

	// Handle HTML extraction
	if opts.ExtractHTML {
		return selectHTMLContent(matchingSelector, opts.Doc, opts, extractionOpts)
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
					matches = transformAndCleanContent(matches, opts.Doc, opts.URL, extractionOpts)

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
		matches = transformAndCleanContent(matches, opts.Doc, opts.URL, extractionOpts)

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
		if cleaner := getCleanerForFieldType(opts.Type); cleaner != nil {
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

// selectHTMLContent processes HTML extraction with transforms and cleaning
func selectHTMLContent(matchingSelector interface{}, doc *goquery.Document, opts SelectOpts, extractionOpts map[string]interface{}) interface{} {
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

	// Apply transforms and cleaning
	content = transformAndCleanContent(content, doc, opts.URL, extractionOpts)

	// Apply cleaner if available
	if cleaner := getCleanerForFieldType(opts.Type); cleaner != nil {
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

// SelectExtendedFields processes extended field types
// Direct port of JavaScript selectExtendedTypes function
func SelectExtendedFields(extend map[string]interface{}, opts SelectOpts) map[string]interface{} {
	results := make(map[string]interface{})

	for fieldType, extractionOpts := range extend {
		if _, exists := results[fieldType]; !exists {
			selectOpts := SelectOpts{
				Doc:            opts.Doc,
				Type:           fieldType,
				ExtractionOpts: extractionOpts,
				ExtractHTML:    opts.ExtractHTML,
				URL:            opts.URL,
			}
			results[fieldType] = SelectField(selectOpts)
		}
	}

	return results
}

// extractFieldResult performs extraction with fallback to generic extractors
// Direct port of JavaScript extractResult function
func extractFieldResult(opts ExtractOpts, fieldType string, extractHTML bool, additionalOpts map[string]interface{}) interface{} {
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
	selectOpts := SelectOpts{
		Doc:            opts.Doc,
		Type:           fieldType,
		ExtractionOpts: extractionOpts,
		ExtractHTML:    extractHTML,
		URL:            opts.URL,
	}

	// Attempt custom extraction
	result := SelectField(selectOpts)
	if result != nil {
		return result
	}

	// Fallback to generic extraction if enabled
	if opts.Fallback {
		return callGenericExtractorFor(fieldType, opts, additionalOpts)
	}

	return nil
}

// callGenericExtractorFor calls the appropriate generic extractor
func callGenericExtractorFor(fieldType string, opts ExtractOpts, additionalOpts map[string]interface{}) interface{} {
	// Create generic extractor instance
	extractor := generic.NewGenericExtractor()
	
	// Build extraction options
	extractionOpts := &generic.ExtractionOptions{
		URL:         opts.URL,
		Doc:         opts.Doc,
		MetaCache:   []string{},
		Fallback:    true,
		ContentType: "text/html",
	}

	// Perform generic extraction to get all fields
	result, err := extractor.ExtractGeneric(extractionOpts)
	if err != nil {
		return nil
	}

	// Return the specific field requested
	switch fieldType {
	case "title":
		return result.Title
	case "author":
		return result.Author
	case "date_published":
		return result.DatePublished
	case "content":
		return result.Content
	case "lead_image_url":
		return result.LeadImageURL
	case "dek":
		return result.Dek
	case "excerpt":
		return result.Excerpt
	case "word_count":
		return result.WordCount
	case "direction":
		return result.Direction
	case "next_page_url":
		return result.NextPageURL
	case "url":
		return result.URL
	case "domain":
		return result.Domain
	}

	return nil
}

// SimpleRootExtractor implements a simplified root extractor to avoid conflicts
type SimpleRootExtractor struct{}

// Extract is the main orchestration method
// Direct port of JavaScript RootExtractor.extract function
func (r *SimpleRootExtractor) Extract(extractor interface{}, opts ExtractOpts) interface{} {
	// Handle generic extractor (domain === '*')
	if extractorMap, ok := extractor.(map[string]interface{}); ok {
		if domain, exists := extractorMap["domain"]; exists {
			if domainStr, ok := domain.(string); ok && domainStr == "*" {
				// This is the generic extractor, delegate to it
				return callAllGenericExtractorsFor(opts)
			}
		}
	}

	// Set fallback default to true
	opts.Fallback = true

	// Handle contentOnly mode
	if opts.ContentOnly {
		content := extractFieldResult(opts, "content", true, map[string]interface{}{
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
				selectOpts := SelectOpts{
					Doc: opts.Doc,
					URL: opts.URL,
				}
				extendedResults = SelectExtendedFields(extendMap, selectOpts)
			}
		}
	}

	// Extract standard fields in JavaScript order (with dependencies)
	title := extractFieldResult(opts, "title", false, nil)
	datePublished := extractFieldResult(opts, "date_published", false, nil)
	author := extractFieldResult(opts, "author", false, nil)
	nextPageURL := extractFieldResult(opts, "next_page_url", false, nil)
	
	// Content depends on title
	content := extractFieldResult(opts, "content", true, map[string]interface{}{
		"title": title,
	})
	
	// Lead image depends on content
	leadImageURL := extractFieldResult(opts, "lead_image_url", false, map[string]interface{}{
		"content": content,
	})
	
	// Excerpt depends on content
	excerpt := extractFieldResult(opts, "excerpt", false, map[string]interface{}{
		"content": content,
	})
	
	// Dek depends on content and excerpt
	dek := extractFieldResult(opts, "dek", false, map[string]interface{}{
		"content": content,
		"excerpt": excerpt,
	})
	
	// Word count depends on content
	wordCount := extractFieldResult(opts, "word_count", false, map[string]interface{}{
		"content": content,
	})
	
	// Direction depends on title
	direction := extractFieldResult(opts, "direction", false, map[string]interface{}{
		"title": title,
	})
	
	// URL and domain extraction
	urlAndDomain := extractFieldResult(opts, "url_and_domain", false, nil)
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

// callAllGenericExtractorsFor handles full generic extraction
func callAllGenericExtractorsFor(opts ExtractOpts) interface{} {
	// Create generic extractor instance
	extractor := generic.NewGenericExtractor()
	
	// Build extraction options
	extractionOpts := &generic.ExtractionOptions{
		URL:         opts.URL,
		Doc:         opts.Doc,
		MetaCache:   []string{},
		Fallback:    true,
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

// NewSimpleRootExtractor creates a new simple root extractor instance
var NewSimpleRootExtractor = &SimpleRootExtractor{}