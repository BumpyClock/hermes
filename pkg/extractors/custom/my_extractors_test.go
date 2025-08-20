// ABOUTME: Tests for the 15 content platform extractors I implemented
// ABOUTME: Basic compilation and registration verification

package custom

import (
	"testing"
)

func TestMyContentPlatformExtractors(t *testing.T) {
	// Test that all my extractors are implemented and can be retrieved
	extractorFuncs := map[string]func() *CustomExtractor{
		"Medium":        GetMediumExtractor,
		"Blogspot":      GetBlogspotExtractor,
		"BuzzFeed":      GetBuzzFeedExtractor,
		"HuffPost":      GetHuffingtonPostExtractor,
		"Vox":           GetVoxExtractor,
		"Wikipedia":     GetWikipediaExtractor,
		"Reddit":        GetRedditExtractor,
		"Twitter":       GetTwitterExtractor,
		"YouTube":       GetYouTubeExtractor,
		"LinkedIn":      GetLinkedInExtractor,
		"FandomWikia":   GetFandomWikiaExtractor,
		"Qdaily":        GetQdailyExtractor,
		"Pastebin":      GetPastebinExtractor,
		"Genius":        GetGeniusExtractor,
		"ThoughtCatalog": GetThoughtCatalogExtractor,
	}
	
	for name, getExtractor := range extractorFuncs {
		extractor := getExtractor()
		if extractor == nil {
			t.Errorf("%s extractor returned nil", name)
			continue
		}
		
		if extractor.Domain == "" {
			t.Errorf("%s extractor has empty domain", name)
		}
		
		if extractor.Content == nil {
			t.Errorf("%s extractor missing content field", name)
		}
		
		t.Logf("✓ %s extractor: domain=%s", name, extractor.Domain)
	}
}

func TestExpectedDomains(t *testing.T) {
	expectedDomains := []string{
		"medium.com",
		"blogspot.com", 
		"www.buzzfeed.com",
		"www.huffingtonpost.com",
		"www.vox.com",
		"wikipedia.org",
		"www.reddit.com",
		"twitter.com",
		"www.youtube.com",
		"www.linkedin.com",
		"fandom.wikia.com",
		"www.qdaily.com",
		"pastebin.com",
		"genius.com",
		"thoughtcatalog.com",
	}
	
	for _, domain := range expectedDomains {
		extractor, found := GetCustomExtractorByDomain(domain)
		if !found {
			t.Errorf("No extractor found for domain: %s", domain)
			continue
		}
		
		if extractor.Domain != domain {
			t.Errorf("Domain mismatch for %s: got %s", domain, extractor.Domain)
		}
		
		t.Logf("✓ Domain %s -> Extractor found", domain)
	}
}