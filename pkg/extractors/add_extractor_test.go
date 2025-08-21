package extractors

import (
	"fmt"
	"sync"
	"testing"

	"github.com/BumpyClock/parser-go/pkg/parser"
)

func TestAddExtractor(t *testing.T) {
	// Clear registry before each test
	ClearAPIExtractors()

	tests := []struct {
		name            string
		extractor       *FullExtractor
		expectedError   bool
		expectedCount   int
		expectedDomains []string
	}{
		{
			name: "valid extractor with single domain",
			extractor: &FullExtractor{
				Domain: "example.com",
			},
			expectedError:   false,
			expectedCount:   1,
			expectedDomains: []string{"example.com"},
		},
		{
			name: "valid extractor with supported domains",
			extractor: &FullExtractor{
				Domain:           "nytimes.com",
				SupportedDomains: []string{"www.nytimes.com", "mobile.nytimes.com"},
			},
			expectedError:   false,
			expectedCount:   3,
			expectedDomains: []string{"nytimes.com", "www.nytimes.com", "mobile.nytimes.com"},
		},
		{
			name:          "invalid extractor - nil",
			extractor:     nil,
			expectedError: true,
			expectedCount: 0,
		},
		{
			name: "invalid extractor with empty domain",
			extractor: &FullExtractor{
				Domain: "",
			},
			expectedError: true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear registry before each test case
			ClearAPIExtractors()

			result := AddExtractor(tt.extractor)

			if tt.expectedError {
				// Check if result is an error
				if errorResult, ok := result.(ExtractorError); ok {
					if !errorResult.Error {
						t.Errorf("Expected error result, got success")
					}
					if errorResult.Message != "Unable to add custom extractor. Invalid parameters." {
						t.Errorf("Expected error message 'Unable to add custom extractor. Invalid parameters.', got '%s'", errorResult.Message)
					}
				} else {
					t.Errorf("Expected ExtractorError, got %T", result)
				}

				// Verify registry is unchanged
				if count := GetExtractorCount(); count != tt.expectedCount {
					t.Errorf("Expected %d extractors in registry, got %d", tt.expectedCount, count)
				}
			} else {
				// Check if result is a map of extractors
				if extractorMap, ok := result.(map[string]*FullExtractor); ok {
					if len(extractorMap) != tt.expectedCount {
						t.Errorf("Expected %d extractors, got %d", tt.expectedCount, len(extractorMap))
					}

					// Verify all expected domains are present
					for _, domain := range tt.expectedDomains {
						if _, exists := extractorMap[domain]; !exists {
							t.Errorf("Expected domain %s not found in result", domain)
						}
					}
				} else {
					t.Errorf("Expected map[string]*FullExtractor, got %T", result)
				}

				// Verify registry state
				if count := GetExtractorCount(); count != tt.expectedCount {
					t.Errorf("Expected %d extractors in registry, got %d", tt.expectedCount, count)
				}

				// Verify individual domain lookups
				for _, domain := range tt.expectedDomains {
					if !HasExtractor(domain) {
						t.Errorf("Domain %s not found in registry", domain)
					}

					if extractor, exists := GetExtractorByDomain(domain); !exists {
						t.Errorf("Could not retrieve extractor for domain %s", domain)
					} else if extractor.Domain != tt.extractor.Domain {
						t.Errorf("Retrieved extractor has wrong primary domain: expected %s, got %s", tt.extractor.Domain, extractor.Domain)
					}
				}
			}
		})
	}
}

func TestAddExtractorWithFullConfiguration(t *testing.T) {
	ClearAPIExtractors()

	// Test with a full extractor configuration including extended types
	extractor := &FullExtractor{
		Domain:           "test.com",
		SupportedDomains: []string{"www.test.com"},
		Title: &FieldExtractor{
			Selectors:      parser.NewSelectorList([]interface{}{"h1.title", "h1"}),
			DefaultCleaner: true,
		},
		Author: &FieldExtractor{
			Selectors:      parser.NewSelectorList([]interface{}{[]interface{}{"meta[name='author']", "content"}}),
			DefaultCleaner: true,
		},
		Content: &ContentExtractor{
			Selectors:      parser.NewSelectorList([]interface{}{".article-body", "main"}),
			DefaultCleaner: true,
			Clean:          []string{".ads", ".related"},
		},
		DatePublished: &FieldExtractor{
			Selectors:      parser.NewSelectorList([]interface{}{[]interface{}{"time[datetime]", "datetime"}}),
			DefaultCleaner: true,
		},
		Extend: map[string]*FieldExtractor{
			"category": {
				Selectors:     parser.NewSelectorList([]interface{}{".category", ".tag"}),
				AllowMultiple: true,
			},
		},
	}

	result := AddExtractor(extractor)

	// Should return a map (not an error)
	extractorMap, ok := result.(map[string]*FullExtractor)
	if !ok {
		t.Fatalf("Expected map[string]*FullExtractor, got %T", result)
	}

	// Should have exactly 2 entries
	if len(extractorMap) != 2 {
		t.Errorf("Expected 2 extractors, got %d", len(extractorMap))
	}

	// Check that extended types are preserved
	if retrievedExtractor, exists := extractorMap["test.com"]; exists {
		if retrievedExtractor.Extend == nil {
			t.Errorf("Extended types not preserved")
		} else if categoryExtractor, hasCat := retrievedExtractor.Extend["category"]; !hasCat {
			t.Errorf("Category extended type not found")
		} else if !categoryExtractor.AllowMultiple {
			t.Errorf("Category AllowMultiple not preserved")
		}
	}
}

func TestAddExtractorConcurrency(t *testing.T) {
	ClearAPIExtractors()

	// Test concurrent access to the registry
	var wg sync.WaitGroup
	numGoroutines := 10
	extractorsPerGoroutine := 5

	// Add extractors concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < extractorsPerGoroutine; j++ {
				extractor := &FullExtractor{
					Domain: fmt.Sprintf("test%d-%d.com", routineID, j),
				}
				AddExtractor(extractor)
			}
		}(i)
	}

	wg.Wait()

	// Verify all extractors were added
	expectedCount := numGoroutines * extractorsPerGoroutine
	if count := GetExtractorCount(); count != expectedCount {
		t.Errorf("Expected %d extractors after concurrent addition, got %d", expectedCount, count)
	}

	// Test concurrent reads while writing
	var readWg sync.WaitGroup
	readResults := make([]int, 10)

	// Start readers
	for i := 0; i < 10; i++ {
		readWg.Add(1)
		go func(readerID int) {
			defer readWg.Done()
			for j := 0; j < 100; j++ {
				count := GetExtractorCount()
				readResults[readerID] = count
			}
		}(i)
	}

	// Add more extractors while readers are running
	for i := 0; i < 5; i++ {
		extractor := &FullExtractor{
			Domain: fmt.Sprintf("concurrent-test-%d.com", i),
		}
		AddExtractor(extractor)
	}

	readWg.Wait()

	// Verify final count
	finalCount := GetExtractorCount()
	if finalCount < expectedCount {
		t.Errorf("Expected at least %d extractors after concurrent operations, got %d", expectedCount, finalCount)
	}
}

func TestMergeSupportedDomains(t *testing.T) {
	tests := []struct {
		name      string
		extractor *FullExtractor
		expected  map[string]*FullExtractor
	}{
		{
			name: "single domain only",
			extractor: &FullExtractor{
				Domain: "example.com",
			},
			expected: map[string]*FullExtractor{
				"example.com": {Domain: "example.com"},
			},
		},
		{
			name: "domain with supported domains",
			extractor: &FullExtractor{
				Domain:           "example.com",
				SupportedDomains: []string{"www.example.com", "m.example.com"},
			},
			expected: map[string]*FullExtractor{
				"example.com": {
					Domain:           "example.com",
					SupportedDomains: []string{"www.example.com", "m.example.com"},
				},
				"www.example.com": {
					Domain:           "example.com",
					SupportedDomains: []string{"www.example.com", "m.example.com"},
				},
				"m.example.com": {
					Domain:           "example.com",
					SupportedDomains: []string{"www.example.com", "m.example.com"},
				},
			},
		},
		{
			name: "empty supported domains array",
			extractor: &FullExtractor{
				Domain:           "test.com",
				SupportedDomains: []string{},
			},
			expected: map[string]*FullExtractor{
				"test.com": {
					Domain:           "test.com",
					SupportedDomains: []string{},
				},
			},
		},
		{
			name: "nil supported domains",
			extractor: &FullExtractor{
				Domain:           "nil-test.com",
				SupportedDomains: nil,
			},
			expected: map[string]*FullExtractor{
				"nil-test.com": {
					Domain:           "nil-test.com",
					SupportedDomains: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeSupportedDomains(tt.extractor)

			// Verify all expected domains are present
			for domain := range tt.expected {
				if _, exists := result[domain]; !exists {
					t.Errorf("Expected domain %s not found in result", domain)
				}
			}

			// Verify no extra domains are present
			if len(result) != len(tt.expected) {
				t.Errorf("Result has %d domains, expected %d", len(result), len(tt.expected))
			}

			// Verify each domain maps to the correct extractor
			for domain, expectedExtractor := range tt.expected {
				if resultExtractor, exists := result[domain]; exists {
					if resultExtractor.Domain != expectedExtractor.Domain {
						t.Errorf("Domain %s maps to wrong extractor domain: expected %s, got %s",
							domain, expectedExtractor.Domain, resultExtractor.Domain)
					}
				}
			}
		})
	}
}

func TestJavaScriptCompatibility(t *testing.T) {
	ClearAPIExtractors()

	// Test the exact JavaScript behavior from add-extractor.js
	extractor := &FullExtractor{
		Domain:           "test.com",
		SupportedDomains: []string{"www.test.com"},
	}

	result := AddExtractor(extractor)

	// Should return a map (not an error) - JavaScript returns apiExtractors object
	extractorMap, ok := result.(map[string]*FullExtractor)
	if !ok {
		t.Fatalf("Expected map[string]*FullExtractor, got %T", result)
	}

	// Should have exactly 2 entries
	if len(extractorMap) != 2 {
		t.Errorf("Expected 2 extractors, got %d", len(extractorMap))
	}

	// Should contain both domains
	expectedDomains := []string{"test.com", "www.test.com"}
	for _, domain := range expectedDomains {
		if _, exists := extractorMap[domain]; !exists {
			t.Errorf("Domain %s not found in result", domain)
		}
	}

	// Test error case compatibility
	invalidResult := AddExtractor(nil)

	if errorStruct, ok := invalidResult.(ExtractorError); !ok {
		t.Errorf("Expected ExtractorError for invalid extractor, got %T", invalidResult)
	} else {
		if !errorStruct.Error {
			t.Errorf("Expected error field to be true")
		}
		if errorStruct.Message != "Unable to add custom extractor. Invalid parameters." {
			t.Errorf("Expected specific error message, got: %s", errorStruct.Message)
		}
	}

	// Test that registry persists between calls (JavaScript Object.assign behavior)
	extractor2 := &FullExtractor{
		Domain: "another.com",
	}
	result2 := AddExtractor(extractor2)

	extractorMap2, ok := result2.(map[string]*FullExtractor)
	if !ok {
		t.Fatalf("Expected map[string]*FullExtractor, got %T", result2)
	}

	// Should now have 3 entries (2 from first + 1 from second)
	if len(extractorMap2) != 3 {
		t.Errorf("Expected 3 extractors after second addition, got %d", len(extractorMap2))
	}

	// Should contain all domains from both extractors
	allExpectedDomains := []string{"test.com", "www.test.com", "another.com"}
	for _, domain := range allExpectedDomains {
		if _, exists := extractorMap2[domain]; !exists {
			t.Errorf("Domain %s not found in result after second addition", domain)
		}
	}
}

func TestExtendedTypesSupport(t *testing.T) {
	ClearAPIExtractors()

	// Test extractor with extended types
	extractor := &FullExtractor{
		Domain: "news.example.com",
		Title: &FieldExtractor{
			Selectors: parser.NewSelectorList([]interface{}{"h1.headline"}),
		},
		Extend: map[string]*FieldExtractor{
			"category": {
				Selectors:     parser.NewSelectorList([]interface{}{".category a"}),
				AllowMultiple: true,
			},
			"tags": {
				Selectors:     parser.NewSelectorList([]interface{}{".tags .tag"}),
				AllowMultiple: true,
			},
			"source": {
				Selectors: parser.NewSelectorList([]interface{}{".source"}),
			},
		},
	}

	result := AddExtractor(extractor)

	extractorMap, ok := result.(map[string]*FullExtractor)
	if !ok {
		t.Fatalf("Expected map[string]*FullExtractor, got %T", result)
	}

	retrievedExtractor := extractorMap["news.example.com"]
	if retrievedExtractor == nil {
		t.Fatal("Extractor not found in registry")
	}

	// Verify extended types are preserved
	if retrievedExtractor.Extend == nil {
		t.Fatal("Extended types not preserved")
	}

	if len(retrievedExtractor.Extend) != 3 {
		t.Errorf("Expected 3 extended types, got %d", len(retrievedExtractor.Extend))
	}

	// Verify specific extended type configuration
	if categoryExtractor, exists := retrievedExtractor.Extend["category"]; exists {
		if !categoryExtractor.AllowMultiple {
			t.Errorf("Category AllowMultiple not preserved")
		}
		if len(categoryExtractor.Selectors) != 1 {
			t.Errorf("Expected 1 selector for category, got %d", len(categoryExtractor.Selectors))
		}
	} else {
		t.Errorf("Category extended type not found")
	}

	if sourceExtractor, exists := retrievedExtractor.Extend["source"]; exists {
		if sourceExtractor.AllowMultiple {
			t.Errorf("Source AllowMultiple should be false by default")
		}
	} else {
		t.Errorf("Source extended type not found")
	}
}
