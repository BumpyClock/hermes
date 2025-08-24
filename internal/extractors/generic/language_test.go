// ABOUTME: Comprehensive test suite for language extraction functionality
// ABOUTME: Tests HTML attributes, meta tags, JSON-LD parsing, and language code normalization

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/hermes/internal/resource"
)

func TestGenericLanguageExtractor_Extract(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		url       string
		metaCache []string
		expected  string
	}{
		{
			name: "extracts from html lang attribute",
			html: `<html lang="en-US">
				<head><title>Test</title></head>
			</html>`,
			url:       "https://example.com",
			metaCache: []string{},
			expected:  "en-US",
		},
		{
			name: "extracts from html xml:lang attribute",
			html: `<html xml:lang="fr-FR">
				<head><title>Test</title></head>
			</html>`,
			url:       "https://example.fr",
			metaCache: []string{},
			expected:  "fr-FR",
		},
		{
			name: "extracts from og:locale meta tag",
			html: `<html>
				<head>
					<meta property="og:locale" content="en_US" />
				</head>
			</html>`,
			url:       "https://example.com",
			metaCache: []string{},
			expected:  "en-US",
		},
		{
			name: "extracts from content-language meta name",
			html: `<html>
				<head>
					<meta name="content-language" content="es-ES" />
				</head>
			</html>`,
			url:       "https://example.es",
			metaCache: []string{},
			expected:  "es-ES",
		},
		{
			name: "extracts from content-language http-equiv",
			html: `<html>
				<head>
					<meta http-equiv="Content-Language" content="de-DE" />
				</head>
			</html>`,
			url:       "https://example.de",
			metaCache: []string{},
			expected:  "de-DE",
		},
		{
			name: "extracts from dc.language meta tag",
			html: `<html>
				<head>
					<meta name="dc.language" content="it" />
				</head>
			</html>`,
			url:       "https://example.it",
			metaCache: []string{},
			expected:  "it",
		},
		{
			name: "prefers html lang over meta tags",
			html: `<html lang="en-GB">
				<head>
					<meta property="og:locale" content="en_US" />
				</head>
			</html>`,
			url:       "https://example.com",
			metaCache: []string{},
			expected:  "en-GB",
		},
		{
			name: "prefers og:locale over other meta tags",
			html: `<html>
				<head>
					<meta property="og:locale" content="pt_BR" />
					<meta name="content-language" content="en" />
				</head>
			</html>`,
			url:       "https://example.com",
			metaCache: []string{},
			expected:  "pt-BR",
		},
		{
			name: "extracts from JSON-LD inLanguage",
			html: `<html>
				<head>
					<script type="application/ld+json">
					{
						"@context": "https://schema.org",
						"@type": "WebSite",
						"name": "Test Site",
						"inLanguage": "ja"
					}
					</script>
				</head>
			</html>`,
			url:       "https://example.jp",
			metaCache: []string{},
			expected:  "ja",
		},
		{
			name: "extracts from JSON-LD @language",
			html: `<html>
				<head>
					<script type="application/ld+json">
					{
						"@context": "https://schema.org",
						"@type": "Article",
						"@language": "ko"
					}
					</script>
				</head>
			</html>`,
			url:       "https://example.kr",
			metaCache: []string{},
			expected:  "ko",
		},
		{
			name: "extracts from JSON-LD Article contentLanguage",
			html: `<html>
				<head>
					<script type="application/ld+json">
					{
						"@context": "https://schema.org",
						"@type": "Article",
						"headline": "Test Article",
						"contentLanguage": "zh-CN"
					}
					</script>
				</head>
			</html>`,
			url:       "https://example.cn",
			metaCache: []string{},
			expected:  "zh-CN",
		},
		{
			name: "handles invalid JSON-LD gracefully",
			html: `<html>
				<head>
					<script type="application/ld+json">
					{ invalid json }
					</script>
					<meta property="og:locale" content="en_US" />
				</head>
			</html>`,
			url:       "https://example.com",
			metaCache: []string{},
			expected:  "en-US",
		},
		{
			name: "empty when no valid language data",
			html: `<html>
				<head>
					<meta name="other" content="not relevant" />
				</head>
			</html>`,
			url:       "https://example.com",
			metaCache: []string{},
			expected:  "",
		},
		{
			name: "normalizes underscore to hyphen",
			html: `<html>
				<head>
					<meta property="og:locale" content="zh_TW" />
				</head>
			</html>`,
			url:       "https://example.tw",
			metaCache: []string{},
			expected:  "zh-TW",
		},
		{
			name: "handles simple language codes",
			html: `<html lang="ru">
				<head><title>Test</title></head>
			</html>`,
			url:       "https://example.ru",
			metaCache: []string{},
			expected:  "ru",
		},
	}

	extractor := &GenericLanguageExtractor{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Apply same normalization as in real extraction pipeline
			doc = resource.NormalizeMetaTags(doc)

			result := extractor.Extract(doc.Selection, tt.url, tt.metaCache)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenericLanguageExtractor_IsValidLanguageCode(t *testing.T) {
	extractor := &GenericLanguageExtractor{}

	tests := []struct {
		input    string
		expected bool
	}{
		// Valid simple language codes
		{"en", true},
		{"fr", true},
		{"de", true},
		{"es", true},
		{"it", true},
		{"ja", true},
		{"ko", true},
		{"ru", true},
		{"ar", true},
		{"zh", true},

		// Valid locale codes
		{"en-US", true},
		{"en-GB", true},
		{"fr-FR", true},
		{"fr-CA", true},
		{"es-ES", true},
		{"es-MX", true},
		{"pt-BR", true},
		{"zh-CN", true},
		{"zh-TW", true},

		// Valid underscore variants
		{"en_US", true},
		{"pt_BR", true},
		{"zh_CN", true},

		// Valid special cases
		{"zh-hans", true},
		{"zh-hant", true},
		{"ar-sa", true},

		// Invalid codes
		{"", false},
		{"english", false},
		{"EN", false}, // uppercase simple codes
		{"e", false},  // too short
		{"eng", false}, // too long for simple
		{"en-", false}, // incomplete
		{"en-USA", false}, // region too long
		{"123", false}, // numeric
		{"en@US", false}, // wrong separator
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractor.isValidLanguageCode(tt.input)
			if result != tt.expected {
				t.Errorf("For input %q, expected %v, got %v", tt.input, tt.expected, result)
			}
		})
	}
}

func TestGenericLanguageExtractor_NormalizeLanguageCode(t *testing.T) {
	extractor := &GenericLanguageExtractor{}

	tests := []struct {
		input    string
		expected string
	}{
		// Simple language codes
		{"en", "en"},
		{"fr", "fr"},
		{"de", "de"},
		{"ES", "es"}, // Uppercase simple codes get normalized to lowercase

		// Locale codes with correct case
		{"en-US", "en-US"},
		{"fr-CA", "fr-CA"},
		{"pt-BR", "pt-BR"},

		// Locale codes needing case correction
		{"en-us", "en-US"},
		{"fr-ca", "fr-CA"},
		{"PT-br", "pt-BR"},

		// Underscore to hyphen conversion
		{"en_US", "en-US"},
		{"pt_BR", "pt-BR"},
		{"zh_CN", "zh-CN"},
		{"fr_CA", "fr-CA"},

		// Mixed case underscore conversion
		{"EN_us", "en-US"},
		{"pt_br", "pt-BR"},

		// Special Chinese variants
		{"zh-hans", "zh-Hans"},
		{"zh-hant", "zh-Hant"},
		{"ZH-hans", "zh-Hans"},

		// Complex cases
		{"en-US-x-custom", "en-US-x-custom"}, // Multiple parts - leave complex tags as-is

		// Empty/invalid
		{"", ""},
		{"   ", ""},

		// Cases that can't be normalized
		{"invalid", "invalid"},
		{"toolong", "toolong"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractor.normalizeLanguageCode(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenericLanguageExtractor_ExtractFromHTMLLang(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "html lang attribute",
			html:     `<html lang="en-US"><head></head></html>`,
			expected: "en-US",
		},
		{
			name:     "html xml:lang attribute",
			html:     `<html xml:lang="fr-FR"><head></head></html>`,
			expected: "fr-FR",
		},
		{
			name:     "both lang and xml:lang (lang takes priority)",
			html:     `<html lang="en-US" xml:lang="en-GB"><head></head></html>`,
			expected: "en-US",
		},
		{
			name:     "no lang attributes",
			html:     `<html><head></head></html>`,
			expected: "",
		},
		{
			name:     "empty lang attribute",
			html:     `<html lang=""><head></head></html>`,
			expected: "",
		},
	}

	extractor := &GenericLanguageExtractor{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Apply same normalization as in real extraction pipeline
			doc = resource.NormalizeMetaTags(doc)

			result := extractor.extractFromHTMLLang(doc.Selection)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenericLanguageExtractor_GetLanguageName(t *testing.T) {
	extractor := &GenericLanguageExtractor{}

	tests := []struct {
		code     string
		expected string
	}{
		{"en", "English"},
		{"en-US", "English (United States)"},
		{"en-GB", "English (United Kingdom)"},
		{"es", "Spanish"},
		{"es-ES", "Spanish (Spain)"},
		{"es-MX", "Spanish (Mexico)"},
		{"fr", "French"},
		{"fr-FR", "French (France)"},
		{"fr-CA", "French (Canada)"},
		{"pt-BR", "Portuguese (Brazil)"},
		{"zh", "Chinese"},
		{"zh-CN", "Chinese (Simplified)"},
		{"zh-TW", "Chinese (Traditional)"},
		{"unknown", "unknown"}, // Unknown code returns itself
		{"", ""},               // Empty code returns empty
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := extractor.getLanguageName(tt.code)
			if result != tt.expected {
				t.Errorf("For code %q, expected %q, got %q", tt.code, tt.expected, result)
			}
		})
	}
}