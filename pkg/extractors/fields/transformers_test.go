// ABOUTME: Test suite for field transformers
// ABOUTME: Validates string normalization, URL resolution, date parsing, and array transformation functionality

package fields

import (
	"reflect"
	"testing"
	"time"
)

func TestStringTransformer(t *testing.T) {
	transformer := NewStringTransformer()
	
	testCases := []struct {
		input    interface{}
		expected interface{}
	}{
		{"  hello world  ", "hello world"},
		{"hello\n\nworld", "hello world"},
		{"hello\t\tworld", "hello world"},
		{"  multiple   spaces  ", "multiple spaces"},
		{"", ""},
		{123, 123}, // Non-string should pass through
	}
	
	for _, tc := range testCases {
		result := transformer.Transform(tc.input)
		if result != tc.expected {
			t.Errorf("StringTransformer.Transform(%v) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
	
	if transformer.TargetType() != "string" {
		t.Errorf("Expected target type 'string', got %s", transformer.TargetType())
	}
}

func TestURLTransformer(t *testing.T) {
	baseURL := "https://example.com/articles/"
	transformer := NewURLTransformer(baseURL)
	
	testCases := []struct {
		input    interface{}
		expected string
	}{
		{"../other-article", "https://example.com/other-article"},
		{"./related", "https://example.com/articles/related"},
		{"https://other.com/page", "https://other.com/page"},
		{"", ""},
		{"/absolute/path", "https://example.com/absolute/path"},
	}
	
	for _, tc := range testCases {
		result := transformer.Transform(tc.input)
		if result != tc.expected {
			t.Errorf("URLTransformer.Transform(%v) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
	
	// Test URL normalization (tracking parameter removal)
	trackingURL := "https://example.com/page?utm_source=test&param=value&fbclid=123"
	result := transformer.Transform(trackingURL)
	if !contains(result.(string), "param=value") || contains(result.(string), "utm_source") {
		t.Errorf("URL normalization failed: %s", result)
	}
}

func TestDateTransformer(t *testing.T) {
	transformer := NewDateTransformer()
	
	testCases := []struct {
		input    interface{}
		expected bool // Whether it should parse to time.Time
	}{
		{"2023-12-25T10:30:00Z", true},
		{"December 25, 2023", true},
		{"2023/12/25", true},
		{"invalid-date", false},
		{"", false},
		{time.Now(), true}, // time.Time should pass through
	}
	
	for _, tc := range testCases {
		result := transformer.Transform(tc.input)
		_, isTime := result.(time.Time)
		
		if tc.expected && !isTime {
			t.Errorf("DateTransformer.Transform(%v) should parse to time.Time", tc.input)
		} else if !tc.expected && isTime && tc.input != result {
			t.Errorf("DateTransformer.Transform(%v) should not parse to time.Time", tc.input)
		}
	}
}

func TestArrayTransformer(t *testing.T) {
	stringTransformer := NewStringTransformer()
	arrayTransformer := NewArrayTransformer(stringTransformer)
	
	testCases := []struct {
		input    interface{}
		expected []interface{}
	}{
		{
			[]string{"  hello  ", "  world  "},
			[]interface{}{"hello", "world"},
		},
		{
			"tag1, tag2, tag3",
			[]interface{}{"tag1", "tag2", "tag3"},
		},
		{
			[]interface{}{"test", 123, "  spaces  "},
			[]interface{}{"test", 123, "spaces"},
		},
	}
	
	for _, tc := range testCases {
		result := arrayTransformer.Transform(tc.input)
		resultArray, ok := result.([]interface{})
		if !ok {
			t.Errorf("ArrayTransformer.Transform(%v) should return []interface{}", tc.input)
			continue
		}
		
		if !reflect.DeepEqual(resultArray, tc.expected) {
			t.Errorf("ArrayTransformer.Transform(%v) = %v, expected %v", tc.input, resultArray, tc.expected)
		}
	}
	
	// Test deduplication
	arrayTransformer.SetDeduplication(true)
	duplicateInput := []string{"test", "test", "unique"}
	result := arrayTransformer.Transform(duplicateInput)
	resultArray := result.([]interface{})
	if len(resultArray) != 2 {
		t.Errorf("Deduplication failed: expected 2 items, got %d", len(resultArray))
	}
	
	// Test max items
	arrayTransformer.SetMaxItems(1)
	result = arrayTransformer.Transform([]string{"first", "second", "third"})
	resultArray = result.([]interface{})
	if len(resultArray) != 1 {
		t.Errorf("Max items constraint failed: expected 1 item, got %d", len(resultArray))
	}
}

func TestJSONTransformer(t *testing.T) {
	transformer := NewJSONTransformer()
	transformer.AddFieldMapping("title", NewStringTransformer())
	transformer.AddFieldMapping("url", NewURLTransformer("https://example.com"))
	
	input := map[string]interface{}{
		"title":       "  Test Article  ",
		"url":         "/article/123",
		"description": "unchanged",
	}
	
	result := transformer.Transform(input)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("JSONTransformer should return map[string]interface{}")
	}
	
	if resultMap["title"] != "Test Article" {
		t.Errorf("Expected title to be trimmed, got %v", resultMap["title"])
	}
	
	expectedURL := "https://example.com/article/123"
	if resultMap["url"] != expectedURL {
		t.Errorf("Expected URL to be resolved, got %v", resultMap["url"])
	}
	
	if resultMap["description"] != "unchanged" {
		t.Errorf("Unmapped field should remain unchanged, got %v", resultMap["description"])
	}
}

func TestChainTransformer(t *testing.T) {
	// Chain string transformer and array transformer
	stringTransformer := NewStringTransformer()
	arrayTransformer := NewArrayTransformer(nil)
	
	chainTransformer := NewChainTransformer(stringTransformer, arrayTransformer)
	
	// This should first normalize the string, then convert to array
	input := "  tag1, tag2, tag3  "
	result := chainTransformer.Transform(input)
	
	resultArray, ok := result.([]interface{})
	if !ok {
		t.Fatal("ChainTransformer should return []interface{} from string input")
	}
	
	if len(resultArray) != 3 {
		t.Errorf("Expected 3 items in array, got %d", len(resultArray))
	}
	
	// Check that the string was normalized before array conversion
	if resultArray[0] != "tag1" || resultArray[1] != "tag2" || resultArray[2] != "tag3" {
		t.Errorf("Array items not properly trimmed: %v", resultArray)
	}
}

func TestNormalizeSpaces(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"hello  world", "hello world"},
		{"  hello  world  ", " hello world "},
		{"hello\nworld", "hello world"},
		{"hello\tworld", "hello world"},
		{"hello\r\nworld", "hello world"},
		{"multiple   \t\n  spaces", "multiple spaces"},
		{"", ""},
		{"single", "single"},
	}
	
	for _, tc := range testCases {
		result := normalizeSpaces(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeSpaces(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestParseDate(t *testing.T) {
	testCases := []struct {
		input   string
		isValid bool
	}{
		{"2023-12-25T10:30:00Z", true},
		{"2023-12-25 10:30:00", true},
		{"2023-12-25", true},
		{"December 25, 2023", true},
		{"Dec 25, 2023", true},
		{"2023/12/25", true},
		{"12/25/2023", true},
		{"25-12-2023", true},
		{"Mon, 25 Dec 2023 10:30:00 MST", true},
		{"invalid-date", false},
		{"", false},
		{"2023-13-45", false}, // Invalid month/day
	}
	
	for _, tc := range testCases {
		result, err := parseDate(tc.input)
		isValid := err == nil
		
		if isValid != tc.isValid {
			t.Errorf("parseDate(%q) validity = %v, expected %v (error: %v)", tc.input, isValid, tc.isValid, err)
		}
		
		if tc.isValid && result.IsZero() {
			t.Errorf("parseDate(%q) returned zero time for valid date", tc.input)
		}
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || (len(s) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}