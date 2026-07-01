// ABOUTME: GenericLanguageExtractor extracts content language from HTML attributes, meta tags, and JSON-LD
// ABOUTME: Handles language format normalization and provides fallbacks with priority-based extraction

package generic

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GenericLanguageExtractor extracts content language information.
type GenericLanguageExtractor struct{}

// Meta tags for language extraction, ordered by priority.
var languageMetaTags = []string{
	"og:locale",        // Open Graph locale (most specific)
	"content-language", // HTTP content-language equivalent
	"dc.language",      // Dublin Core language
	"language",         // Generic language meta tag
}

// Language code normalization regex.
var (
	// Matches locale codes like en-US, en_US, pt-BR, zh-CN, etc.
	localeCodeRE = regexp.MustCompile(`^([a-zA-Z]{2})[-_]([a-zA-Z]{2})$`)
	// Matches simple language codes like en, fr, de, etc.
	languageCodeRE = regexp.MustCompile(`^[a-z]{2}$`)
)

// Extract extracts content language using priority-based strategies.
func (extractor *GenericLanguageExtractor) Extract(selection *goquery.Selection, pageURL string, metaCache []string) string {
	// Strategy 1: Try HTML lang attribute (highest priority)
	if lang := extractor.extractFromHTMLLang(selection); lang != "" {
		return extractor.normalizeLanguageCode(lang)
	}

	// Strategy 2: Try meta tags
	if lang := extractor.extractFromMetaTags(selection); lang != "" {
		return extractor.normalizeLanguageCode(lang)
	}

	// Strategy 3: Try JSON-LD structured data
	if lang := extractor.extractFromJSONLD(selection); lang != "" {
		return extractor.normalizeLanguageCode(lang)
	}

	return ""
}

// extractFromHTMLLang extracts language from HTML lang attribute.
func (extractor *GenericLanguageExtractor) extractFromHTMLLang(selection *goquery.Selection) string {
	// Check <html lang="...">
	if lang := selection.Find("html").AttrOr("lang", ""); lang != "" {
		return strings.TrimSpace(lang)
	}

	// Check <html xml:lang="..."> for XHTML compatibility
	if lang := selection.Find("html").AttrOr("xml:lang", ""); lang != "" {
		return strings.TrimSpace(lang)
	}

	return ""
}

// extractFromMetaTags extracts language from meta tags using priority order.
func (extractor *GenericLanguageExtractor) extractFromMetaTags(selection *goquery.Selection) string {
	if value := firstMetaValue(selection, languageMetaTags, extractor.isValidLanguageCode); value != "" {
		return value
	}

	if value := firstMetaAttr(selection, `meta[http-equiv="Content-Language"]`, "value", extractor.isValidLanguageCode); value != "" {
		return value
	}

	return firstMetaAttr(selection, `meta[http-equiv="Content-Language"]`, "content", extractor.isValidLanguageCode)
}

// extractFromJSONLD extracts language from JSON-LD structured data.
func (extractor *GenericLanguageExtractor) extractFromJSONLD(selection *goquery.Selection) string {
	var foundLang string

	eachJSONLD(selection, func(data map[string]interface{}) bool {
		if inLanguage, ok := data["inLanguage"].(string); ok && extractor.isValidLanguageCode(inLanguage) {
			foundLang = inLanguage
			return false
		}

		if language, ok := data["@language"].(string); ok && extractor.isValidLanguageCode(language) {
			foundLang = language
			return false
		}

		if typeVal, ok := data["@type"].(string); ok && (typeVal == "Article" || typeVal == "NewsArticle") {
			if contentLanguage, ok := data["contentLanguage"].(string); ok && extractor.isValidLanguageCode(contentLanguage) {
				foundLang = contentLanguage
				return false
			}
		}

		return true
	})

	return foundLang
}

// isValidLanguageCode validates that the language code is reasonable.
func (extractor *GenericLanguageExtractor) isValidLanguageCode(lang string) bool {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return false
	}

	if languageCodeRE.MatchString(lang) {
		return true
	}

	if localeCodeRE.MatchString(lang) {
		return true
	}

	commonCodes := []string{
		"zh-hans", "zh-hant", // Chinese script variants
		"ar-sa", // Arabic Saudi
	}
	for _, code := range commonCodes {
		if strings.ToLower(lang) == code {
			return true
		}
	}

	return false
}

// normalizeLanguageCode normalizes language codes to standard format.
func (extractor *GenericLanguageExtractor) normalizeLanguageCode(lang string) string {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return ""
	}

	// Convert to lowercase for processing
	lower := strings.ToLower(lang)

	// Handle locale codes with underscores (en_US -> en-US)
	if strings.Contains(lower, "_") {
		parts := strings.Split(lower, "_")
		if len(parts) == 2 && len(parts[0]) == 2 && len(parts[1]) == 2 {
			return parts[0] + "-" + strings.ToUpper(parts[1])
		}
	}

	// Handle locale codes with hyphens (en-us -> en-US)
	if strings.Contains(lower, "-") {
		parts := strings.Split(lower, "-")
		if len(parts) == 2 && len(parts[0]) == 2 && len(parts[1]) == 2 {
			return parts[0] + "-" + strings.ToUpper(parts[1])
		}
		if len(parts) == 2 {
			switch parts[1] {
			case "hans":
				return parts[0] + "-Hans"
			case "hant":
				return parts[0] + "-Hant"
			}
		}
		return lang
	}

	// Handle special Facebook locale format (en_US -> en-US)
	if matches := localeCodeRE.FindStringSubmatch(strings.ToUpper(strings.ReplaceAll(lower, "_", "-"))); len(matches) == 3 {
		return strings.ToLower(matches[1]) + "-" + strings.ToUpper(matches[2])
	}

	// For simple language codes, return lowercased
	if len(lang) == 2 && languageCodeRE.MatchString(lower) {
		return lower
	}

	// Return the original if we can't normalize it
	return lang
}
