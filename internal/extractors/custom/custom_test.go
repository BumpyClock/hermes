// ABOUTME: Test file for custom extractors foundation and structure
// ABOUTME: Verifies Medium and Blogger extractors match JavaScript behavior exactly

package custom

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestMediumExtractor(t *testing.T) {
	extractor := GetMediumExtractor()
	
	// Test basic structure
	if extractor.Domain != "medium.com" {
		t.Errorf("Expected domain 'medium.com', got %s", extractor.Domain)
	}
	
	// Test selectors are properly defined
	if extractor.Title == nil || len(extractor.Title.Selectors) != 2 {
		t.Error("Medium extractor title selectors not properly defined")
	}
	
	if extractor.Content == nil || len(extractor.Content.Selectors) != 1 {
		t.Error("Medium extractor content selectors not properly defined")
	}
	
	// Test that transforms are defined
	if len(extractor.Content.Transforms) == 0 {
		t.Error("Medium extractor should have transforms defined")
	}
	
	// Verify specific selectors match JavaScript
	expectedSelectors := map[string]bool{
		"section span:first-of-type": true,
		"iframe":                     true,
		"figure":                     true,
		"img":                        true,
	}
	
	for selector := range expectedSelectors {
		if _, exists := extractor.Content.Transforms[selector]; !exists {
			t.Errorf("Missing expected transform for selector: %s", selector)
		}
	}
}

func TestBloggerExtractor(t *testing.T) {
	extractor := GetBloggerExtractor()
	
	// Test basic structure
	if extractor.Domain != "blogspot.com" {
		t.Errorf("Expected domain 'blogspot.com', got %s", extractor.Domain)
	}
	
	// Test supported domains are properly defined
	if len(extractor.SupportedDomains) == 0 {
		t.Error("Blogger extractor should have supported domains")
	}
	
	expectedDomains := []string{
		"www.blogspot.com",
		"blogspot.co.uk",
		"blogspot.ca",
	}
	
	for _, expectedDomain := range expectedDomains {
		found := false
		for _, supportedDomain := range extractor.SupportedDomains {
			if supportedDomain == expectedDomain {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing expected supported domain: %s", expectedDomain)
		}
	}
	
	// Test content selector matches JavaScript
	if extractor.Content == nil || len(extractor.Content.Selectors) != 1 {
		t.Error("Blogger extractor content selectors not properly defined")
	}
	
	// Verify noscript transform
	if _, exists := extractor.Content.Transforms["noscript"]; !exists {
		t.Error("Missing noscript transform for Blogger extractor")
	}
}

func TestCustomExtractorRegistry(t *testing.T) {
	extractors := GetAllCustomExtractors()
	
	// Should have at least Medium and Blogger
	if len(extractors) < 2 {
		t.Errorf("Expected at least 2 extractors, got %d", len(extractors))
	}
	
	// Check for expected extractors
	expectedExtractors := []string{"MediumExtractor", "BloggerExtractor"}
	for _, expected := range expectedExtractors {
		if _, exists := extractors[expected]; !exists {
			t.Errorf("Missing expected extractor: %s", expected)
		}
	}
}

func TestGetCustomExtractorByDomain(t *testing.T) {
	tests := []struct {
		domain   string
		expected string
		found    bool
	}{
		{"medium.com", "medium.com", true},
		{"blogspot.com", "blogspot.com", true},
		{"blogspot.co.uk", "blogspot.com", true}, // Supported domain
		{"www.blogspot.com", "blogspot.com", true}, // Supported domain
		{"example.com", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			extractor, found := GetCustomExtractorByDomain(tt.domain)
			
			if found != tt.found {
				t.Errorf("Expected found=%v, got %v", tt.found, found)
			}
			
			if found && extractor.Domain != tt.expected {
				t.Errorf("Expected domain %s, got %s", tt.expected, extractor.Domain)
			}
		})
	}
}

func TestMediumImageTransform(t *testing.T) {
	// Test the Medium image transform logic
	html := `<img width="50" /><img width="150" /><img />`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	
	// Apply transform to each image
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		transformMediumImage(s)
	})
	
	// Should remove images with width < 100
	remaining := doc.Find("img").Length()
	if remaining != 2 { // One with width 150, one without width attribute
		t.Errorf("Expected 2 remaining images, got %d", remaining)
	}
}

func TestStringTransform(t *testing.T) {
	// Test the string transform functionality
	transform := &StringTransform{TargetTag: "div"}
	
	html := `<noscript>Content here</noscript>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	
	selection := doc.Find("noscript")
	err = transform.Transform(selection)
	if err != nil {
		t.Error(err)
	}
	
	// Should have converted noscript to div
	if doc.Find("div").Length() != 1 {
		t.Error("String transform should have converted noscript to div")
	}
	
	if doc.Find("noscript").Length() != 0 {
		t.Error("Original noscript should have been replaced")
	}
}

func TestExtractorRegistryOperations(t *testing.T) {
	registry := NewExtractorRegistry()
	
	// Register Medium extractor
	medium := GetMediumExtractor()
	registry.Register(medium)
	
	// Should be able to retrieve it
	retrieved, exists := registry.Get("medium.com")
	if !exists {
		t.Error("Failed to retrieve registered extractor")
	}
	
	if retrieved.Domain != "medium.com" {
		t.Error("Retrieved wrong extractor")
	}
	
	// Test count
	if registry.Count() == 0 {
		t.Error("Registry count should be > 0")
	}
	
	// Test list
	domains := registry.List()
	if len(domains) == 0 {
		t.Error("Registry list should not be empty")
	}
}

func TestCustomExtractorCount(t *testing.T) {
	count := CountCustomExtractors()
	
	// Should have at least 2 (Medium + Blogger)
	if count < 2 {
		t.Errorf("Expected at least 2 extractors, got %d", count)
	}
	
	// Should be reasonable (not more than expected max)
	if count > 200 {
		t.Errorf("Extractor count %d seems too high", count)
	}
}

func TestCustomExtractorDomains(t *testing.T) {
	domains := GetCustomExtractorDomains()
	
	// Should have domains
	if len(domains) == 0 {
		t.Error("Should have custom extractor domains")
	}
	
	// Should include expected domains
	expectedDomains := []string{"medium.com", "blogspot.com"}
	for _, expected := range expectedDomains {
		found := false
		for _, domain := range domains {
			if domain == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing expected domain: %s", expected)
		}
	}
}

func BenchmarkGetAllCustomExtractors(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetAllCustomExtractors()
	}
}

func BenchmarkGetCustomExtractorByDomain(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetCustomExtractorByDomain("medium.com")
	}
}