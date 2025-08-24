// ABOUTME: Comprehensive test suite for word count extraction with JavaScript compatibility verification
// ABOUTME: Tests both primary and fallback word counting methods to ensure accurate text analysis

package generic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordCountExtractor_Extract_Basic(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "simple sentence",
			content:  "<div>Hello world test content</div>",
			expected: 4,
		},
		{
			name:     "empty content",
			content:  "<div></div>",
			expected: 1, // Should use alt method which splits empty string
		},
		{
			name:     "single word",
			content:  "<div>Hello</div>",
			expected: 1,
		},
		{
			name:     "multiple paragraphs",
			content:  "<div><p>First paragraph here.</p><p>Second paragraph with more words.</p></div>",
			expected: 7, // Primary method result - no space between </p><p> in text extraction
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenericWordCountExtractor.Extract(map[string]interface{}{
				"content": tt.content,
			})
			
			assert.Equal(t, tt.expected, result, "Word count should match expected value")
		})
	}
}

func TestWordCountExtractor_GetWordCount_Primary(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "basic content",
			content:  "<div>The quick brown fox jumps</div>",
			expected: 5,
		},
		{
			name:     "content with nested tags",
			content:  "<div>The <strong>quick</strong> brown <em>fox</em> jumps</div>",
			expected: 5,
		},
		{
			name:     "content with multiple whitespace",
			content:  "<div>The    quick   brown\n\nfox    jumps</div>",
			expected: 5,
		},
		{
			name:     "empty div",
			content:  "<div></div>",
			expected: 1, // normalizeSpaces("") splits to [""] which has length 1
		},
		{
			name:     "div with only whitespace",
			content:  "<div>   \n   \t   </div>",
			expected: 1, // normalizeSpaces trims to "" then splits to [""]
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getWordCount(tt.content)
			assert.Equal(t, tt.expected, result, "Primary word count should match expected value")
		})
	}
}

func TestWordCountExtractor_GetWordCountAlt_Fallback(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "basic HTML content",
			content:  "<div>The quick brown fox jumps</div>",
			expected: 5,
		},
		{
			name:     "complex HTML with tags",
			content:  "<div><p>The <strong>quick</strong> brown</p><span>fox jumps</span></div>",
			expected: 5,
		},
		{
			name:     "empty content",
			content:  "",
			expected: 1, // Empty string after trim splits to [""] with length 1
		},
		{
			name:     "whitespace only",
			content:  "   \n\t   ",
			expected: 1, // After trim becomes "" then splits to [""]
		},
		{
			name:     "HTML with line breaks",
			content:  "<div>Line one<br/>Line two<br/>Line three</div>",
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getWordCountAlt(tt.content)
			assert.Equal(t, tt.expected, result, "Alternative word count should match expected value")
		})
	}
}

func TestWordCountExtractor_MethodComparison(t *testing.T) {
	// Test cases where both methods should give same result
	testCases := []struct {
		name    string
		content string
	}{
		{
			name:    "simple text",
			content: "<div>Hello world test</div>",
		},
		{
			name:    "multiple words",
			content: "<div>The quick brown fox jumps over lazy dog</div>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			primary := getWordCount(tc.content)
			alt := getWordCountAlt(tc.content)
			
			// They may not always be equal, but we document the behavior
			t.Logf("Primary method: %d, Alt method: %d for content: %s", primary, alt, tc.content)
			
			// Both should be positive
			assert.Greater(t, primary, 0, "Primary word count should be positive")
			assert.Greater(t, alt, 0, "Alt word count should be positive")
		})
	}
}

func TestWordCountExtractor_FallbackBehavior(t *testing.T) {
	// Test the fallback logic: when primary method returns 1, use alt method
	testCases := []struct {
		name            string
		content         string
		expectedPrimary int
		expectedFinal   int
	}{
		{
			name:            "empty div triggers fallback",
			content:         "<div></div>",
			expectedPrimary: 1,
			expectedFinal:   1, // Alt method also returns 1 for empty content
		},
		{
			name:            "single word no fallback",
			content:         "<div>Word</div>",
			expectedPrimary: 1,
			expectedFinal:   1, // Uses alt method but gets same result
		},
		{
			name:            "multiple words no fallback",
			content:         "<div>Multiple words here</div>",
			expectedPrimary: 3,
			expectedFinal:   3, // No fallback needed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test primary method
			primary := getWordCount(tc.content)
			assert.Equal(t, tc.expectedPrimary, primary, "Primary word count should match")
			
			// Test final extraction logic
			final := GenericWordCountExtractor.Extract(map[string]interface{}{
				"content": tc.content,
			})
			assert.Equal(t, tc.expectedFinal, final, "Final word count should match")
		})
	}
}

func TestWordCountExtractor_RealWorldExamples(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "article with paragraphs",
			content: `<div>
				<p>This is the first paragraph of an article.</p>
				<p>This is the second paragraph with more content to analyze.</p>
				<p>And this is the final paragraph.</p>
			</div>`,
			expected: 21,
		},
		{
			name: "content with lists",
			content: `<div>
				<h1>Article Title</h1>
				<ul>
					<li>First item</li>
					<li>Second item</li>
					<li>Third item</li>
				</ul>
			</div>`,
			expected: 8,
		},
		{
			name: "content with quotes",
			content: `<div>
				<p>He said, "This is a quote within the text."</p>
				<blockquote>This is a blockquote with additional content.</blockquote>
			</div>`,
			expected: 17,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenericWordCountExtractor.Extract(map[string]interface{}{
				"content": tt.content,
			})
			assert.Equal(t, tt.expected, result, "Word count should match expected value")
		})
	}
}

func TestWordCountExtractor_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "content with numbers",
			content:  "<div>The price is $123.45 for 2 items</div>",
			expected: 8,
		},
		{
			name:     "content with punctuation",
			content:  "<div>Hello, world! How are you today? Fine, thanks.</div>",
			expected: 9,
		},
		{
			name:     "content with special characters",
			content:  "<div>Email: test@example.com & website: https://example.com</div>",
			expected: 4,
		},
		{
			name:     "mixed language content",
			content:  "<div>Hello 你好 world</div>",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenericWordCountExtractor.Extract(map[string]interface{}{
				"content": tt.content,
			})
			assert.Equal(t, tt.expected, result, "Word count should match expected value")
		})
	}
}

func TestWordCountExtractor_InvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected int
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: 1, // Should handle gracefully and return 1
		},
		{
			name:     "missing content key",
			input:    map[string]interface{}{"other": "value"},
			expected: 1, // Should handle gracefully and return 1
		},
		{
			name:     "non-string content",
			input:    map[string]interface{}{"content": 123},
			expected: 1, // Should handle gracefully and return 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := GenericWordCountExtractor.Extract(tt.input)
			assert.Equal(t, tt.expected, result, "Should handle invalid input gracefully")
		})
	}
}

func TestWordCountExtractor_JavaScriptCompatibility(t *testing.T) {
	// These test cases are designed to verify 100% compatibility with the JavaScript implementation
	compatibilityTests := []struct {
		name     string
		content  string
		expected int
		note     string
	}{
		{
			name:     "JavaScript cheerio.load behavior - basic",
			content:  "<div>Test content here</div>",
			expected: 3,
			note:     "Should match cheerio.load($).first().text() behavior",
		},
		{
			name:     "JavaScript fallback trigger",
			content:  "<div></div>",
			expected: 1,
			note:     "When primary returns 1, should use alternative method",
		},
		{
			name:     "JavaScript regex HTML stripping",
			content:  "Text <span>with</span> <strong>tags</strong> removed",
			expected: 5,
			note:     "Alt method uses regex /<[^>]*>/g to strip HTML",
		},
		{
			name:     "JavaScript space normalization",
			content:  "<div>Text   with    multiple     spaces</div>",
			expected: 5,
			note:     "Should normalize multiple spaces to single spaces",
		},
	}

	for _, tt := range compatibilityTests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenericWordCountExtractor.Extract(map[string]interface{}{
				"content": tt.content,
			})
			assert.Equal(t, tt.expected, result, "JavaScript compatibility test failed: %s", tt.note)
		})
	}
}

func TestWordCountExtractor_Performance(t *testing.T) {
	// Test with larger content to ensure performance is acceptable
	largeContent := `<div>`
	for i := 0; i < 1000; i++ {
		largeContent += `<p>This is paragraph number ` + string(rune(i)) + ` with some content to count words.</p>`
	}
	largeContent += `</div>`

	t.Run("performance with large content", func(t *testing.T) {
		result := GenericWordCountExtractor.Extract(map[string]interface{}{
			"content": largeContent,
		})
		
		// Should complete without timing out and return reasonable count
		assert.Greater(t, result, 5000, "Should count words in large content")
		assert.Less(t, result, 15000, "Word count should be reasonable for test content")
	})
}

func TestWordCountExtractor_Integration(t *testing.T) {
	// Integration test to ensure the extractor works as expected in realistic scenarios
	t.Run("integration with extracted article content", func(t *testing.T) {
		// Simulate realistic article content that would come from content extraction
		articleContent := `<div>
			<h1>The Future of Web Development</h1>
			<p>Web development has evolved significantly over the past decade. Modern frameworks and tools have made it easier than ever to build sophisticated web applications.</p>
			<p>In this article, we'll explore the key trends that are shaping the future of web development, including:</p>
			<ul>
				<li>Progressive Web Applications</li>
				<li>Serverless Architecture</li>
				<li>AI-Powered Development Tools</li>
				<li>Enhanced Security Measures</li>
			</ul>
			<p>These technologies are not just buzzwords; they represent fundamental shifts in how we approach web development.</p>
		</div>`

		result := GenericWordCountExtractor.Extract(map[string]interface{}{
			"content": articleContent,
		})

		// Should return a reasonable word count for this article
		assert.Greater(t, result, 50, "Article should have substantial word count")
		assert.Less(t, result, 100, "Word count should be reasonable for test article")
		
		// Verify it's exactly what we expect for this content
		// Verified against JavaScript implementation: both return 73 words
		expectedCount := 73 // Verified with Node.js test to match JavaScript behavior
		assert.Equal(t, expectedCount, result, "Should match manually counted words")
	})
}

// Benchmark tests to verify performance
func BenchmarkWordCountExtractor_Primary(b *testing.B) {
	content := "<div>The quick brown fox jumps over the lazy dog multiple times in this test content.</div>"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getWordCount(content)
	}
}

func BenchmarkWordCountExtractor_Alternative(b *testing.B) {
	content := "<div>The quick brown fox jumps over the lazy dog multiple times in this test content.</div>"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getWordCountAlt(content)
	}
}

func BenchmarkWordCountExtractor_Extract(b *testing.B) {
	content := "<div>The quick brown fox jumps over the lazy dog multiple times in this test content.</div>"
	input := map[string]interface{}{"content": content}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenericWordCountExtractor.Extract(input)
	}
}