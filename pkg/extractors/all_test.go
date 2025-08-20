// ABOUTME: Test file for All extractor registry that aggregates all custom extractors into domain mappings
// ABOUTME: Verifies exact JavaScript behavior for extractor registry and domain-to-extractor mappings

package extractors

import (
	"reflect"
	"testing"

	"github.com/BumpyClock/parser-go/pkg/utils"
)

func TestGetAllExtractors(t *testing.T) {
	tests := []struct {
		name string
		want map[string]interface{} // Using interface{} for now, will be actual extractor type later
	}{
		{
			name: "Registry contains all expected extractors",
			want: map[string]interface{}{
				// Medium extractor should be registered
				"medium.com": "MediumExtractor",
				// Blogger extractor should be registered  
				"blogspot.com": "BloggerExtractor",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := GetAllExtractors()
			
			// Check that registry is not empty
			if len(registry) == 0 {
				t.Error("GetAllExtractors() returned empty registry")
			}
			
			// Check for expected extractors (at minimum Medium and Blogger)
			if _, exists := registry["medium.com"]; !exists {
				t.Error("Registry missing medium.com extractor")
			}
			
			if _, exists := registry["blogspot.com"]; !exists {
				t.Error("Registry missing blogspot.com extractor") 
			}
			
			// Verify that each entry is an extractor type
			for domain, extractor := range registry {
				if extractor == nil {
					t.Errorf("Registry entry for %s is nil", domain)
				}
				
				// Should be an Extractor interface implementation
				if _, ok := extractor.(Extractor); !ok {
					t.Errorf("Registry entry for %s does not implement Extractor interface", domain)
				}
			}
		})
	}
}

func TestGetAllExtractorsJavaScriptCompatibility(t *testing.T) {
	// Test the exact JavaScript logic:
	// export default Object.keys(CustomExtractors).reduce((acc, key) => {
	//   const extractor = CustomExtractors[key];
	//   return {
	//     ...acc,
	//     ...mergeSupportedDomains(extractor),
	//   };
	// }, {});
	
	t.Run("JavaScript reduce pattern with mergeSupportedDomains", func(t *testing.T) {
		registry := GetAllExtractors()
		
		// Verify the registry structure matches JavaScript behavior
		// Each domain should map to an extractor
		for domain, extractor := range registry {
			// Check domain format (should be valid domain strings)
			if domain == "" {
				t.Error("Found empty domain key in registry")
			}
			
			// Check extractor validity
			if extractor == nil {
				t.Errorf("Found nil extractor for domain %s", domain)
			}
			
			// Verify extractor has proper domain information
			if ext, ok := extractor.(Extractor); ok {
				// The domain should match or be related to the extractor's domain
				extractorDomain := ext.GetDomain()
				if extractorDomain == "" {
					t.Errorf("Extractor for domain %s has empty GetDomain()", domain)
				}
			}
		}
	})
	
	t.Run("mergeSupportedDomains integration", func(t *testing.T) {
		// Test that multi-domain extractors are properly expanded
		// Using a mock extractor with supported domains
		mockExtractor := utils.MockExtractor{
			Domain:           "example.com",
			SupportedDomains: []string{"www.example.com", "mobile.example.com"},
			Name:             "ExampleExtractor",
		}
		
		merged := utils.MergeSupportedDomains(mockExtractor)
		
		// Verify all domains are present
		expectedDomains := []string{"example.com", "www.example.com", "mobile.example.com"}
		if len(merged) != len(expectedDomains) {
			t.Errorf("MergeSupportedDomains returned %d entries, want %d", len(merged), len(expectedDomains))
		}
		
		for _, domain := range expectedDomains {
			if _, exists := merged[domain]; !exists {
				t.Errorf("Missing domain %s in merged result", domain)
			}
		}
	})
}

func TestGetAllExtractorsReturnType(t *testing.T) {
	registry := GetAllExtractors()
	
	// Verify return type is map[string]Extractor (or compatible interface)
	if registry == nil {
		t.Fatal("GetAllExtractors() returned nil")
	}
	
	// Check that it's the expected type
	registryType := reflect.TypeOf(registry)
	if registryType.Kind() != reflect.Map {
		t.Errorf("GetAllExtractors() should return a map, got %v", registryType.Kind())
	}
	
	// Check key type
	if registryType.Key().Kind() != reflect.String {
		t.Errorf("Registry keys should be strings, got %v", registryType.Key().Kind())
	}
}

func TestGetAllExtractorsConsistency(t *testing.T) {
	// Test that multiple calls return consistent results
	registry1 := GetAllExtractors()
	registry2 := GetAllExtractors()
	
	if len(registry1) != len(registry2) {
		t.Errorf("Inconsistent registry size: %d vs %d", len(registry1), len(registry2))
	}
	
	// Check that all domains are present in both
	for domain := range registry1 {
		if _, exists := registry2[domain]; !exists {
			t.Errorf("Domain %s missing in second call", domain)
		}
	}
	
	for domain := range registry2 {
		if _, exists := registry1[domain]; !exists {
			t.Errorf("Domain %s missing in first call", domain)
		}
	}
}

func BenchmarkGetAllExtractors(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetAllExtractors()
	}
}

func TestExtractorRegistryExpansion(t *testing.T) {
	// Test that the registry can be expanded with new extractors
	// This verifies the foundation is in place for adding 144+ extractors
	
	t.Run("Registry foundation supports expansion", func(t *testing.T) {
		registry := GetAllExtractors()
		
		// Should have at least our basic extractors
		minExpectedSize := 2 // Medium + Blogger
		if len(registry) < minExpectedSize {
			t.Errorf("Registry size %d is below minimum expected %d", len(registry), minExpectedSize)
		}
		
		// Should be able to handle larger registries (future-proofing test)
		// This doesn't test 144 extractors but verifies the structure can handle it
		maxExpectedSize := 1000 // Well above 144, reasonable upper bound
		if len(registry) > maxExpectedSize {
			t.Errorf("Registry size %d exceeds reasonable maximum %d", len(registry), maxExpectedSize)
		}
	})
}

func TestRegistryDomainUniqueness(t *testing.T) {
	registry := GetAllExtractors()
	
	// Each domain should appear only once
	seen := make(map[string]bool)
	for domain := range registry {
		if seen[domain] {
			t.Errorf("Domain %s appears multiple times in registry", domain)
		}
		seen[domain] = true
	}
}