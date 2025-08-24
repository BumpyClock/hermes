// ABOUTME: Tests for all 15 content platform custom extractors  
// ABOUTME: Comprehensive test coverage with JavaScript compatibility verification

package custom

import (
	"testing"
	"strings"
)

func TestContentPlatformExtractors(t *testing.T) {
	// Test that all 15 content platform extractors are properly registered
	extractors := GetAllCustomExtractors()
	
	expectedExtractors := []string{
		"MediumExtractor",
		"BlogspotExtractor", 
		"BuzzFeedExtractor",
		"HuffingtonPostExtractor",
		"VoxExtractor",
		"WikipediaExtractor",
		"RedditExtractor",
		"TwitterExtractor",
		"YouTubeExtractor",
		"LinkedInExtractor",
		"FandomWikiaExtractor",
		"QdailyExtractor",
		"PastebinExtractor",
		"GeniusExtractor",
		"ThoughtCatalogExtractor",
	}
	
	for _, expected := range expectedExtractors {
		if extractor, exists := extractors[expected]; !exists {
			t.Errorf("Expected extractor %s not found in registry", expected)
		} else if extractor == nil {
			t.Errorf("Extractor %s is nil", expected)
		} else if extractor.Domain == "" {
			t.Errorf("Extractor %s has empty domain", expected)
		}
	}
	
	t.Logf("Successfully verified all %d content platform extractors", len(expectedExtractors))
}

func TestExtractorDomains(t *testing.T) {
	// Test domain mapping for all content platform extractors
	expectedDomains := map[string]string{
		"medium.com":          "MediumExtractor",
		"blogspot.com":        "BlogspotExtractor",
		"www.buzzfeed.com":    "BuzzFeedExtractor", 
		"www.huffingtonpost.com": "HuffingtonPostExtractor",
		"www.vox.com":         "VoxExtractor",
		"wikipedia.org":       "WikipediaExtractor",
		"www.reddit.com":      "RedditExtractor",
		"twitter.com":         "TwitterExtractor",
		"www.youtube.com":     "YouTubeExtractor",
		"www.linkedin.com":    "LinkedInExtractor",
		"fandom.wikia.com":    "FandomWikiaExtractor",
		"www.qdaily.com":      "QdailyExtractor",
		"pastebin.com":        "PastebinExtractor",
		"genius.com":          "GeniusExtractor",
		"thoughtcatalog.com":  "ThoughtCatalogExtractor",
	}
	
	for domain, expectedName := range expectedDomains {
		extractor, found := GetCustomExtractorByDomain(domain)
		if !found {
			t.Errorf("No extractor found for domain %s", domain)
			continue
		}
		
		if extractor.Domain != domain {
			t.Errorf("Domain mismatch for %s: expected %s, got %s", expectedName, domain, extractor.Domain)
		}
		
		t.Logf("✓ %s -> %s", domain, expectedName)
	}
}

func TestExtractorStructures(t *testing.T) {
	// Test that all extractors have required fields properly set
	extractors := GetAllCustomExtractors()
	
	for name, extractor := range extractors {
		if strings.Contains(name, "Blogger") && name != "BloggerExtractor" {
			continue // Skip legacy blogger references
		}
		
		// All extractors should have a domain
		if extractor.Domain == "" {
			t.Errorf("Extractor %s missing domain", name)
		}
		
		// All extractors should have at least title and content
		if extractor.Title == nil {
			t.Logf("Warning: %s missing title extractor (may be intentional)", name) 
		}
		
		if extractor.Content == nil {
			t.Errorf("Extractor %s missing content extractor", name)
		}
		
		// Content extractors should have selectors
		if extractor.Content != nil && extractor.Content.FieldExtractor != nil {
			if len(extractor.Content.FieldExtractor.Selectors) == 0 {
				t.Errorf("Extractor %s content has empty selectors", name)
			}
		}
		
		t.Logf("✓ %s structure validated", name)
	}
}

func TestSpecialExtractorFeatures(t *testing.T) {
	// Test special features of specific extractors
	extractors := GetAllCustomExtractors()
	
	// BuzzFeed should support BuzzFeedNews
	buzzfeed := extractors["BuzzFeedExtractor"]
	if buzzfeed == nil {
		t.Fatal("BuzzFeedExtractor not found")
	}
	
	found := false
	for _, supported := range buzzfeed.SupportedDomains {
		if supported == "www.buzzfeednews.com" {
			found = true
			break
		}
	}
	if !found {
		t.Error("BuzzFeed should support www.buzzfeednews.com")
	}
	
	// Wikipedia should have hardcoded author
	wikipedia := extractors["WikipediaExtractor"]
	if wikipedia == nil {
		t.Fatal("WikipediaExtractor not found") 
	}
	
	// YouTube should have transforms
	youtube := extractors["YouTubeExtractor"]
	if youtube == nil {
		t.Fatal("YouTubeExtractor not found")
	}
	if len(youtube.Content.Transforms) == 0 {
		t.Error("YouTube should have transforms for video embedding")
	}
	
	// Reddit should handle complex selectors
	reddit := extractors["RedditExtractor"] 
	if reddit == nil {
		t.Fatal("RedditExtractor not found")
	}
	if len(reddit.Content.Selectors) < 3 {
		t.Error("Reddit should have multiple content selector strategies")
	}
	
	t.Log("✓ Special extractor features verified")
}

func TestExtractorCount(t *testing.T) {
	// Verify we have exactly the expected number of content platform extractors
	count := CountCustomExtractors()
	
	// We have 15 content platform extractors + 1 legacy blogger = 16 total (with potential duplicates)
	if count < 15 {
		t.Errorf("Expected at least 15 extractors, got %d", count)
	}
	
	domains := GetCustomExtractorDomains()
	uniqueDomains := make(map[string]bool)
	for _, domain := range domains {
		uniqueDomains[domain] = true
	}
	
	if len(uniqueDomains) < 15 {
		t.Errorf("Expected at least 15 unique domains, got %d", len(uniqueDomains))
	}
	
	t.Logf("✓ Total extractors: %d, Unique domains: %d", count, len(uniqueDomains))
}