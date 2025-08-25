// ABOUTME: Test file for MergeSupportedDomains function that creates domain-to-extractor mappings
// ABOUTME: Verifies exact JavaScript behavior for multi-domain extractor registration

package utils

import (
	"reflect"
	"testing"
)

func TestMergeSupportedDomains(t *testing.T) {
	tests := []struct {
		name      string
		extractor MockExtractor
		want      map[string]MockExtractor
	}{
		{
			name: "Single domain extractor (no supportedDomains)",
			extractor: MockExtractor{
				Domain: "example.com",
				Name:   "ExampleExtractor",
			},
			want: map[string]MockExtractor{
				"example.com": {
					Domain: "example.com", 
					Name:   "ExampleExtractor",
				},
			},
		},
		{
			name: "Multi-domain extractor with supportedDomains",
			extractor: MockExtractor{
				Domain:           "nytimes.com",
				SupportedDomains: []string{"www.nytimes.com", "mobile.nytimes.com"},
				Name:             "NYTimesExtractor",
			},
			want: map[string]MockExtractor{
				"nytimes.com": {
					Domain:           "nytimes.com",
					SupportedDomains: []string{"www.nytimes.com", "mobile.nytimes.com"},
					Name:             "NYTimesExtractor",
				},
				"www.nytimes.com": {
					Domain:           "nytimes.com",
					SupportedDomains: []string{"www.nytimes.com", "mobile.nytimes.com"},
					Name:             "NYTimesExtractor",
				},
				"mobile.nytimes.com": {
					Domain:           "nytimes.com",
					SupportedDomains: []string{"www.nytimes.com", "mobile.nytimes.com"},
					Name:             "NYTimesExtractor",
				},
			},
		},
		{
			name: "Medium extractor (real-world example)",
			extractor: MockExtractor{
				Domain: "medium.com",
				Name:   "MediumExtractor",
			},
			want: map[string]MockExtractor{
				"medium.com": {
					Domain: "medium.com",
					Name:   "MediumExtractor",
				},
			},
		},
		{
			name: "Empty supportedDomains slice",
			extractor: MockExtractor{
				Domain:           "test.com",
				SupportedDomains: []string{},
				Name:             "TestExtractor",
			},
			want: map[string]MockExtractor{
				"test.com": {
					Domain:           "test.com",
					SupportedDomains: []string{},
					Name:             "TestExtractor",
				},
			},
		},
		{
			name: "Large supportedDomains list",
			extractor: MockExtractor{
				Domain: "news.com",
				SupportedDomains: []string{
					"www.news.com",
					"mobile.news.com", 
					"amp.news.com",
					"beta.news.com",
					"archive.news.com",
				},
				Name: "NewsExtractor",
			},
			want: map[string]MockExtractor{
				"news.com": {
					Domain: "news.com",
					SupportedDomains: []string{
						"www.news.com",
						"mobile.news.com",
						"amp.news.com", 
						"beta.news.com",
						"archive.news.com",
					},
					Name: "NewsExtractor",
				},
				"www.news.com": {
					Domain: "news.com",
					SupportedDomains: []string{
						"www.news.com",
						"mobile.news.com",
						"amp.news.com",
						"beta.news.com", 
						"archive.news.com",
					},
					Name: "NewsExtractor",
				},
				"mobile.news.com": {
					Domain: "news.com",
					SupportedDomains: []string{
						"www.news.com",
						"mobile.news.com",
						"amp.news.com",
						"beta.news.com",
						"archive.news.com",
					},
					Name: "NewsExtractor",
				},
				"amp.news.com": {
					Domain: "news.com",
					SupportedDomains: []string{
						"www.news.com",
						"mobile.news.com",
						"amp.news.com",
						"beta.news.com",
						"archive.news.com",
					},
					Name: "NewsExtractor",
				},
				"beta.news.com": {
					Domain: "news.com",
					SupportedDomains: []string{
						"www.news.com",
						"mobile.news.com",
						"amp.news.com",
						"beta.news.com",
						"archive.news.com",
					},
					Name: "NewsExtractor",
				},
				"archive.news.com": {
					Domain: "news.com",
					SupportedDomains: []string{
						"www.news.com",
						"mobile.news.com",
						"amp.news.com",
						"beta.news.com",
						"archive.news.com",
					},
					Name: "NewsExtractor",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeSupportedDomains(tt.extractor)
			
			// Check that we got the expected number of domains
			if len(got) != len(tt.want) {
				t.Errorf("MergeSupportedDomains() returned %d entries, want %d", len(got), len(tt.want))
			}
			
			// Check each domain mapping
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeSupportedDomains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeSupportedDomainsJavaScriptCompatibility(t *testing.T) {
	// This test verifies the exact JavaScript behavior:
	// extractor.supportedDomains
	//   ? merge(extractor, [extractor.domain, ...extractor.supportedDomains])
	//   : merge(extractor, [extractor.domain]);
	
	t.Run("JavaScript ternary operator behavior", func(t *testing.T) {
		// Test with nil supportedDomains (should use single domain)
		extractorWithNil := MockExtractor{
			Domain: "single.com",
			Name:   "SingleExtractor",
			// SupportedDomains is nil by default
		}
		
		resultWithNil := MergeSupportedDomains(extractorWithNil)
		expectedWithNil := map[string]MockExtractor{
			"single.com": extractorWithNil,
		}
		
		if !reflect.DeepEqual(resultWithNil, expectedWithNil) {
			t.Errorf("Nil supportedDomains: got %v, want %v", resultWithNil, expectedWithNil)
		}
		
		// Test with empty supportedDomains slice (should still use single domain)
		extractorWithEmpty := MockExtractor{
			Domain:           "empty.com",
			SupportedDomains: []string{},
			Name:             "EmptyExtractor",
		}
		
		resultWithEmpty := MergeSupportedDomains(extractorWithEmpty)
		expectedWithEmpty := map[string]MockExtractor{
			"empty.com": extractorWithEmpty,
		}
		
		if !reflect.DeepEqual(resultWithEmpty, expectedWithEmpty) {
			t.Errorf("Empty supportedDomains: got %v, want %v", resultWithEmpty, expectedWithEmpty)
		}
	})
}

func BenchmarkMergeSupportedDomains(b *testing.B) {
	extractor := MockExtractor{
		Domain: "benchmark.com",
		SupportedDomains: []string{
			"www.benchmark.com",
			"mobile.benchmark.com",
			"amp.benchmark.com",
		},
		Name: "BenchmarkExtractor",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MergeSupportedDomains(extractor)
	}
}