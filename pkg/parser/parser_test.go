package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BumpyClock/hermes/pkg/parser"
)

func TestParseURL(t *testing.T) {
	p := parser.New()

	// Test with invalid URL
	result, err := p.Parse("not-a-url", &parser.ParserOptions{})
	require.NoError(t, err)
	assert.True(t, result.IsError())
	assert.Contains(t, result.Message, "valid URL")

	// Test with malformed URL
	result, err = p.Parse("foo.com", &parser.ParserOptions{})
	require.NoError(t, err)
	assert.True(t, result.IsError())
	assert.Contains(t, result.Message, "valid URL")

	// Test with valid URL structure - may fail due to HTTP issues, which is expected
	result, err = p.Parse("https://example.com/article", &parser.ParserOptions{})
	if err != nil {
		// HTTP errors are acceptable for this test - we're testing URL parsing, not HTTP fetching
		t.Logf("HTTP fetch failed as expected: %v", err)
		return
	}
	// If it succeeds (unlikely), verify the basic structure
	if !result.IsError() {
		assert.Equal(t, "https://example.com/article", result.URL)
		assert.Equal(t, "example.com", result.Domain)
	}
}

func TestParseHTML(t *testing.T) {
	p := parser.New()

	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Test Article</title>
		<meta property="og:title" content="Test Article">
		<meta name="author" content="Test Author">
	</head>
	<body>
		<article>
			<h1>Test Article</h1>
			<p>This is test content.</p>
		</article>
	</body>
	</html>
	`

	result, err := p.ParseHTML(html, "https://example.com/article", &parser.ParserOptions{})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError())
	assert.Equal(t, "Test Article", result.Title)
	assert.Contains(t, result.Content, "This is test content")
	assert.Equal(t, "https://example.com/article", result.URL)
	assert.Equal(t, "example.com", result.Domain)
}

func TestParseHTMLWithoutTitle(t *testing.T) {
	p := parser.New()

	html := `
	<!DOCTYPE html>
	<html>
	<body>
		<h1>Header Title</h1>
		<p>Some content here.</p>
	</body>
	</html>
	`

	result, err := p.ParseHTML(html, "https://example.com/article", &parser.ParserOptions{})
	require.NoError(t, err)
	assert.Equal(t, "Header Title", result.Title)
	assert.Contains(t, result.Content, "Some content here")
}

func TestParserOptions(t *testing.T) {
	// Test default options
	p := parser.New()
	assert.NotNil(t, p)

	// Test custom options
	opts := parser.ParserOptions{
		FetchAllPages: false,
		Fallback:      false,
		ContentType:   "markdown",
		Headers:       map[string]string{"User-Agent": "test"},
	}
	p = parser.New(&opts)
	assert.NotNil(t, p)
}

// Test fixture loading preparation
func TestFixtureDirectory(t *testing.T) {
	fixturesDir := "../../internal/fixtures"

	// Check if fixtures directory exists (may not exist yet)
	_, err := os.Stat(fixturesDir)
	if os.IsNotExist(err) {
		t.Skip("Fixtures directory not yet created")
		return
	}

	require.NoError(t, err)

	// Count potential fixture files
	var count int
	err = filepath.Walk(fixturesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".html" {
			count++
		}
		return nil
	})

	require.NoError(t, err)
	t.Logf("Found %d HTML fixture files", count)
}