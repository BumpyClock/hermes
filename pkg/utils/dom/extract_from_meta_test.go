// ABOUTME: Test file for ExtractFromMeta function 
// ABOUTME: Validates meta tag extraction with 100% JavaScript compatibility

package dom

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test StripTags function first (dependency of ExtractFromMeta)
func TestStripTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "simple HTML tags",
			input:    "<p>hello world</p>",
			expected: "hello world",
		},
		{
			name:     "nested HTML tags",
			input:    "<div><p>hello <strong>world</strong></p></div>",
			expected: "hello world",
		},
		{
			name:     "multiple tags",
			input:    "<a href='#'>link</a> and <em>emphasis</em>",
			expected: "link and emphasis",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only HTML tags",
			input:    "<div></div>",
			expected: "<div></div>", // JavaScript behavior: if cleanText is empty, return original
		},
		{
			name:     "HTML entities",
			input:    "<p>hello &amp; world</p>",
			expected: "hello & world",
		},
		{
			name:     "malformed HTML - JavaScript behavior: if text has no html, return original",
			input:    "<invalid>text",
			expected: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body></body></html>"))
			require.NoError(t, err)
			
			result := StripTags(tt.input, doc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFromMeta(t *testing.T) {
	t.Run("extracts an arbitrary meta tag by name", func(t *testing.T) {
		html := `
		<html>
			<meta name="foo" value="bar" />
		</html>
		`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		metaNames := []string{"foo", "baz"}
		cachedNames := []string{"foo", "bat"}
		
		result := ExtractFromMeta(doc, metaNames, cachedNames, true)
		require.NotNil(t, result)
		assert.Equal(t, "bar", *result)
	})

	t.Run("returns nil if a meta name is duplicated", func(t *testing.T) {
		html := `
		<html>
			<meta name="foo" value="bar" />
			<meta name="foo" value="baz" />
		</html>
		`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		metaNames := []string{"foo", "baz"}
		cachedNames := []string{"foo", "bat"}
		
		result := ExtractFromMeta(doc, metaNames, cachedNames, true)
		assert.Nil(t, result)
	})

	t.Run("ignores duplicate meta names with empty values", func(t *testing.T) {
		html := `
		<html>
			<meta name="foo" value="bar" />
			<meta name="foo" value="" />
		</html>
		`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		metaNames := []string{"foo", "baz"}
		cachedNames := []string{"foo", "bat"}
		
		result := ExtractFromMeta(doc, metaNames, cachedNames, true)
		require.NotNil(t, result)
		assert.Equal(t, "bar", *result)
	})

	t.Run("strips HTML tags when cleanTags is true", func(t *testing.T) {
		html := `
		<html>
			<meta name="description" value="<p>hello <strong>world</strong></p>" />
		</html>
		`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		metaNames := []string{"description"}
		cachedNames := []string{"description"}
		
		result := ExtractFromMeta(doc, metaNames, cachedNames, true)
		require.NotNil(t, result)
		assert.Equal(t, "hello world", *result)
	})

	t.Run("does not strip HTML tags when cleanTags is false", func(t *testing.T) {
		html := `
		<html>
			<meta name="description" value="<p>hello <strong>world</strong></p>" />
		</html>
		`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		metaNames := []string{"description"}
		cachedNames := []string{"description"}
		
		result := ExtractFromMeta(doc, metaNames, cachedNames, false)
		require.NotNil(t, result)
		assert.Equal(t, "<p>hello <strong>world</strong></p>", *result)
	})

	t.Run("returns nil when no meta names match", func(t *testing.T) {
		html := `
		<html>
			<meta name="foo" value="bar" />
		</html>
		`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		metaNames := []string{"baz", "qux"}
		cachedNames := []string{"foo", "bat"}
		
		result := ExtractFromMeta(doc, metaNames, cachedNames, true)
		assert.Nil(t, result)
	})

	t.Run("returns nil when cachedNames is empty", func(t *testing.T) {
		html := `
		<html>
			<meta name="foo" value="bar" />
		</html>
		`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		metaNames := []string{"foo", "baz"}
		cachedNames := []string{}
		
		result := ExtractFromMeta(doc, metaNames, cachedNames, true)
		assert.Nil(t, result)
	})

	t.Run("handles meta tags with content attribute", func(t *testing.T) {
		html := `
		<html>
			<meta name="description" content="test description" />
		</html>
		`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		metaNames := []string{"description"}
		cachedNames := []string{"description"}
		
		// This test shows the JavaScript function is hardcoded to use 'value' attribute
		// So this should return nil since there's no 'value' attribute
		result := ExtractFromMeta(doc, metaNames, cachedNames, true)
		assert.Nil(t, result)
	})

	t.Run("prioritizes first matching name in metaNames order", func(t *testing.T) {
		html := `
		<html>
			<meta name="foo" value="foo-value" />
			<meta name="bar" value="bar-value" />
		</html>
		`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		metaNames := []string{"foo", "bar"}
		cachedNames := []string{"bar", "foo"} // both are available, but foo comes first in metaNames
		
		result := ExtractFromMeta(doc, metaNames, cachedNames, true)
		require.NotNil(t, result)
		assert.Equal(t, "foo-value", *result) // Returns first match in metaNames order
	})

	// Additional tests for comprehensive coverage of meta tag patterns
	t.Run("works with OpenGraph-style meta tags", func(t *testing.T) {
		// Note: JavaScript implementation only looks for name="*", not property="*"
		// So OpenGraph tags won't be found with this function as it stands
		html := `
		<html>
			<meta property="og:title" value="OpenGraph Title" />
			<meta name="og:title" value="OpenGraph Title with name" />
		</html>
		`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		metaNames := []string{"og:title"}
		cachedNames := []string{"og:title"}
		
		result := ExtractFromMeta(doc, metaNames, cachedNames, true)
		// Should find the name="og:title", not property="og:title"
		require.NotNil(t, result)
		assert.Equal(t, "OpenGraph Title with name", *result)
	})

	t.Run("multiple meta tags with different names", func(t *testing.T) {
		html := `
		<html>
			<meta name="description" value="Main description" />
			<meta name="twitter:description" value="Twitter description" />
			<meta name="og:description" value="OpenGraph description" />
		</html>
		`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		// Test prioritization - should return first match in metaNames order
		metaNames := []string{"twitter:description", "description", "og:description"}
		cachedNames := []string{"description", "twitter:description", "og:description"}
		
		result := ExtractFromMeta(doc, metaNames, cachedNames, true)
		require.NotNil(t, result)
		assert.Equal(t, "Twitter description", *result)
	})

	t.Run("edge case with special characters in values", func(t *testing.T) {
		html := `
		<html>
			<meta name="special" value="Value with &quot;quotes&quot; and &amp; symbols" />
		</html>
		`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		metaNames := []string{"special"}
		cachedNames := []string{"special"}
		
		result := ExtractFromMeta(doc, metaNames, cachedNames, true)
		require.NotNil(t, result)
		assert.Equal(t, "Value with \"quotes\" and & symbols", *result)
	})

	t.Run("performance test with large meta list", func(t *testing.T) {
		// Build HTML with many meta tags
		htmlBuilder := []string{"<html>"}
		for i := 0; i < 100; i++ {
			htmlBuilder = append(htmlBuilder, fmt.Sprintf("<meta name=\"test%d\" value=\"value%d\" />", i, i))
		}
		htmlBuilder = append(htmlBuilder, "</html>")
		html := strings.Join(htmlBuilder, "\n")

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		// Look for the 50th meta tag
		metaNames := []string{"test50", "test99", "test1"}
		cachedNames := []string{"test1", "test50", "test99"}
		
		result := ExtractFromMeta(doc, metaNames, cachedNames, true)
		require.NotNil(t, result)
		assert.Equal(t, "value50", *result) // Should find test50 first
	})
}