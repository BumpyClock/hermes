package custom

import "testing"

func TestBlogspotUsesBloggerCanonicalExtractor(t *testing.T) {
	blogger := GetBloggerExtractor()
	if blogger == nil {
		t.Fatal("expected Blogger extractor")
	}
	if GetBlogspotExtractor() != blogger {
		t.Fatal("expected Blogspot compatibility getter to return Blogger extractor")
	}

	registered := GetAllCustomExtractors()
	if registered["BloggerExtractor"] != blogger {
		t.Fatal("expected Blogger extractor to be registered")
	}
	if _, exists := registered["BlogspotExtractor"]; exists {
		t.Fatal("expected duplicate Blogspot registry entry to be removed")
	}

	for _, domain := range []string{"blogspot.com", "www.blogspot.com", "blogspot.co.uk"} {
		extractor, found := GetCustomExtractorByDomain(domain)
		if !found {
			t.Fatalf("expected extractor for %s", domain)
		}
		if extractor != blogger {
			t.Fatalf("expected %s to resolve to Blogger extractor", domain)
		}
	}
}
