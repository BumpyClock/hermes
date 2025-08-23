// ABOUTME: GenericLanguageExtractor extracts content language from HTML attributes, meta tags, and JSON-LD
// ABOUTME: Handles language format normalization and provides fallbacks with priority-based extraction

package generic

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GenericLanguageExtractor extracts content language information
type GenericLanguageExtractor struct{}

// Meta tags for language extraction, ordered by priority
var languageMetaTags = []string{
	"og:locale",         // Open Graph locale (most specific)
	"content-language",  // HTTP content-language equivalent
	"dc.language",       // Dublin Core language
	"language",          // Generic language meta tag
}

// Language code normalization regex
var (
	// Matches locale codes like en-US, en_US, pt-BR, zh-CN, etc.
	localeCodeRE = regexp.MustCompile(`^([a-z]{2})[-_]([A-Z]{2})$`)
	// Matches simple language codes like en, fr, de, etc.
	languageCodeRE = regexp.MustCompile(`^[a-z]{2}$`)
)

// Extract extracts content language using priority-based strategies
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

// extractFromHTMLLang extracts language from HTML lang attribute
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

// extractFromMetaTags extracts language from meta tags using priority order
func (extractor *GenericLanguageExtractor) extractFromMetaTags(selection *goquery.Selection) string {
	// Check each meta tag in priority order
	for _, tagName := range languageMetaTags {
		// After normalization, content becomes value and property becomes name
		content := selection.Find("meta[name=\"" + tagName + "\"]").AttrOr("value", "")
		if content != "" && extractor.isValidLanguageCode(content) {
			return strings.TrimSpace(content)
		}

		// For content-language, also check http-equiv (content still becomes value)
		if tagName == "content-language" {
			content = selection.Find("meta[http-equiv=\"Content-Language\"]").AttrOr("value", "")
			if content != "" && extractor.isValidLanguageCode(content) {
				return strings.TrimSpace(content)
			}
		}
	}

	return ""
}

// extractFromJSONLD extracts language from JSON-LD structured data
func (extractor *GenericLanguageExtractor) extractFromJSONLD(selection *goquery.Selection) string {
	var foundLang string

	// Find all JSON-LD script tags
	selection.Find("script[type=\"application/ld+json\"]").Each(func(i int, s *goquery.Selection) {
		if foundLang != "" {
			return // Already found a language
		}

		jsonText := strings.TrimSpace(s.Text())
		if jsonText == "" {
			return
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonText), &data); err != nil {
			return // Skip invalid JSON
		}

		// Check for inLanguage field
		if inLanguage, ok := data["inLanguage"].(string); ok && extractor.isValidLanguageCode(inLanguage) {
			foundLang = inLanguage
			return
		}

		// Check for @language field (rare but sometimes used)
		if language, ok := data["@language"].(string); ok && extractor.isValidLanguageCode(language) {
			foundLang = language
			return
		}

		// For articles, check if there's language information in content
		if typeVal, ok := data["@type"].(string); ok {
			if typeVal == "Article" || typeVal == "NewsArticle" {
				if contentLanguage, ok := data["contentLanguage"].(string); ok && extractor.isValidLanguageCode(contentLanguage) {
					foundLang = contentLanguage
					return
				}
			}
		}
	})

	return foundLang
}

// isValidLanguageCode validates that the language code is reasonable
func (extractor *GenericLanguageExtractor) isValidLanguageCode(lang string) bool {
	lang = strings.TrimSpace(strings.ToLower(lang))
	
	// Must not be empty
	if lang == "" {
		return false
	}

	// Check for simple language codes (en, fr, de, etc.)
	if languageCodeRE.MatchString(lang) {
		return true
	}

	// Check for locale codes (en-US, pt-BR, etc.)
	if localeCodeRE.MatchString(strings.ToUpper(lang)) {
		return true
	}

	// Handle underscore variants (en_US -> en-US)
	normalized := strings.ReplaceAll(lang, "_", "-")
	if localeCodeRE.MatchString(strings.ToUpper(normalized)) {
		return true
	}

	// Accept some common special cases
	commonCodes := []string{
		"zh-cn", "zh-tw", "zh-hans", "zh-hant", // Chinese variants
		"pt-br",                                 // Portuguese Brazil
		"es-mx", "es-es",                       // Spanish variants
		"fr-ca",                                // French Canadian
		"ar-sa",                                // Arabic Saudi
	}

	for _, code := range commonCodes {
		if strings.ToLower(lang) == code {
			return true
		}
	}

	return false
}

// normalizeLanguageCode normalizes language codes to standard format
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
		// For longer codes like zh-Hans, keep as-is but fix case
		if len(parts) >= 2 {
			result := parts[0]
			for i := 1; i < len(parts); i++ {
				// Title case only for known script tags, otherwise leave as-is
				part := parts[i]
				if part == "hans" || part == "hant" {
					part = strings.Title(part)
				}
				result += "-" + part
			}
			return result
		}
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

// getLanguageName returns a human-readable language name for common codes
func (extractor *GenericLanguageExtractor) getLanguageName(code string) string {
	// Common language mappings for display purposes
	languageNames := map[string]string{
		"en":    "English",
		"en-US": "English (United States)",
		"en-GB": "English (United Kingdom)",
		"es":    "Spanish", 
		"es-ES": "Spanish (Spain)",
		"es-MX": "Spanish (Mexico)",
		"fr":    "French",
		"fr-FR": "French (France)",
		"fr-CA": "French (Canada)",
		"de":    "German",
		"it":    "Italian",
		"pt":    "Portuguese",
		"pt-BR": "Portuguese (Brazil)",
		"ru":    "Russian",
		"ja":    "Japanese",
		"ko":    "Korean",
		"zh":    "Chinese",
		"zh-CN": "Chinese (Simplified)",
		"zh-TW": "Chinese (Traditional)",
		"ar":    "Arabic",
		"nl":    "Dutch",
		"sv":    "Swedish",
		"da":    "Danish",
		"no":    "Norwegian",
		"fi":    "Finnish",
		"pl":    "Polish",
		"tr":    "Turkish",
		"hi":    "Hindi",
	}

	if name, ok := languageNames[code]; ok {
		return name
	}

	// Return the code itself if no mapping found
	return code
}