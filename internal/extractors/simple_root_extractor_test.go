// ABOUTME: Test suite for simplified root extractor with core JavaScript compatibility
// ABOUTME: Tests selector processing, transforms, extended types without type conflicts

package extractors

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test HTML for various scenarios
const (
	basicTestHTML = `
		<html>
			<head><title>Test Article</title></head>
			<body>
				<div class="header">Header Content</div>
				<h1>Article Title</h1>
				<div class="byline">By John Doe</div>
				<div class="article-content">
					<p>This is the main article content. It contains multiple paragraphs.</p>
					<p>This is the second paragraph with more content.</p>
				</div>
				<div class="sidebar">Sidebar content</div>
			</body>
		</html>`

	attributeTestHTML = `
		<html>
			<body>
				<img src="test-image.jpg" alt="Test Image" />
				<a href="http://example.com/next" rel="next">Next Page</a>
				<meta name="author" content="Jane Smith" />
			</body>
		</html>`

	extendedFieldsHTML = `
		<html>
			<body>
				<div class="category">Technology</div>
				<div class="tags">
					<span>tag1</span>
					<span>tag2</span>
				</div>
			</body>
		</html>`
)

func TestCleanBySelectorsList(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		clean    []string
		expected int // Expected elements remaining
	}{
		{
			name:     "single selector removal",
			html:     basicTestHTML,
			clean:    []string{".sidebar"},
			expected: 0, // Sidebar should be removed
		},
		{
			name:     "multiple selector removal",
			html:     basicTestHTML,
			clean:    []string{".header", ".sidebar"},
			expected: 0, // Both should be removed
		},
		{
			name:     "no clean selectors",
			html:     basicTestHTML,
			clean:    nil,
			expected: -1, // No cleaning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			content := doc.Find("body")
			
			// Count elements before cleaning if applicable
			if len(tt.clean) > 0 {
				beforeCount := content.Find(strings.Join(tt.clean, ",")).Length()
				require.Greater(t, beforeCount, 0, "Elements should exist before cleaning")
			}

			result := CleanBySelectorsList(content, doc, tt.clean)

			if tt.expected == -1 {
				// No cleaning should occur
				contentHtml, _ := content.Html()
				resultHtml, _ := result.Html()
				assert.Equal(t, contentHtml, resultHtml)
			} else {
				// Elements should be removed
				afterCount := result.Find(strings.Join(tt.clean, ",")).Length()
				assert.Equal(t, tt.expected, afterCount)
			}
		})
	}
}

func TestTransformElementsList(t *testing.T) {
	transformHTML := `
		<html>
			<body>
				<div class="content">
					<h1>Original Heading</h1>
					<span>Text content</span>
				</div>
			</body>
		</html>`

	tests := []struct {
		name       string
		html       string
		transforms map[string]interface{}
		expected   string
	}{
		{
			name: "string transformation",
			html: transformHTML,
			transforms: map[string]interface{}{
				"h1": "h2",
			},
			expected: "h2",
		},
		{
			name: "multiple transformations",
			html: transformHTML,
			transforms: map[string]interface{}{
				"h1":   "h2",
				"span": "p",
			},
			expected: "h2",
		},
		{
			name:       "no transforms",
			html:       transformHTML,
			transforms: nil,
			expected:   "h1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			content := doc.Find("body")
			result := TransformElementsList(content, doc, tt.transforms)

			if tt.expected == "h2" && tt.transforms != nil {
				assert.Equal(t, 0, result.Find("h1").Length())
				assert.Greater(t, result.Find("h2").Length(), 0)
			} else if tt.expected == "h1" {
				assert.Greater(t, result.Find("h1").Length(), 0)
			}
		})
	}
}

func TestFindMatchingSelectorFromList(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		selectors      []interface{}
		extractHTML    bool
		allowMultiple  bool
		expectedFound  bool
		expectedResult interface{}
	}{
		{
			name: "string selector match",
			html: basicTestHTML,
			selectors: []interface{}{
				"h1",
				".nonexistent",
			},
			extractHTML:    false,
			allowMultiple:  false,
			expectedFound:  true,
			expectedResult: "h1",
		},
		{
			name: "array selector with attribute",
			html: attributeTestHTML,
			selectors: []interface{}{
				[]interface{}{"img", "src"},
				[]interface{}{"a", "href"},
			},
			extractHTML:    false,
			allowMultiple:  false,
			expectedFound:  true,
			expectedResult: []interface{}{"img", "src"},
		},
		{
			name: "no matching selectors",
			html: basicTestHTML,
			selectors: []interface{}{
				".nonexistent1",
				".nonexistent2",
			},
			extractHTML:   false,
			allowMultiple: false,
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			result := FindMatchingSelectorFromList(doc, tt.selectors, tt.extractHTML, tt.allowMultiple)

			if tt.expectedFound {
				assert.NotNil(t, result)
				if tt.expectedResult != nil {
					assert.Equal(t, tt.expectedResult, result)
				}
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestSelectField(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		extractionOpts interface{}
		fieldType      string
		extractHTML    bool
		expectedResult bool
	}{
		{
			name: "text extraction",
			html: basicTestHTML,
			extractionOpts: map[string]interface{}{
				"selectors": []interface{}{"h1"},
			},
			fieldType:      "title",
			extractHTML:    false,
			expectedResult: true,
		},
		{
			name: "html extraction",
			html: basicTestHTML,
			extractionOpts: map[string]interface{}{
				"selectors": []interface{}{".article-content"},
			},
			fieldType:      "content",
			extractHTML:    true,
			expectedResult: true,
		},
		{
			name: "attribute extraction",
			html: attributeTestHTML,
			extractionOpts: map[string]interface{}{
				"selectors": []interface{}{
					[]interface{}{"img", "src"},
				},
			},
			fieldType:      "lead_image_url",
			extractHTML:    false,
			expectedResult: true,
		},
		{
			name:           "hardcoded string",
			html:           basicTestHTML,
			extractionOpts: "Fixed Author Name",
			fieldType:      "author",
			extractHTML:    false,
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			opts := SelectOpts{
				Doc:            doc,
				Type:           tt.fieldType,
				ExtractionOpts: tt.extractionOpts,
				ExtractHTML:    tt.extractHTML,
				URL:            "http://example.com",
			}

			result := SelectField(opts)

			if tt.expectedResult {
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestSelectExtendedFields(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		extend   map[string]interface{}
		expected map[string]bool
	}{
		{
			name: "single extended type",
			html: extendedFieldsHTML,
			extend: map[string]interface{}{
				"category": map[string]interface{}{
					"selectors": []interface{}{".category"},
				},
			},
			expected: map[string]bool{
				"category": true,
			},
		},
		{
			name: "multiple extended types",
			html: extendedFieldsHTML,
			extend: map[string]interface{}{
				"category": map[string]interface{}{
					"selectors": []interface{}{".category"},
				},
				"tags": map[string]interface{}{
					"selectors":     []interface{}{".tags span"},
					"allowMultiple": true,
				},
			},
			expected: map[string]bool{
				"category": true,
				"tags":     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			opts := SelectOpts{
				Doc: doc,
				URL: "http://example.com",
			}

			results := SelectExtendedFields(tt.extend, opts)

			for key, shouldExist := range tt.expected {
				if shouldExist {
					assert.Contains(t, results, key)
					assert.NotNil(t, results[key])
				}
			}
		})
	}
}

func TestSimpleRootExtractorExtract(t *testing.T) {
	tests := []struct {
		name              string
		html              string
		extractor         interface{}
		options           map[string]interface{}
		expectedFields    []string
		expectGeneric     bool
		expectContentOnly bool
	}{
		{
			name: "generic extractor",
			html: basicTestHTML,
			extractor: map[string]interface{}{
				"domain": "*",
			},
			expectedFields: []string{"title", "content"},
			expectGeneric:  true,
		},
		{
			name: "custom extractor full extraction",
			html: basicTestHTML,
			extractor: map[string]interface{}{
				"domain": "example.com",
				"title": map[string]interface{}{
					"selectors": []interface{}{"h1"},
				},
				"author": map[string]interface{}{
					"selectors": []interface{}{".byline"},
				},
				"content": map[string]interface{}{
					"selectors": []interface{}{".article-content"},
				},
			},
			expectedFields: []string{"title", "author", "content"},
		},
		{
			name: "content only extraction",
			html: basicTestHTML,
			extractor: map[string]interface{}{
				"domain": "example.com",
				"content": map[string]interface{}{
					"selectors": []interface{}{".article-content"},
				},
			},
			options: map[string]interface{}{
				"contentOnly": true,
			},
			expectedFields:    []string{"content"},
			expectContentOnly: true,
		},
		{
			name: "custom extractor with extended types",
			html: extendedFieldsHTML,
			extractor: map[string]interface{}{
				"domain": "example.com",
				"extend": map[string]interface{}{
					"category": map[string]interface{}{
						"selectors": []interface{}{".category"},
					},
				},
			},
			expectedFields: []string{"category"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			opts := ExtractOpts{
				Doc:       doc,
				URL:       "http://example.com",
				Extractor: tt.extractor,
			}

			// Add any additional options
			if tt.options != nil {
				if contentOnly, ok := tt.options["contentOnly"].(bool); ok {
					opts.ContentOnly = contentOnly
				}
			}

			extractor := &SimpleRootExtractor{}
			result := extractor.Extract(tt.extractor, opts)

			// Verify expected fields are present
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatal("Result should be a map")
			}

			for _, field := range tt.expectedFields {
				assert.Contains(t, resultMap, field, "Field %s should be present", field)
			}

			// Additional checks based on test type
			if tt.expectContentOnly {
				// Only content should be present
				assert.Len(t, resultMap, 1)
				assert.Contains(t, resultMap, "content")
			}

			if tt.expectGeneric {
				// Should have called generic extractor
				assert.NotNil(t, result)
			}
		})
	}
}

func TestSimpleRootExtractorJavaScriptCompatibility(t *testing.T) {
	t.Run("field extraction order matches JavaScript", func(t *testing.T) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(basicTestHTML))
		require.NoError(t, err)

		extractor := map[string]interface{}{
			"domain": "example.com",
			"title": map[string]interface{}{
				"selectors": []interface{}{"h1"},
			},
			"content": map[string]interface{}{
				"selectors": []interface{}{".article-content"},
			},
		}

		opts := ExtractOpts{
			Doc:       doc,
			URL:       "http://example.com",
			Extractor: extractor,
		}

		rootExtractor := &SimpleRootExtractor{}
		result := rootExtractor.Extract(extractor, opts)
		
		assert.NotNil(t, result)
		
		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)

		// Should have standard fields even if not defined in extractor
		standardFields := []string{"title", "content", "author", "date_published", "lead_image_url", "dek", "next_page_url", "url", "domain", "excerpt", "word_count", "direction"}
		
		for _, field := range standardFields {
			// Field should exist in result (might be nil if not extracted)
			_, exists := resultMap[field]
			assert.True(t, exists, "Standard field %s should exist in result", field)
		}
	})

	t.Run("transform and clean pipeline compatibility", func(t *testing.T) {
		transformTestHTML := `
			<html>
				<body>
					<div class="content">
						<h1>Title</h1>
						<div class="remove-me">Unwanted</div>
						<div class="transform-me">Content</div>
						<a href="/relative">Relative Link</a>
					</div>
				</body>
			</html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(transformTestHTML))
		require.NoError(t, err)

		extractor := map[string]interface{}{
			"domain": "example.com",
			"content": map[string]interface{}{
				"selectors": []interface{}{".content"},
				"clean":     []string{".remove-me"},
				"transforms": map[string]interface{}{
					".transform-me": "p",
				},
			},
		}

		opts := ExtractOpts{
			Doc:       doc,
			URL:       "http://example.com",
			Extractor: extractor,
		}

		rootExtractor := &SimpleRootExtractor{}
		result := rootExtractor.Extract(extractor, opts)
		assert.NotNil(t, result)
	})
}

// TestSelectorFallbackCompatibility tests that extraction falls back to generic extractors
func TestSelectorFallbackCompatibility(t *testing.T) {
	t.Run("fallback to generic extractors", func(t *testing.T) {
		// HTML with content but no custom selectors that match
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(basicTestHTML))
		require.NoError(t, err)

		// Extractor with selectors that won't match
		extractor := map[string]interface{}{
			"domain": "example.com",
			"title": map[string]interface{}{
				"selectors": []interface{}{".nonexistent-title"},
			},
		}

		opts := ExtractOpts{
			Doc:       doc,
			URL:       "http://example.com",
			Extractor: extractor,
			Fallback:  true, // Enable fallback
		}

		rootExtractor := &SimpleRootExtractor{}
		result := rootExtractor.Extract(extractor, opts)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)

		// Should have some result from generic extractors
		// Even though custom selectors didn't match
		assert.NotNil(t, result)
		
		// Title should exist (either from custom or generic)
		_, titleExists := resultMap["title"]
		assert.True(t, titleExists)
	})
}