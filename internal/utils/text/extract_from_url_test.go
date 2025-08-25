package text

import (
	"regexp"
	"testing"
)

func TestExtractFromURL(t *testing.T) {
	// Test case 1: Extract date from URL (matching JavaScript test)
	t.Run("extracts datePublished from url", func(t *testing.T) {
		url := "https://example.com/2012/08/01/this-is-good"
		regexList := []*regexp.Regexp{
			regexp.MustCompile(`/(20\d{2}/\d{2}/\d{2})/`),
		}
		
		result, found := ExtractFromURL(url, regexList)
		if !found {
			t.Fatal("Expected to find a match, but got false")
		}
		if result != "2012/08/01" {
			t.Errorf("Expected '2012/08/01', got '%s'", result)
		}
	})

	// Test case 2: No match found (matching JavaScript test)
	t.Run("returns empty string and false if nothing found", func(t *testing.T) {
		url := "https://example.com/this-is-good"
		regexList := []*regexp.Regexp{
			regexp.MustCompile(`/(20\d{2}/\d{2}/\d{2})/`),
		}
		
		result, found := ExtractFromURL(url, regexList)
		if found {
			t.Fatal("Expected no match, but got true")
		}
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})

	// Test case 3: Multiple regex patterns - first match wins
	t.Run("returns first matching pattern", func(t *testing.T) {
		url := "https://example.com/2012/08/01/article-2012-08-01-title"
		regexList := []*regexp.Regexp{
			regexp.MustCompile(`/(20\d{2}/\d{2}/\d{2})/`),                    // Matches 2012/08/01
			regexp.MustCompile(`/article-(20\d{2}-\d{2}-\d{2})-/`),           // Matches 2012-08-01
		}
		
		result, found := ExtractFromURL(url, regexList)
		if !found {
			t.Fatal("Expected to find a match, but got false")
		}
		if result != "2012/08/01" {
			t.Errorf("Expected '2012/08/01' (first pattern), got '%s'", result)
		}
	})

	// Test case 4: Multiple regex patterns - second pattern matches
	t.Run("returns second pattern when first doesn't match", func(t *testing.T) {
		url := "https://example.com/article-2012-08-01-title"
		regexList := []*regexp.Regexp{
			regexp.MustCompile(`/(20\d{2}/\d{2}/\d{2})/`),                    // No match
			regexp.MustCompile(`article-(20\d{2}-\d{2}-\d{2})-`),             // Matches 2012-08-01
		}
		
		result, found := ExtractFromURL(url, regexList)
		if !found {
			t.Fatal("Expected to find a match, but got false")
		}
		if result != "2012-08-01" {
			t.Errorf("Expected '2012-08-01', got '%s'", result)
		}
	})

	// Test case 5: Real-world date patterns (from JavaScript constants)
	t.Run("extracts dates using real-world patterns", func(t *testing.T) {
		testCases := []struct {
			name     string
			url      string
			expected string
		}{
			{
				name:     "YYYY/MM/DD format",
				url:      "https://news.com/2023/12/25/christmas-story",
				expected: "2023/12/25",
			},
			{
				name:     "YYYY-MM-DD format",
				url:      "https://blog.com/posts/2023-12-25-holiday",
				expected: "2023-12-25",
			},
			{
				name:     "YYYY/MMM/DD format",
				url:      "https://site.com/2023/dec/25/article",
				expected: "2023/dec/25",
			},
		}

		// Patterns from JavaScript DATE_PUBLISHED_URL_RES
		regexList := []*regexp.Regexp{
			regexp.MustCompile(`(?i)/(20\d{2}/\d{2}/\d{2})/`),
			regexp.MustCompile(`(?i)(20\d{2}-[01]\d-[0-3]\d)`),
			regexp.MustCompile(`(?i)/(20\d{2}/(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)/[0-3]\d)/`),
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, found := ExtractFromURL(tc.url, regexList)
				if !found {
					t.Fatal("Expected to find a match, but got false")
				}
				if result != tc.expected {
					t.Errorf("Expected '%s', got '%s'", tc.expected, result)
				}
			})
		}
	})

	// Test case 6: Empty inputs
	t.Run("handles empty inputs", func(t *testing.T) {
		// Empty URL
		result, found := ExtractFromURL("", []*regexp.Regexp{regexp.MustCompile(`/(20\d{2})/`)})
		if found {
			t.Error("Expected no match for empty URL")
		}
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}

		// Empty regex list
		result, found = ExtractFromURL("https://example.com/2023/article", []*regexp.Regexp{})
		if found {
			t.Error("Expected no match for empty regex list")
		}
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}

		// Nil regex list
		result, found = ExtractFromURL("https://example.com/2023/article", nil)
		if found {
			t.Error("Expected no match for nil regex list")
		}
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})

	// Test case 7: Regex without capture groups
	t.Run("handles regex without capture groups", func(t *testing.T) {
		url := "https://example.com/2023/article"
		regexList := []*regexp.Regexp{
			regexp.MustCompile(`20\d{2}`), // No capture groups
		}
		
		result, found := ExtractFromURL(url, regexList)
		if found {
			t.Error("Expected no match when regex has no capture groups")
		}
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})

	// Test case 8: Case insensitive matching
	t.Run("case insensitive matching", func(t *testing.T) {
		url := "https://example.com/2023/DEC/25/article"
		regexList := []*regexp.Regexp{
			regexp.MustCompile(`(?i)/(20\d{2}/(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)/\d{2})/`),
		}
		
		result, found := ExtractFromURL(url, regexList)
		if !found {
			t.Fatal("Expected to find a match, but got false")
		}
		if result != "2023/DEC/25" {
			t.Errorf("Expected '2023/DEC/25', got '%s'", result)
		}
	})

	// Test case 9: Multiple capture groups - should return first one
	t.Run("returns first capture group when multiple exist", func(t *testing.T) {
		url := "https://example.com/2023/12/25/article"
		regexList := []*regexp.Regexp{
			regexp.MustCompile(`/(20\d{2})/(\d{2})/(\d{2})/`), // Three capture groups
		}
		
		result, found := ExtractFromURL(url, regexList)
		if !found {
			t.Fatal("Expected to find a match, but got false")
		}
		if result != "2023" {
			t.Errorf("Expected '2023' (first capture group), got '%s'", result)
		}
	})

	// Test case 10: Special characters in URL
	t.Run("handles special characters in URL", func(t *testing.T) {
		url := "https://example.com/2023/12/25/article?date=2023-12-25&utm_source=test#section"
		regexList := []*regexp.Regexp{
			regexp.MustCompile(`\?date=(20\d{2}-\d{2}-\d{2})&`),
		}
		
		result, found := ExtractFromURL(url, regexList)
		if !found {
			t.Fatal("Expected to find a match, but got false")
		}
		if result != "2023-12-25" {
			t.Errorf("Expected '2023-12-25', got '%s'", result)
		}
	})
}

// Benchmark tests to ensure performance is acceptable
func BenchmarkExtractFromURL(b *testing.B) {
	url := "https://example.com/2023/12/25/this-is-a-very-long-article-title-with-many-words"
	regexList := []*regexp.Regexp{
		regexp.MustCompile(`/(20\d{2}/\d{2}/\d{2})/`),
		regexp.MustCompile(`(20\d{2}-[01]\d-[0-3]\d)`),
		regexp.MustCompile(`/(20\d{2}/(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)/[0-3]\d)/`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractFromURL(url, regexList)
	}
}

func BenchmarkExtractFromURLNoMatch(b *testing.B) {
	url := "https://example.com/this-is-a-very-long-article-title-without-any-dates"
	regexList := []*regexp.Regexp{
		regexp.MustCompile(`/(20\d{2}/\d{2}/\d{2})/`),
		regexp.MustCompile(`(20\d{2}-[01]\d-[0-3]\d)`),
		regexp.MustCompile(`/(20\d{2}/(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)/[0-3]\d)/`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractFromURL(url, regexList)
	}
}