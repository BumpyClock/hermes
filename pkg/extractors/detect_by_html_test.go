// ABOUTME: Test file for DetectByHTML function that identifies extractors based on HTML meta tags
// ABOUTME: Verifies exact JavaScript behavior for HTML-based extractor detection

package extractors

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDetectByHTML(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string // Expected extractor name or empty for no match
	}{
		{
			name: "Medium extractor detection",
			html: `<!DOCTYPE html>
<html>
<head>
    <meta name="al:ios:app_name" value="Medium" />
    <title>Test Article</title>
</head>
<body>
    <p>Content here</p>
</body>
</html>`,
			expected: "MediumExtractor",
		},
		{
			name: "Blogger extractor detection", 
			html: `<!DOCTYPE html>
<html>
<head>
    <meta name="generator" value="blogger" />
    <title>Blog Post</title>
</head>
<body>
    <p>Blog content</p>
</body>
</html>`,
			expected: "BloggerExtractor",
		},
		{
			name: "Medium detection with additional meta tags",
			html: `<!DOCTYPE html>
<html>
<head>
    <meta property="og:title" content="Article Title" />
    <meta name="al:ios:app_name" value="Medium" />
    <meta name="author" content="John Doe" />
    <title>Test Article</title>
</head>
<body>
    <p>Content here</p>
</body>
</html>`,
			expected: "MediumExtractor",
		},
		{
			name: "Blogger detection with additional meta tags",
			html: `<!DOCTYPE html>
<html>
<head>
    <meta property="og:title" content="Blog Post Title" />
    <meta name="generator" value="blogger" />
    <meta name="description" content="Blog description" />
    <title>Blog Post</title>
</head>
<body>
    <p>Blog content</p>
</body>
</html>`,
			expected: "BloggerExtractor",
		},
		{
			name: "No matching extractor",
			html: `<!DOCTYPE html>
<html>
<head>
    <meta property="og:title" content="Generic Article" />
    <meta name="description" content="Generic description" />
    <title>Generic Article</title>
</head>
<body>
    <p>Generic content</p>
</body>
</html>`,
			expected: "", // No extractor should match
		},
		{
			name: "Wrong Medium app name (case sensitive)",
			html: `<!DOCTYPE html>
<html>
<head>
    <meta name="al:ios:app_name" value="medium" />
    <title>Test Article</title>
</head>
<body>
    <p>Content here</p>
</body>
</html>`,
			expected: "", // Should not match due to case sensitivity
		},
		{
			name: "Wrong Blogger generator (case sensitive)",
			html: `<!DOCTYPE html>
<html>
<head>
    <meta name="generator" value="Blogger" />
    <title>Blog Post</title>
</head>
<body>
    <p>Blog content</p>
</body>
</html>`,
			expected: "", // Should not match due to case sensitivity
		},
		{
			name: "Medium with different attribute structure",
			html: `<!DOCTYPE html>
<html>
<head>
    <meta name="al:ios:app_name" content="Medium" />
    <title>Test Article</title>
</head>
<body>
    <p>Content here</p>
</body>
</html>`,
			expected: "", // Should not match - JavaScript looks for value="Medium", not content="Medium"
		},
		{
			name: "Blogger with different attribute structure",
			html: `<!DOCTYPE html>
<html>
<head>
    <meta name="generator" content="blogger" />
    <title>Blog Post</title>
</head>
<body>
    <p>Blog content</p>
</body>
</html>`,
			expected: "", // Should not match - JavaScript looks for value="blogger", not content="blogger"
		},
		{
			name: "Both Medium and Blogger meta tags (should return first match)",
			html: `<!DOCTYPE html>
<html>
<head>
    <meta name="al:ios:app_name" value="Medium" />
    <meta name="generator" value="blogger" />
    <title>Conflicting Article</title>
</head>
<body>
    <p>Content here</p>
</body>
</html>`,
			expected: "MediumExtractor", // First match in selector order should win
		},
		{
			name: "Empty HTML",
			html: "",
			expected: "", // No extractor should match
		},
		{
			name: "Malformed HTML with Medium tag",
			html: `<meta name="al:ios:app_name" value="Medium" />
<p>Some content</p>`,
			expected: "MediumExtractor", // Should still work with malformed HTML
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			extractor := DetectByHTML(doc)
			
			if tt.expected == "" {
				// Expect nil/no extractor
				if extractor != nil {
					t.Errorf("DetectByHTML() = %v, want nil", extractor)
				}
			} else {
				// Expect specific extractor
				if extractor == nil {
					t.Errorf("DetectByHTML() = nil, want %s", tt.expected)
				} else if extractorName := getExtractorName(extractor); extractorName != tt.expected {
					t.Errorf("DetectByHTML() = %s, want %s", extractorName, tt.expected)
				}
			}
		})
	}
}

func TestDetectByHTMLJavaScriptCompatibility(t *testing.T) {
	// Test the exact JavaScript logic:
	// const selector = Reflect.ownKeys(Detectors).find(s => $(s).length > 0);
	// return Detectors[selector];
	
	t.Run("JavaScript selector matching behavior", func(t *testing.T) {
		// Test that the selectors match exactly as they would in JavaScript/jQuery
		
		testCases := []struct {
			name           string
			selector       string
			html           string
			shouldMatch    bool
			expectedLength int
		}{
			{
				name:     "Medium selector exact match",
				selector: `meta[name="al:ios:app_name"][value="Medium"]`,
				html:     `<meta name="al:ios:app_name" value="Medium" />`,
				shouldMatch: true,
				expectedLength: 1,
			},
			{
				name:     "Blogger selector exact match", 
				selector: `meta[name="generator"][value="blogger"]`,
				html:     `<meta name="generator" value="blogger" />`,
				shouldMatch: true,
				expectedLength: 1,
			},
			{
				name:     "Medium selector with extra attributes",
				selector: `meta[name="al:ios:app_name"][value="Medium"]`,
				html:     `<meta name="al:ios:app_name" value="Medium" id="app-meta" />`,
				shouldMatch: true,
				expectedLength: 1,
			},
			{
				name:     "Medium selector case sensitive value",
				selector: `meta[name="al:ios:app_name"][value="Medium"]`,
				html:     `<meta name="al:ios:app_name" value="medium" />`,
				shouldMatch: false,
				expectedLength: 0,
			},
			{
				name:     "Blogger selector case sensitive value",
				selector: `meta[name="generator"][value="blogger"]`,
				html:     `<meta name="generator" value="Blogger" />`,
				shouldMatch: false,
				expectedLength: 0,
			},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
				if err != nil {
					t.Fatalf("Failed to parse HTML: %v", err)
				}
				
				selection := doc.Find(tc.selector)
				actualLength := selection.Length()
				
				if tc.shouldMatch && actualLength != tc.expectedLength {
					t.Errorf("Selector %q should match %d elements, got %d", tc.selector, tc.expectedLength, actualLength)
				}
				
				if !tc.shouldMatch && actualLength != 0 {
					t.Errorf("Selector %q should not match any elements, got %d", tc.selector, actualLength)
				}
			})
		}
	})
}

func BenchmarkDetectByHTML(b *testing.B) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <meta name="al:ios:app_name" value="Medium" />
    <meta property="og:title" content="Test Article" />
    <title>Test Article</title>
</head>
<body>
    <p>Content here</p>
</body>
</html>`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		b.Fatalf("Failed to parse HTML: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectByHTML(doc)
	}
}

// Helper function to get extractor name for testing
// This will need to be implemented based on the actual extractor interface
func getExtractorName(extractor interface{}) string {
	// For now, return placeholder names
	// In real implementation, this would inspect the actual extractor type
	switch extractor.(type) {
	case *MediumExtractor:
		return "MediumExtractor"
	case *BloggerExtractor:
		return "BloggerExtractor"
	default:
		return "UnknownExtractor"
	}
}

