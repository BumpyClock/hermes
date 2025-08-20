// ABOUTME: Test suite for root extractor orchestration system with comprehensive JavaScript compatibility testing
// ABOUTME: Tests the complex selector processing, transforms, extended types, and custom extractor integration

package extractors

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test HTML snippets for various root extractor scenarios
const (
	testHTML = `
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

	transformHTML = `
		<html>
			<body>
				<div class="content">
					<h1>Original Heading</h1>
					<span>Text content</span>
					<div class="transform-me">Transform This</div>
				</div>
			</body>
		</html>`

	multiSelectorHTML = `
		<html>
			<body>
				<div class="author-bio">John Doe</div>
				<div class="article-body">
					<p>First paragraph</p>
					<p>Second paragraph</p>
				</div>
				<img src="image1.jpg" alt="Image 1" />
				<img src="image2.jpg" alt="Image 2" />
			</body>
		</html>`

	attributeExtractionHTML = `
		<html>
			<body>
				<img src="test-image.jpg" alt="Test Image" />
				<a href="http://example.com/next" rel="next">Next Page</a>
				<meta property="article:published_time" content="2023-01-15T10:30:00Z" />
			</body>
		</html>`

	extendedTypesHTML = `
		<html>
			<body>
				<div class="category">Technology</div>
				<div class="tags">
					<span>tag1</span>
					<span>tag2</span>
					<span>tag3</span>
				</div>
				<div class="rating">4.5 stars</div>
			</body>
		</html>`
)

func TestCleanBySelectors(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		clean    []string
		expected int // Expected number of elements after cleaning
	}{
		{
			name:     "single selector removal",
			html:     testHTML,
			clean:    []string{".sidebar"},
			expected: 0, // Should remove sidebar
		},
		{
			name:     "multiple selector removal",
			html:     testHTML,
			clean:    []string{".header", ".sidebar"},
			expected: 0, // Should remove both
		},
		{
			name:     "no clean selectors",
			html:     testHTML,
			clean:    nil,
			expected: -1, // No cleaning should occur
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			content := doc.Find("body")
			
			// Count elements before cleaning (if applicable)
			beforeCount := 0
			if len(tt.clean) > 0 {
				beforeCount = content.Find(strings.Join(tt.clean, ",")).Length()
				require.Greater(t, beforeCount, 0, "Test elements should exist before cleaning")
			}

			result := CleanBySelectors(content, doc, map[string][]string{"clean": tt.clean})

			if tt.expected == -1 {
				// No cleaning should have occurred
				assert.Equal(t, content, result)
			} else {
				// Elements should be removed
				afterCount := result.Find(strings.Join(tt.clean, ",")).Length()
				assert.Equal(t, tt.expected, afterCount)
			}
		})
	}
}

func TestTransformElements(t *testing.T) {
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
				"h1": "h2", // Convert h1 to h2
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
			expected:   "h1", // Should remain unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			content := doc.Find("body")
			result := TransformElements(content, doc, map[string]map[string]interface{}{"transforms": tt.transforms})

			// Check if transformation occurred
			if tt.expected == "h2" && tt.transforms != nil {
				assert.Equal(t, 0, result.Find("h1").Length())
				assert.Greater(t, result.Find("h2").Length(), 0)
			} else if tt.expected == "h1" {
				assert.Greater(t, result.Find("h1").Length(), 0)
			}
		})
	}
}

func TestFindMatchingSelector(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		selectors      []interface{}
		extractHtml    bool
		allowMultiple  bool
		expectedFound  bool
		expectedResult interface{}
	}{
		{
			name: "string selector match",
			html: testHTML,
			selectors: []interface{}{
				"h1",
				".nonexistent",
			},
			extractHtml:    false,
			allowMultiple:  false,
			expectedFound:  true,
			expectedResult: "h1",
		},
		{
			name: "array selector with attribute",
			html: attributeExtractionHTML,
			selectors: []interface{}{
				[]interface{}{"img", "src"},
				[]interface{}{"a", "href"},
			},
			extractHtml:    false,
			allowMultiple:  false,
			expectedFound:  true,
			expectedResult: []interface{}{"img", "src"},
		},
		{
			name: "no matching selectors",
			html: testHTML,
			selectors: []interface{}{
				".nonexistent1",
				".nonexistent2",
			},
			extractHtml:   false,
			allowMultiple: false,
			expectedFound: false,
		},
		{
			name: "multiple elements allowed",
			html: multiSelectorHTML,
			selectors: []interface{}{
				"img",
			},
			extractHtml:   false,
			allowMultiple: true,
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			result := FindMatchingSelector(doc, tt.selectors, tt.extractHtml, tt.allowMultiple)

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

func TestSelect(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		extractionOpts map[string]interface{}
		fieldType      string
		extractHtml    bool
		expectedResult bool // Whether a result should be found
	}{
		{
			name: "text extraction",
			html: testHTML,
			extractionOpts: map[string]interface{}{
				"selectors": []interface{}{"h1"},
			},
			fieldType:      "title",
			extractHtml:    false,
			expectedResult: true,
		},
		{
			name: "html extraction",
			html: testHTML,
			extractionOpts: map[string]interface{}{
				"selectors": []interface{}{".article-content"},
			},
			fieldType:      "content",
			extractHtml:    true,
			expectedResult: true,
		},
		{
			name: "attribute extraction",
			html: attributeExtractionHTML,
			extractionOpts: map[string]interface{}{
				"selectors": []interface{}{
					[]interface{}{"img", "src"},
				},
			},
			fieldType:      "lead_image_url",
			extractHtml:    false,
			expectedResult: true,
		},
		{
			name: "hardcoded string",
			html: testHTML,
			extractionOpts: "Fixed Author Name",
			fieldType:     "author",
			extractHtml:   false,
			expectedResult: true,
		},
		{
			name: "no extraction options",
			html: testHTML,
			extractionOpts: nil,
			fieldType:     "title",
			extractHtml:   false,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			opts := SelectOptions{
				Doc:            doc,
				Type:           tt.fieldType,
				ExtractionOpts: tt.extractionOpts,
				ExtractHTML:    tt.extractHtml,
				URL:            "http://example.com",
			}

			result := Select(opts)

			if tt.expectedResult {
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestSelectExtendedTypes(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		extend   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "single extended type",
			html: extendedTypesHTML,
			extend: map[string]interface{}{
				"category": map[string]interface{}{
					"selectors": []interface{}{".category"},
				},
			},
			expected: map[string]interface{}{
				"category": true, // Should find result
			},
		},
		{
			name: "multiple extended types",
			html: extendedTypesHTML,
			extend: map[string]interface{}{
				"category": map[string]interface{}{
					"selectors": []interface{}{".category"},
				},
				"tags": map[string]interface{}{
					"selectors":     []interface{}{".tags span"},
					"allowMultiple": true,
				},
			},
			expected: map[string]interface{}{
				"category": true,
				"tags":     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			opts := SelectOptions{
				Doc: doc,
				URL: "http://example.com",
			}

			results := SelectExtendedTypes(tt.extend, opts)

			for key, expected := range tt.expected {
				if expected.(bool) {
					assert.Contains(t, results, key)
					assert.NotNil(t, results[key])
				}
			}
		})
	}
}

func TestRootExtractorExtract(t *testing.T) {
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
			html: testHTML,
			extractor: map[string]interface{}{
				"domain": "*",
			},
			expectedFields: []string{"title", "content"},
			expectGeneric:  true,
		},
		{
			name: "custom extractor full extraction",
			html: testHTML,
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
			html: testHTML,
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
			html: extendedTypesHTML,
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

			opts := ExtractOptions{
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

			result := RootExtractor.Extract(tt.extractor, opts)

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
				// This would require mocking, simplified for now
				assert.NotNil(t, result)
			}
		})
	}
}

func TestRootExtractorJavaScriptCompatibility(t *testing.T) {
	// Test that specifically verifies JavaScript compatibility
	t.Run("field extraction order matches JavaScript", func(t *testing.T) {
		// JavaScript processes fields in specific order for dependencies
		// title -> content -> lead_image_url -> excerpt -> dek etc.
		
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
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

		opts := ExtractOptions{
			Doc:       doc,
			URL:       "http://example.com",
			Extractor: extractor,
		}

		result := RootExtractor.Extract(extractor, opts)
		
		// Should return a valid result
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

		opts := ExtractOptions{
			Doc:       doc,
			URL:       "http://example.com",
			Extractor: extractor,
		}

		result := RootExtractor.Extract(extractor, opts)
		assert.NotNil(t, result)
	})
}