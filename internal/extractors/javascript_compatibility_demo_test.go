// ABOUTME: Comprehensive tests verifying 100% JavaScript compatibility of extractor selection logic
// ABOUTME: Tests demonstrate exact behavioral matching with JavaScript getExtractor() function

package extractors

import (
	"testing"
	"net/url"
	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
)

// TestJavaScriptCompatibleGetExtractor verifies the core JavaScript behavior
func TestJavaScriptCompatibleGetExtractor(t *testing.T) {
	tests := []struct {
		name           string
		inputURL       string
		expectedDomain string
		description    string
	}{
		{
			name:           "API extractor by hostname priority",
			inputURL:       "https://api.example.com/endpoint",
			expectedDomain: "api.example.com",
			description:    "Should match API extractor by exact hostname (Priority 1)",
		},
		{
			name:           "API extractor by base domain priority",
			inputURL:       "https://subdomain.example.com/page",
			expectedDomain: "example.com",
			description:    "Should match API extractor by base domain when hostname not found (Priority 2)",
		},
		{
			name:           "Static extractor by hostname priority",
			inputURL:       "https://www.nytimes.com/article",
			expectedDomain: "www.nytimes.com",
			description:    "Should match static extractor by hostname when API not found (Priority 3)",
		},
		{
			name:           "Static extractor by base domain priority",
			inputURL:       "https://edition.cnn.com/news",
			expectedDomain: "cnn.com",
			description:    "Should match static extractor by base domain when hostname not found (Priority 4)",
		},
		{
			name:           "Generic extractor fallback",
			inputURL:       "https://unknown.website.com/page",
			expectedDomain: "*",
			description:    "Should fallback to GenericExtractor when no matches found (Priority 6)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor, err := JavaScriptCompatibleGetExtractor(tt.inputURL, nil, nil)
			
			require.NoError(t, err, "Should not error on valid URLs")
			assert.Equal(t, tt.expectedDomain, extractor.GetDomain(), tt.description)
		})
	}
}

// TestCalculateBaseDomainJSCompat verifies exact JavaScript .split('.').slice(-2).join('.') behavior
func TestCalculateBaseDomainJSCompat(t *testing.T) {
	tests := []struct {
		name         string
		hostname     string
		expectedBase string
		jsEquivalent string
	}{
		{
			name:         "Standard domain",
			hostname:     "www.example.com",
			expectedBase: "example.com",
			jsEquivalent: "'www.example.com'.split('.').slice(-2).join('.')",
		},
		{
			name:         "Single part hostname",
			hostname:     "localhost",
			expectedBase: "localhost",
			jsEquivalent: "'localhost'.split('.').slice(-2).join('.')",
		},
		{
			name:         "Deep subdomain",
			hostname:     "a.b.c.d.example.com",
			expectedBase: "example.com",
			jsEquivalent: "'a.b.c.d.example.com'.split('.').slice(-2).join('.')",
		},
		{
			name:         "Two-part TLD behavior",
			hostname:     "www.bbc.co.uk", 
			expectedBase: "co.uk",
			jsEquivalent: "'www.bbc.co.uk'.split('.').slice(-2).join('.')  // JavaScript just takes last 2",
		},
		{
			name:         "Port included in split (JavaScript behavior)",
			hostname:     "example.com:3000",
			expectedBase: "example.com:3000",
			jsEquivalent: "'example.com:3000'.split('.').slice(-2).join('.')",
		},
		{
			name:         "IP address",
			hostname:     "192.168.1.1",
			expectedBase: "1.1",
			jsEquivalent: "'192.168.1.1'.split('.').slice(-2).join('.')",
		},
		{
			name:         "Empty hostname edge case",
			hostname:     "",
			expectedBase: "",
			jsEquivalent: "''.split('.').slice(-2).join('.')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBaseDomainJSCompat(tt.hostname)
			assert.Equal(t, tt.expectedBase, result, 
				"Go implementation should match JavaScript: %s", tt.jsEquivalent)
		})
	}
}

// TestJavaScriptURLParsing verifies URL parsing matches JavaScript URL.parse()
func TestJavaScriptURLParsing(t *testing.T) {
	tests := []struct {
		name             string
		inputURL         string
		expectedHostname string
		expectedBase     string
		expectError      bool
		jsNote           string
	}{
		{
			name:             "HTTPS URL",
			inputURL:         "https://www.example.com/path",
			expectedHostname: "www.example.com",
			expectedBase:     "example.com",
			expectError:      false,
			jsNote:           "new URL('https://www.example.com/path').hostname",
		},
		{
			name:             "Subdomain",
			inputURL:         "https://blog.example.com/post?id=123",
			expectedHostname: "blog.example.com",
			expectedBase:     "example.com",
			expectError:      false,
			jsNote:           "new URL('https://blog.example.com/post?id=123').hostname",
		},
		{
			name:        "Invalid URL",
			inputURL:    "not-a-url",
			expectError: true,
			jsNote:      "new URL('not-a-url') throws TypeError in JavaScript",
		},
		{
			name:        "Missing hostname",
			inputURL:    "https:///path/only",
			expectError: true,
			jsNote:      "URL with empty hostname should error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor, err := JavaScriptCompatibleGetExtractor(tt.inputURL, nil, nil)
			
			if tt.expectError {
				assert.Error(t, err, "Should error like JavaScript: %s", tt.jsNote)
			} else {
				require.NoError(t, err, "Should parse like JavaScript: %s", tt.jsNote)
				assert.NotEmpty(t, extractor.GetDomain(), "Should return valid extractor")
			}
		})
	}
}

// TestJavaScriptPriorityOrder verifies the exact priority order from JavaScript
func TestJavaScriptPriorityOrder(t *testing.T) {
	t.Run("API extractor beats static extractor", func(t *testing.T) {
		// Test case where both API and static have the same domain - API should win
		// This matches JavaScript: apiExtractors[hostname] || Extractors[hostname]
		
		url := "https://example.com/page"
		extractor, err := JavaScriptCompatibleGetExtractor(url, nil, nil)
		
		require.NoError(t, err)
		// Should get API extractor (priority 2) because of base domain match, 
		// not static extractor even though it might exist
		assert.Equal(t, "example.com", extractor.GetDomain())
	})
	
	t.Run("Hostname beats base domain", func(t *testing.T) {
		// Test demonstrating that hostname match beats base domain match
		// JavaScript: apiExtractors[hostname] || apiExtractors[baseDomain]
		
		url := "https://api.example.com/endpoint" 
		extractor, err := JavaScriptCompatibleGetExtractor(url, nil, nil)
		
		require.NoError(t, err)
		// Should get specific API extractor (priority 1), not base domain (priority 2)
		assert.Equal(t, "api.example.com", extractor.GetDomain())
	})
}

// TestJavaScriptHTMLDetection verifies HTML-based detection priority
func TestJavaScriptHTMLDetection(t *testing.T) {
	// Test HTML detection (priority 5) when no registry matches
	htmlContent := `<html>
		<head>
			<meta property="al:ios:app_name" content="Medium">
		</head>
		<body></body>
	</html>`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	require.NoError(t, err)
	
	// Use a URL that won't match any registry
	extractor, err := JavaScriptCompatibleGetExtractor("https://unknown.medium.site.com/article", nil, doc)
	
	require.NoError(t, err)
	assert.Equal(t, "medium.com", extractor.GetDomain(), 
		"Should detect Medium via HTML meta tags when no registry match")
}

// TestJavaScriptParsedURLParameter verifies pre-parsed URL parameter handling
func TestJavaScriptParsedURLParameter(t *testing.T) {
	// Test JavaScript function signature: getExtractor(url, parsedUrl, $)
	urlStr := "https://www.nytimes.com/article"
	parsedURL, err := url.Parse(urlStr)
	require.NoError(t, err)
	
	// Test with pre-parsed URL (JavaScript: parsedUrl || URL.parse(url))
	extractor, err := JavaScriptCompatibleGetExtractor(urlStr, parsedURL, nil)
	
	require.NoError(t, err)
	assert.Equal(t, "www.nytimes.com", extractor.GetDomain())
}

// TestJavaScriptEdgeCases verifies edge case handling matches JavaScript
func TestJavaScriptEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		inputURL    string
		expectError bool
		description string
	}{
		{
			name:        "Empty URL",
			inputURL:    "",
			expectError: true,
			description: "JavaScript would fail to parse empty URL",
		},
		{
			name:        "Protocol only",
			inputURL:    "https://",
			expectError: true, 
			description: "JavaScript URL parsing would fail",
		},
		{
			name:        "Valid minimal URL",
			inputURL:    "https://example.com",
			expectError: false,
			description: "JavaScript would parse successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor, err := JavaScriptCompatibleGetExtractor(tt.inputURL, nil, nil)
			
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotEmpty(t, extractor.GetDomain())
			}
		})
	}
}

// BenchmarkJavaScriptCompatibleGetExtractor tests performance
func BenchmarkJavaScriptCompatibleGetExtractor(b *testing.B) {
	testURLs := []string{
		"https://www.nytimes.com/article",
		"https://api.example.com/data",
		"https://edition.cnn.com/news", 
		"https://unknown.site.com/page",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := testURLs[i%len(testURLs)]
		_, err := JavaScriptCompatibleGetExtractor(url, nil, nil)
		if err != nil && !strings.Contains(err.Error(), "invalid URL") {
			b.Fatal(err)
		}
	}
}

// TestCompleteJavaScriptBehaviorMatch is a comprehensive compatibility test
func TestCompleteJavaScriptBehaviorMatch(t *testing.T) {
	t.Run("Complete priority chain simulation", func(t *testing.T) {
		testCases := []struct {
			url            string
			expectedDomain string
			priorityLevel  string
		}{
			// Priority 1: API extractor by hostname
			{"https://api.example.com/v1", "api.example.com", "Priority 1 (API hostname)"},
			
			// Priority 2: API extractor by base domain
			{"https://subdomain.example.com/page", "example.com", "Priority 2 (API base domain)"},
			
			// Priority 3: Static extractor by hostname 
			{"https://www.nytimes.com/news", "www.nytimes.com", "Priority 3 (Static hostname)"},
			
			// Priority 4: Static extractor by base domain
			{"https://edition.cnn.com/article", "cnn.com", "Priority 4 (Static base domain)"},
			
			// Priority 6: Generic fallback (skipping HTML detection for simplicity)
			{"https://totally.unknown.site.com/page", "*", "Priority 6 (Generic fallback)"},
		}
		
		for _, tc := range testCases {
			extractor, err := JavaScriptCompatibleGetExtractor(tc.url, nil, nil)
			require.NoError(t, err, "URL: %s", tc.url)
			assert.Equal(t, tc.expectedDomain, extractor.GetDomain(), 
				"URL: %s should match %s", tc.url, tc.priorityLevel)
		}
	})
}