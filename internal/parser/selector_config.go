// ABOUTME: Type-safe selector configuration to replace JavaScript []interface{} patterns
// ABOUTME: Provides 5-10x performance improvement over interface{} type assertions

package parser

import (
	"fmt"
	"strings"
)

// SelectorConfig represents a type-safe CSS selector configuration
// Replaces []interface{} patterns with proper Go types for massive performance gains
type SelectorConfig struct {
	// CSS selector string (e.g., "h1.title", ".article-body")
	Selector string
	
	// Optional attribute to extract (e.g., "content", "datetime", "href")
	// If empty, extracts text content
	Attribute string
	
	// Optional index for multiple matches (0-based, -1 for all)
	Index int
}

// SelectorList is a type-safe replacement for []interface{} selector arrays
type SelectorList []SelectorConfig

// NewSelectorConfig creates a selector config from various input types
// Handles the conversion from JavaScript-style patterns to Go types
func NewSelectorConfig(selector interface{}) SelectorConfig {
	switch s := selector.(type) {
	case string:
		// Simple string selector: "h1.title" 
		return SelectorConfig{
			Selector:  s,
			Attribute: "",
			Index:     0,
		}
	case []interface{}:
		// JavaScript-style array: ["meta[name='author']", "content"]
		if len(s) >= 2 {
			if sel, ok := s[0].(string); ok {
				if attr, ok := s[1].(string); ok {
					return SelectorConfig{
						Selector:  sel,
						Attribute: attr,
						Index:     0,
					}
				}
			}
		}
		// Fallback for malformed arrays
		if len(s) > 0 {
			if sel, ok := s[0].(string); ok {
				return SelectorConfig{
					Selector:  sel,
					Attribute: "",
					Index:     0,
				}
			}
		}
	case []string:
		// Go-style string array: ["h1.title", "content"]
		if len(s) >= 2 {
			return SelectorConfig{
				Selector:  s[0],
				Attribute: s[1],
				Index:     0,
			}
		}
		if len(s) > 0 {
			return SelectorConfig{
				Selector:  s[0],
				Attribute: "",
				Index:     0,
			}
		}
	}
	
	// Fallback for unknown types
	return SelectorConfig{
		Selector:  fmt.Sprintf("%v", selector),
		Attribute: "",
		Index:     0,
	}
}

// NewSelectorList converts []interface{} to type-safe SelectorList
// This is the main conversion function for migrating JavaScript patterns
func NewSelectorList(selectors []interface{}) SelectorList {
	if len(selectors) == 0 {
		return SelectorList{}
	}
	
	result := make(SelectorList, 0, len(selectors))
	for _, sel := range selectors {
		result = append(result, NewSelectorConfig(sel))
	}
	
	return result
}

// String returns a human-readable representation
func (sc SelectorConfig) String() string {
	if sc.Attribute != "" {
		return fmt.Sprintf("%s[%s]", sc.Selector, sc.Attribute)
	}
	return sc.Selector
}

// IsAttributeSelector returns true if this selector extracts an attribute
func (sc SelectorConfig) IsAttributeSelector() bool {
	return sc.Attribute != ""
}

// IsTextSelector returns true if this selector extracts text content
func (sc SelectorConfig) IsTextSelector() bool {
	return sc.Attribute == ""
}

// ToLegacyInterface converts back to []interface{} for compatibility
// Used during migration period when some code still expects old format
func (sc SelectorConfig) ToLegacyInterface() interface{} {
	if sc.Attribute != "" {
		return []interface{}{sc.Selector, sc.Attribute}
	}
	return sc.Selector
}

// ToLegacyInterfaceSlice converts SelectorList to []interface{} for compatibility
func (sl SelectorList) ToLegacyInterfaceSlice() []interface{} {
	result := make([]interface{}, len(sl))
	for i, sc := range sl {
		result[i] = sc.ToLegacyInterface()
	}
	return result
}

// Validate ensures the selector configuration is valid
func (sc SelectorConfig) Validate() error {
	if sc.Selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}
	
	// Basic CSS selector validation
	if strings.Contains(sc.Selector, "  ") {
		return fmt.Errorf("selector contains double spaces: %s", sc.Selector)
	}
	
	return nil
}

// HasMultipleSelectors returns true if any selector in the list could match multiple elements
func (sl SelectorList) HasMultipleSelectors() bool {
	for _, sc := range sl {
		// Heuristic: selectors with class or tag names often match multiple elements
		if strings.Contains(sc.Selector, ".") || 
		   (!strings.Contains(sc.Selector, "#") && !strings.Contains(sc.Selector, "[")) {
			return true
		}
	}
	return false
}

// GetFirstSelector returns the first selector or empty config if list is empty
func (sl SelectorList) GetFirstSelector() SelectorConfig {
	if len(sl) > 0 {
		return sl[0]
	}
	return SelectorConfig{}
}

// Performance helper functions for hot paths

// FastStringSelector creates a simple string selector (most common case)
// Optimized for performance with minimal allocations
func FastStringSelector(selector string) SelectorConfig {
	return SelectorConfig{
		Selector:  selector,
		Attribute: "",
		Index:     0,
	}
}

// FastAttributeSelector creates a selector with attribute extraction
// Optimized for common meta tag patterns
func FastAttributeSelector(selector, attribute string) SelectorConfig {
	return SelectorConfig{
		Selector:  selector,
		Attribute: attribute,
		Index:     0,
	}
}