// ABOUTME: Test suite for simplified extractor selection logic with JavaScript compatibility verification
// ABOUTME: Focused tests for URL processing and extractor priority matching JavaScript behavior

package extractors

import (
	"testing"
	"net/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test hostname extraction matching JavaScript URL.parse(url).hostname
func TestExtractURLComponentsSimple(t *testing.T) {
	tests := []struct {
		name      string
		inputURL  string
		expectHostname string
		expectBaseDomain string
		expectError bool
	}{
		{
			name:             "Basic hostname",
			inputURL:         "https://www.nytimes.com/article",
			expectHostname:   "www.nytimes.com",
			expectBaseDomain: "nytimes.com",
			expectError:      false,
		},
		{
			name:             "Subdomain hostname", 
			inputURL:         "https://blog.example.com/post",
			expectHostname:   "blog.example.com",
			expectBaseDomain: "example.com",
			expectError:      false,
		},
		{
			name:             "Deep subdomain",
			inputURL:         "https://api.v2.service.example.com/endpoint",
			expectHostname:   "api.v2.service.example.com",
			expectBaseDomain: "example.com",
			expectError:      false,
		},
		{
			name:             "Two-part TLD",
			inputURL:         "https://www.bbc.co.uk/news",
			expectHostname:   "www.bbc.co.uk",
			expectBaseDomain: "co.uk", // JavaScript behavior - just takes last 2 parts
			expectError:      false,
		},
		{
			name:        "Invalid URL",
			inputURL:    "not-a-valid-url",
			expectError: true,
		},
		{
			name:        "Missing scheme",
			inputURL:    "www.example.com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostname, baseDomain, err := extractURLComponentsSimple(tt.inputURL, nil)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expectHostname, hostname, "Hostname should match JavaScript URL.parse behavior")
			assert.Equal(t, tt.expectBaseDomain, baseDomain, "Base domain should match JavaScript .split('.').slice(-2).join('.')")
		})
	}
}

// Test base domain calculation edge cases matching JavaScript logic exactly
func TestCalculateBaseDomainSimple(t *testing.T) {
	tests := []struct {
		name         string
		hostname     string
		expectedBase string
	}{
		{
			name:         "Standard domain",
			hostname:     "www.example.com",
			expectedBase: "example.com",
		},
		{
			name:         "Single part hostname",
			hostname:     "localhost",
			expectedBase: "localhost", // Only one part available
		},
		{
			name:         "Two part hostname",
			hostname:     "example.com",
			expectedBase: "example.com",
		},
		{
			name:         "Many subdomains",
			hostname:     "a.b.c.d.example.com",
			expectedBase: "example.com",
		},
		{
			name:         "Hostname with port",
			hostname:     "example.com:3000",
			expectedBase: "com:3000", // JavaScript behavior - port included in split
		},
		{
			name:         "Empty hostname",
			hostname:     "",
			expectedBase: "", // Edge case
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDomain := calculateBaseDomainSimple(tt.hostname)
			assert.Equal(t, tt.expectedBase, baseDomain, "Base domain calculation should match JavaScript logic")
		})
	}
}

// Test with pre-parsed URL matching JavaScript function signature
func TestGetExtractorSimpleWithParsedURL(t *testing.T) {
	inputURL := "https://www.nytimes.com/2023/article"
	parsedURL, err := url.Parse(inputURL)
	require.NoError(t, err)
	
	// Test with pre-parsed URL (matching JavaScript function signature)
	extractor, err := GetExtractorSimple(inputURL, parsedURL, nil)
	
	require.NoError(t, err)
	assert.NotNil(t, extractor)
	assert.Equal(t, "*", extractor.Domain, "Should return generic extractor when no matches")
}

// Test error handling
func TestGetExtractorSimpleErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		inputURL    string
		expectError bool
	}{
		{
			name:        "Empty URL should error", 
			inputURL:    "",
			expectError: true,
		},
		{
			name:        "Invalid URL should error",
			inputURL:    "://invalid-url",
			expectError: true,
		},
		{
			name:        "Valid URL should work",
			inputURL:    "https://example.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor, err := GetExtractorSimple(tt.inputURL, nil, nil)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, extractor)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, extractor)
				assert.Equal(t, "*", extractor.Domain) // Should fallback to generic
			}
		})
	}
}

// Test basic functionality without registry conflicts
func TestGetExtractorSimpleBasicFunctionality(t *testing.T) {
	testURLs := []string{
		"https://www.nytimes.com/article",
		"https://www.cnn.com/news", 
		"https://unknown.site.com/page",
	}
	
	for _, testURL := range testURLs {
		t.Run("URL: "+testURL, func(t *testing.T) {
			extractor, err := GetExtractorSimple(testURL, nil, nil)
			
			assert.NoError(t, err, "Should not error on valid URL")
			assert.NotNil(t, extractor, "Should return an extractor")
			assert.Equal(t, "*", extractor.Domain, "Should return generic extractor when no custom extractors exist")
		})
	}
}

// Benchmark extractor selection performance
func BenchmarkGetExtractorSimple(b *testing.B) {
	testURLs := []string{
		"https://www.nytimes.com/article",
		"https://www.cnn.com/news",
		"https://unknown.site.com/page",
		"https://api.service.com/endpoint",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := testURLs[i%len(testURLs)]
		_, err := GetExtractorSimple(url, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}