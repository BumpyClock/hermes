package generic

import (
	"testing"
)

// Test strings matching JavaScript string-direction-spec.js exactly
var (
	numberText        = "1234"
	ltrText          = "Hello, world!"  
	rtlText          = "سلام دنیا"
	ltrWithNumberText = "99 Bottles Of Bear..."
	rtlWithNumberText = "לקובע שלי 3 פינות"
	rtlMultilineText  = "שלום\nכיתה\nא'"
	bidiText         = "Hello in Farsi is سلام"
	LTRMarkTest      = "\u200e"
	RTLMarkTest      = "\u200f"
)

func TestGetDirection_TypeErrors(t *testing.T) {
	// Test error handling - matches JavaScript string-direction behavior exactly
	
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"number input", 1, "TypeError getDirection expects strings"},
		{"boolean input", false, "TypeError getDirection expects strings"},
		{"object input", map[string]string{}, "TypeError getDirection expects strings"},
		{"function input", func() {}, "TypeError getDirection expects strings"},
		{"nil input", nil, "TypeError missing argument"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetDirection(tt.input)
			if err == nil {
				t.Errorf("Expected error for input %v, got nil", tt.input)
			}
			if err.Error() != tt.expected {
				t.Errorf("Expected error '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}

func TestGetDirection_StringInputs(t *testing.T) {
	// Test string inputs - matches JavaScript test cases exactly
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"number text", numberText, "ltr"},
		{"ltr text", ltrText, "ltr"},
		{"rtl text", rtlText, "rtl"},
		{"ltr with number text", ltrWithNumberText, "ltr"},
		{"rtl with number text", rtlWithNumberText, "rtl"},
		{"rtl multiline text", rtlMultilineText, "rtl"},
		{"bidi text", bidiText, "bidi"},
		{"text with LTR mark", LTRMarkTest + ltrText, "ltr"},
		{"text with RTL mark", RTLMarkTest + ltrText, "rtl"},
		{"text with both marks", LTRMarkTest + RTLMarkTest + ltrText, "bidi"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetDirection(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s' for input '%s'", tt.expected, result, tt.input)
			}
		})
	}
}

func TestGetDirection_UnicodeBlocks(t *testing.T) {
	// Test Unicode block detection for RTL languages
	
	tests := []struct {
		name     string
		input    string
		expected string
		desc     string
	}{
		{
			name:     "Hebrew text",
			input:    "שלום עולם", // Hello world in Hebrew
			expected: "rtl",
			desc:     "Hebrew Unicode block 0590-05FF",
		},
		{
			name:     "Arabic text", 
			input:    "مرحبا بالعالم", // Hello world in Arabic
			expected: "rtl",
			desc:     "Arabic Unicode block 0600-06FF",
		},
		{
			name:     "Mixed Hebrew English",
			input:    "שלום Hello",
			expected: "bidi",
			desc:     "Hebrew + English = bidirectional",
		},
		{
			name:     "Mixed Arabic English",
			input:    "مرحبا Hello", 
			expected: "bidi",
			desc:     "Arabic + English = bidirectional",
		},
		{
			name:     "Numbers only",
			input:    "12345",
			expected: "ltr",
			desc:     "Pure numbers default to LTR",
		},
		{
			name:     "Hebrew with numbers",
			input:    "שלום 123",
			expected: "rtl", 
			desc:     "RTL text with numbers = RTL",
		},
		{
			name:     "Only punctuation",
			input:    ".,!?",
			expected: "",
			desc:     "Only non-directional chars = no direction",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetDirection(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("%s: Expected '%s', got '%s' for input '%s'", 
					tt.desc, tt.expected, result, tt.input)
			}
		})
	}
}

func TestGetDirection_EdgeCases(t *testing.T) {
	// Test edge cases and special scenarios
	
	tests := []struct {
		name     string
		input    string
		expected string
		desc     string
	}{
		{
			name:     "Only whitespace",
			input:    "   \t\n  ",
			expected: "",
			desc:     "Whitespace-only string has no direction",
		},
		{
			name:     "Only stripped characters",
			input:    "123+-?!'\"",
			expected: "ltr",
			desc:     "Numbers count as LTR when no other chars",
		},
		{
			name:     "RTL mark overrides content",
			input:    RTLMarkTest + "English text",
			expected: "rtl",
			desc:     "RTL mark overrides LTR content",
		},
		{
			name:     "LTR mark overrides RTL content",
			input:    LTRMarkTest + rtlText,
			expected: "ltr", 
			desc:     "LTR mark overrides RTL content",
		},
		{
			name:     "Both marks present",
			input:    LTRMarkTest + RTLMarkTest + "any text",
			expected: "bidi",
			desc:     "Both direction marks = bidirectional",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetDirection(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("%s: Expected '%s', got '%s'", tt.desc, tt.expected, result)
			}
		})
	}
}

func TestDirectionExtractor(t *testing.T) {
	// Test DirectionExtractor function that mimics JavaScript behavior
	
	tests := []struct {
		name     string
		title    string
		expected string
		desc     string
	}{
		{
			name:     "English title",
			title:    "Breaking News: Major Discovery",
			expected: "ltr",
			desc:     "English title should return LTR",
		},
		{
			name:     "Arabic title",
			title:    "أخبار عاجلة من الشرق الأوسط",
			expected: "rtl",
			desc:     "Arabic title should return RTL",
		},
		{
			name:     "Hebrew title", 
			title:    "חדשות חמות מישראל",
			expected: "rtl",
			desc:     "Hebrew title should return RTL",
		},
		{
			name:     "Mixed title",
			title:    "CNN: أخبار اليوم",
			expected: "bidi",
			desc:     "English + Arabic title = bidirectional",
		},
		{
			name:     "Empty title",
			title:    "",
			expected: "",
			desc:     "Empty title should return no direction",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := ExtractorParams{Title: tt.title}
			result, err := DirectionExtractor(params)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("%s: Expected '%s', got '%s'", tt.desc, tt.expected, result)
			}
		})
	}
}

func TestIsInScriptRange(t *testing.T) {
	// Test Unicode range detection function
	
	tests := []struct {
		name     string
		char     rune
		from     int
		to       int
		expected bool
		desc     string
	}{
		{
			name:     "Hebrew aleph in Hebrew range",
			char:     'א', // Hebrew aleph (U+05D0)
			from:     0x0590,
			to:       0x05FF,
			expected: true,
			desc:     "Hebrew character should be in Hebrew range",
		},
		{
			name:     "Arabic letter in Arabic range",
			char:     'ا', // Arabic alif (U+0627)
			from:     0x0600,
			to:       0x06FF,
			expected: true,
			desc:     "Arabic character should be in Arabic range",
		},
		{
			name:     "Latin letter not in Hebrew range",
			char:     'A',
			from:     0x0590,
			to:       0x05FF,
			expected: false,
			desc:     "Latin character should not be in Hebrew range",
		},
		{
			name:     "Boundary test - exactly at from",
			char:     rune(0x0590),
			from:     0x0590,
			to:       0x05FF,
			expected: false,
			desc:     "Boundary test: from bound is exclusive",
		},
		{
			name:     "Boundary test - exactly at to",
			char:     rune(0x05FF),
			from:     0x0590,
			to:       0x05FF,
			expected: false,
			desc:     "Boundary test: to bound is exclusive",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInScriptRange(tt.char, tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("%s: Expected %v, got %v", tt.desc, tt.expected, result)
			}
		})
	}
}

func TestHasDirectionCharacters(t *testing.T) {
	// Test character direction analysis function
	
	tests := []struct {
		name      string
		input     string
		direction string
		expected  bool
		desc      string
	}{
		{
			name:      "Hebrew text has RTL",
			input:     "שלום",
			direction: RTL,
			expected:  true,
			desc:      "Hebrew text should have RTL characters",
		},
		{
			name:      "English text has LTR",
			input:     "Hello",
			direction: LTR,
			expected:  true,
			desc:      "English text should have LTR characters",
		},
		{
			name:      "Numbers have LTR when no RTL",
			input:     "123",
			direction: LTR,
			expected:  true,
			desc:      "Numbers alone count as LTR",
		},
		{
			name:      "Hebrew with numbers still RTL",
			input:     "שלום 123",
			direction: RTL,
			expected:  true,
			desc:      "Hebrew text should have RTL even with numbers",
		},
		{
			name:      "Hebrew with numbers not pure LTR",
			input:     "שלום 123",
			direction: LTR,
			expected:  false,
			desc:      "Text with Hebrew should not be pure LTR",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasDirectionCharacters(tt.input, tt.direction)
			if result != tt.expected {
				t.Errorf("%s: Expected %v, got %v", tt.desc, tt.expected, result)
			}
		})
	}
}

func TestStripNonDirectionalRegex(t *testing.T) {
	// Test the regex that strips non-directional characters
	
	tests := []struct {
		name     string
		input    string
		expected string
		desc     string
	}{
		{
			name:     "Remove whitespace",
			input:    "a b c",
			expected: "abc",
			desc:     "Spaces should be removed",
		},
		{
			name:     "Remove numbers and punctuation",
			input:    "hello123!?+",
			expected: "hello",
			desc:     "Numbers and punctuation should be removed",
		},
		{
			name:     "Keep RTL characters",
			input:    "שלום 123",
			expected: "שלום",
			desc:     "RTL characters should be preserved",
		},
		{
			name:     "Remove newlines and tabs",
			input:    "hello\n\tworld",
			expected: "helloworld",
			desc:     "Newlines and tabs should be removed",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripNonDirectionalRegex.ReplaceAllString(tt.input, "")
			if result != tt.expected {
				t.Errorf("%s: Expected '%s', got '%s'", tt.desc, tt.expected, result)
			}
		})
	}
}

// JavaScript Compatibility Verification Test
func TestJavaScriptCompatibility(t *testing.T) {
	// These test cases directly match the JavaScript string-direction-spec.js
	// This ensures 100% behavioral compatibility
	
	jsTests := []struct {
		input    string
		expected string
	}{
		{"", ""},                                    // Empty string  
		{numberText, "ltr"},                        // "1234"
		{ltrText, "ltr"},                          // "Hello, world!"
		{rtlText, "rtl"},                          // "سلام دنیا"
		{ltrWithNumberText, "ltr"},                // "99 Bottles Of Bear..."
		{rtlWithNumberText, "rtl"},                // "לקובע שלי 3 פינות"  
		{rtlMultilineText, "rtl"},                 // "שלום\nכיתה\nא'"
		{bidiText, "bidi"},                        // "Hello in Farsi is سلام"
		{LTRMarkTest + ltrText, "ltr"},           // LTR mark + text
		{RTLMarkTest + ltrText, "rtl"},           // RTL mark + text
		{LTRMarkTest + RTLMarkTest + "text", "bidi"}, // Both marks
	}
	
	for i, test := range jsTests {
		t.Run(test.input, func(t *testing.T) {
			result, err := GetDirection(test.input)
			if err != nil {
				t.Errorf("Test %d: Unexpected error: %v", i, err)
			}
			if result != test.expected {
				t.Errorf("Test %d: JavaScript compatibility failed. Input: '%s', Expected: '%s', Got: '%s'", 
					i, test.input, test.expected, result)
			}
		})
	}
}

func BenchmarkGetDirection(t *testing.B) {
	// Benchmark direction detection performance
	
	testCases := []string{
		"Hello, world!",
		"سلام دنیا", 
		"Hello in Farsi is سلام",
		"שלום עולם",
		LTRMarkTest + "English text",
		RTLMarkTest + rtlText,
	}
	
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		for _, testCase := range testCases {
			GetDirection(testCase)
		}
	}
}