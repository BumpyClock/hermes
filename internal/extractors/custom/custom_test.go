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

func TestMediumImageTransform(t *testing.T) {
	// Test the Medium image transform logic
	html := `<img width="50" /><img width="150" /><img />`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	// Apply transform to each image
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		if err := transformMediumImage(s); err != nil {
			t.Errorf("transformMediumImage returned error: %v", err)
		}
	})

	// Should remove images with width < 100
	remaining := doc.Find("img").Length()
	if remaining != 2 { // One with width 150, one without width attribute
		t.Errorf("Expected 2 remaining images, got %d", remaining)
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
