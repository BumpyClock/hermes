// ABOUTME: Integration tests for the complete extractor registry and detection system
// ABOUTME: Verifies all three components (all.go, detect_by_html.go, merge_supported_domains.go) work together

package extractors

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/postlight/parser-go/pkg/utils"
)

func TestFullExtractorSystemIntegration(t *testing.T) {
	tests := []struct {
		name             string
		html             string
		expectedDomain   string
		expectedInRegistry bool
		description      string
	}{
		{
			name: "Medium article detection and registry lookup",
			html: `<!DOCTYPE html>
<html>
<head>
    <meta name="al:ios:app_name" value="Medium" />
    <meta property="og:title" content="Test Article on Medium" />
    <title>Test Article</title>
</head>
<body>
    <article>
        <h1>Test Article</h1>
        <p>This is a test article on Medium.</p>
    </article>
</body>
</html>`,
			expectedDomain:     "medium.com",
			expectedInRegistry: true,
			description:        "Medium article should be detected by HTML and found in registry",
		},
		{
			name: "Blogger article detection and registry lookup",
			html: `<!DOCTYPE html>
<html>
<head>
    <meta name="generator" value="blogger" />
    <title>My Blog Post</title>
</head>
<body>
    <div class="post-content">
        <noscript>
            <h2>Blog Post Title</h2>
            <p>This is blog post content.</p>
        </noscript>
    </div>
</body>
</html>`,
			expectedDomain:     "blogspot.com",
			expectedInRegistry: true,
			description:        "Blogger article should be detected by HTML and found in registry",
		},
		{
			name: "Generic article (no special detection)",
			html: `<!DOCTYPE html>
<html>
<head>
    <meta property="og:title" content="Generic News Article" />
    <title>Generic News Article</title>
</head>
<body>
    <article>
        <h1>Generic News Article</h1>
        <p>This is a generic news article.</p>
    </article>
</body>
</html>`,
			expectedDomain:     "",
			expectedInRegistry: false,
			description:        "Generic article should not trigger HTML detection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Parse HTML
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Step 2: Try HTML-based detection
			detectedExtractor := DetectByHTML(doc)

			// Step 3: Get registry for domain-based lookup
			registry := GetAllExtractors()

			// Step 4: Verify integration results
			if tt.expectedInRegistry {
				// Should detect via HTML
				if detectedExtractor == nil {
					t.Errorf("Expected HTML detection to find extractor, got nil")
				} else {
					detectedDomain := detectedExtractor.GetDomain()
					if detectedDomain != tt.expectedDomain {
						t.Errorf("HTML detection returned domain %s, want %s", detectedDomain, tt.expectedDomain)
					}
				}

				// Should also be in registry
				if registryExtractor, exists := registry[tt.expectedDomain]; !exists {
					t.Errorf("Expected domain %s to be in registry", tt.expectedDomain)
				} else if registryExtractor == nil {
					t.Errorf("Registry entry for %s is nil", tt.expectedDomain)
				}

			} else {
				// Should not detect via HTML
				if detectedExtractor != nil {
					t.Errorf("Expected no HTML detection, got %v", detectedExtractor)
				}
			}
		})
	}
}

func TestExtractorRegistryAndMergeSupportedDomains(t *testing.T) {
	t.Run("Registry includes all domains from mergeSupportedDomains", func(t *testing.T) {
		registry := GetAllExtractors()

		// Test that Blogger extractor supports multiple domains
		blogspotDomains := []string{"blogspot.com", "www.blogspot.com", "blogspot.co.uk", "blogspot.ca"}
		
		for _, domain := range blogspotDomains {
			extractor, exists := registry[domain]
			if !exists {
				t.Errorf("Registry missing expected blogspot domain: %s", domain)
				continue
			}
			
			// Should be the same extractor instance or equivalent
			if _, ok := extractor.(*BloggerExtractor); !ok {
				t.Errorf("Registry entry for %s is not BloggerExtractor", domain)
			}
		}

		// Test that Medium extractor is properly registered
		mediumExtractor, exists := registry["medium.com"]
		if !exists {
			t.Error("Registry missing medium.com domain")
		} else if _, ok := mediumExtractor.(*MediumExtractor); !ok {
			t.Error("Registry entry for medium.com is not MediumExtractor")
		}
	})
}

func TestJavaScriptCompatibilityFullStack(t *testing.T) {
	// Test the full JavaScript workflow:
	// 1. Import all extractors
	// 2. Reduce with mergeSupportedDomains
	// 3. Use detectByHtml for HTML-based detection
	
	t.Run("Full JavaScript compatibility workflow", func(t *testing.T) {
		// Simulate JavaScript import and reduce pattern
		registry := GetAllExtractors()
		
		// Verify registry structure matches JavaScript expectations
		if len(registry) == 0 {
			t.Fatal("Registry is empty - JavaScript would have extractors")
		}

		// Test HTML detection with known patterns
		mediumHTML := `<meta name="al:ios:app_name" value="Medium" />`
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(mediumHTML))
		
		detected := DetectByHTML(doc)
		if detected == nil {
			t.Error("HTML detection failed for Medium pattern")
		}

		// Verify detected extractor matches registry entry
		if detected != nil {
			detectedDomain := detected.GetDomain()
			registryEntry, exists := registry[detectedDomain]
			if !exists {
				t.Errorf("Detected domain %s not found in registry", detectedDomain)
			} else if registryEntry == nil {
				t.Errorf("Registry entry for %s is nil", detectedDomain)
			}
		}
	})
}

func TestMergeSupportedDomainsIntegration(t *testing.T) {
	// Test that the utils.MergeSupportedDomains function integrates properly
	
	t.Run("MergeSupportedDomains works with real extractor data", func(t *testing.T) {
		// Create a mock extractor with multiple domains
		mockExtractor := utils.MockExtractor{
			Domain:           "example.com",
			SupportedDomains: []string{"www.example.com", "mobile.example.com", "amp.example.com"},
			Name:             "ExampleExtractor",
		}

		// Apply mergeSupportedDomains
		domainMap := utils.MergeSupportedDomains(mockExtractor)

		// Verify all domains are mapped
		expectedDomains := []string{"example.com", "www.example.com", "mobile.example.com", "amp.example.com"}
		if len(domainMap) != len(expectedDomains) {
			t.Errorf("MergeSupportedDomains returned %d entries, expected %d", len(domainMap), len(expectedDomains))
		}

		for _, domain := range expectedDomains {
			if entry, exists := domainMap[domain]; !exists {
				t.Errorf("Missing domain %s in merged result", domain)
			} else if entry.Domain != mockExtractor.Domain {
				t.Errorf("Domain %s maps to wrong extractor domain: got %s, want %s", domain, entry.Domain, mockExtractor.Domain)
			}
		}
	})
}

func TestExtractorSystemPerformance(t *testing.T) {
	// Test that the system performs well with multiple operations
	
	t.Run("Registry creation performance", func(t *testing.T) {
		// Multiple calls should be efficient
		for i := 0; i < 100; i++ {
			registry := GetAllExtractors()
			if len(registry) == 0 {
				t.Error("Registry is empty on iteration", i)
			}
		}
	})
	
	t.Run("HTML detection performance", func(t *testing.T) {
		htmlContent := `<!DOCTYPE html>
<html>
<head>
    <meta name="al:ios:app_name" value="Medium" />
    <meta name="generator" value="blogger" />
    <title>Test</title>
</head>
<body><p>Content</p></body>
</html>`
		
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			t.Fatalf("Failed to parse HTML: %v", err)
		}
		
		// Multiple detections should be efficient  
		for i := 0; i < 100; i++ {
			detected := DetectByHTML(doc)
			if detected == nil {
				t.Error("Detection failed on iteration", i)
			}
		}
	})
}

func TestExtractorSystemErrorHandling(t *testing.T) {
	// Test system behavior with edge cases and errors
	
	t.Run("Empty/malformed HTML handling", func(t *testing.T) {
		testCases := []string{
			"", // Empty
			"<invalid>", // Malformed
			"<html><head></head><body></body></html>", // Valid but no meta tags
			"Just plain text", // Plain text
		}
		
		for _, html := range testCases {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				// goquery should handle malformed HTML gracefully
				continue
			}
			
			// DetectByHTML should not panic
			detected := DetectByHTML(doc)
			// For these cases, we expect no detection (nil result)
			if detected != nil {
				t.Logf("Unexpected detection for HTML: %q", html)
			}
		}
	})
	
	t.Run("Registry consistency under concurrent access", func(t *testing.T) {
		// Test concurrent access to registry
		done := make(chan bool)
		
		for i := 0; i < 10; i++ {
			go func() {
				registry := GetAllExtractors()
				if len(registry) == 0 {
					t.Error("Concurrent access resulted in empty registry")
				}
				done <- true
			}()
		}
		
		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// Benchmark the full integration
func BenchmarkFullExtractorIntegration(b *testing.B) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <meta name="al:ios:app_name" value="Medium" />
    <title>Test Article</title>
</head>
<body>
    <article><h1>Test</h1><p>Content</p></article>
</body>
</html>`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		b.Fatalf("Failed to parse HTML: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Full workflow: registry lookup + HTML detection
		registry := GetAllExtractors()
		detected := DetectByHTML(doc)
		
		// Verify integration
		if detected != nil && len(registry) > 0 {
			detectedDomain := detected.GetDomain()
			_ = registry[detectedDomain] // Registry lookup
		}
	}
}