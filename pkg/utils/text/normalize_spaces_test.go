// ABOUTME: Tests for NormalizeSpaces function ensuring JavaScript compatibility
// ABOUTME: Tests whitespace normalization with HTML tag preservation for pre, code, textarea
package text

import (
	"testing"
)

func TestNormalizeSpaces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normalizes multiple spaces to single space",
			input:    "What   do  you    think?",
			expected: "What do you think?",
		},
		{
			name:     "trims leading and trailing whitespace",
			input:    "   What do you think?   ",
			expected: "What do you think?",
		},
		{
			name:     "normalizes tabs and newlines to single space",
			input:    "What\t\tdo\n\nyou\r\rthink?",
			expected: "What do you think?",
		},
		{
			name:     "preserves spaces within pre tags",
			input:    "<div><p>What   do  you    think?</p><pre>  What     happens to        spaces?    </pre></div>",
			expected: "<div><p>What do you think?</p><pre>  What     happens to        spaces?    </pre></div>",
		},
		{
			name:     "preserves spaces within code tags",
			input:    "<div><p>Multiple   spaces</p><code>  var x =     'test';    </code><p>More   text</p></div>",
			expected: "<div><p>Multiple spaces</p><code>  var x =     'test';    </code><p>More text</p></div>",
		},
		{
			name:     "preserves spaces within textarea tags",
			input:    "<div><p>Text   with    spaces</p><textarea>  Keep     all    spaces   here  </textarea></div>",
			expected: "<div><p>Text with spaces</p><textarea>  Keep     all    spaces   here  </textarea></div>",
		},
		{
			name:     "handles nested pre tags correctly",
			input:    "<div><p>Normal   text</p><div><pre>    Nested     spaces    </pre></div></div>",
			expected: "<div><p>Normal text</p><div><pre>    Nested     spaces    </pre></div></div>",
		},
		{
			name:     "handles multiple pre/code/textarea tags",
			input:    "<div><p>Text   1</p><pre>  pre  </pre><p>Text   2</p><code>  code  </code><p>Text   3</p><textarea>  textarea  </textarea></div>",
			expected: "<div><p>Text 1</p><pre>  pre  </pre><p>Text 2</p><code>  code  </code><p>Text 3</p><textarea>  textarea  </textarea></div>",
		},
		{
			name:     "handles self-closing and unclosed tags",
			input:    "<div><p>Text   with    spaces</p><pre>  Keep  spaces  ",
			expected: "<div><p>Text with spaces</p><pre> Keep spaces",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles only whitespace",
			input:    "   \t\n\r   ",
			expected: "",
		},
		{
			name:     "handles text without HTML",
			input:    "Simple   text   with    spaces",
			expected: "Simple text with spaces",
		},
		{
			name:     "handles mixed content with line breaks",
			input:    "\n\n      <div>\n        <p>What do you think?</p>\n      </div>\n    ",
			expected: "<div> <p>What do you think?</p> </div>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeSpaces(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeSpaces(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNormalizeSpacesJavaScriptCompatibility ensures exact JavaScript behavior
func TestNormalizeSpacesJavaScriptCompatibility(t *testing.T) {
	// Test cases directly from the JavaScript test file
	
	t.Run("JavaScript test case 1 - normalizes spaces from text", func(t *testing.T) {
		// This simulates cheerio loading and extracting text from:
		// <div><p>What do you think?</p></div>
		// The text extraction adds extra spaces/newlines that need normalization
		input := "\n        What do you think?\n      "
		expected := "What do you think?"
		result := NormalizeSpaces(input)
		
		if result != expected {
			t.Errorf("NormalizeSpaces(%q) = %q, expected %q (JavaScript compatibility test 1)", input, result, expected)
		}
	})
	
	t.Run("JavaScript test case 2 - preserves spaces in preformatted text blocks", func(t *testing.T) {
		// This is the exact HTML from the JavaScript test
		input := `<div> <p>What   do  you    think?</p> <pre>  What     happens to        spaces?    </pre> </div>`
		expected := `<div> <p>What do you think?</p> <pre>  What     happens to        spaces?    </pre> </div>`
		result := NormalizeSpaces(input)
		
		if result != expected {
			t.Errorf("NormalizeSpaces(%q) = %q, expected %q (JavaScript compatibility test 2)", input, result, expected)
		}
	})
}

// BenchmarkNormalizeSpaces measures performance of the function
func BenchmarkNormalizeSpaces(b *testing.B) {
	testHTML := "<div><p>Multiple   spaces   everywhere</p><pre>  Keep     these    spaces  </pre><p>More   text   with    spaces</p></div>"
	
	for i := 0; i < b.N; i++ {
		NormalizeSpaces(testHTML)
	}
}

// TestNormalizeSpacesRegexBehavior tests the specific regex pattern behavior
func TestNormalizeSpacesRegexBehavior(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		desc     string
	}{
		{
			name:     "JavaScript regex pattern - basic multiple spaces",
			input:    "text  with   spaces",
			expected: "text with spaces",
			desc:     "Should collapse 2+ consecutive whitespace to single space",
		},
		{
			name:     "JavaScript regex pattern - mixed whitespace types",
			input:    "text\t\twith\n\nspaces",
			expected: "text with spaces",
			desc:     "Should treat tabs, newlines as whitespace",
		},
		{
			name:     "JavaScript regex pattern - negative lookahead for pre",
			input:    "text  spaces<pre>keep  spaces</pre>more  spaces",
			expected: "text spaces<pre>keep  spaces</pre>more spaces",
			desc:     "Should not collapse spaces before </pre> tag",
		},
		{
			name:     "JavaScript regex pattern - negative lookahead for code",
			input:    "text  spaces<code>keep  spaces</code>more  spaces",
			expected: "text spaces<code>keep  spaces</code>more spaces",
			desc:     "Should not collapse spaces before </code> tag",
		},
		{
			name:     "JavaScript regex pattern - negative lookahead for textarea",
			input:    "text  spaces<textarea>keep  spaces</textarea>more  spaces",
			expected: "text spaces<textarea>keep  spaces</textarea>more spaces",
			desc:     "Should not collapse spaces before </textarea> tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeSpaces(tt.input)
			if result != tt.expected {
				t.Errorf("%s: NormalizeSpaces(%q) = %q, expected %q", tt.desc, tt.input, result, tt.expected)
			}
		})
	}
}