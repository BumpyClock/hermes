package dom_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/postlight/parser-go/pkg/utils/dom"
)

func TestSetAttr_CheerioStyleNode(t *testing.T) {
	// This test mirrors the JavaScript test for "raw cheerio node"
	// In the JS test, they create a cheerioNode with attribs property
	// In Go, we use goquery which handles the attribute setting internally
	html := `<div class="foo bar" id="baz bat">Content</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	selection := doc.Find("div")
	
	// Set the class attribute to 'foo' (matching JS test)
	result := dom.SetAttr(selection, "class", "foo")
	
	// Verify the attribute was set correctly
	class, exists := result.Attr("class")
	assert.True(t, exists)
	assert.Equal(t, "foo", class)
	
	// Verify that other attributes remain unchanged
	id, exists := result.Attr("id")
	assert.True(t, exists)
	assert.Equal(t, "baz bat", id)
}

func TestSetAttr_DOMStyleBehavior(t *testing.T) {
	// This test mirrors the JavaScript test for "raw jquery node" behavior
	// While we can't test actual DOM nodes in Go, we test the equivalent behavior
	// where attributes are set on goquery selections
	html := `<div>Content</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	selection := doc.Find("div")
	
	// Set the class attribute to 'foo' (matching JS MockDomNode test)
	result := dom.SetAttr(selection, "class", "foo")
	
	// Verify the attribute was set correctly
	class, exists := result.Attr("class")
	assert.True(t, exists)
	assert.Equal(t, "foo", class)
}

func TestSetAttr_MethodChaining(t *testing.T) {
	// Test that the function returns the selection for method chaining
	// This mirrors the JavaScript behavior where setAttr returns the node
	html := `<div>Content</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	selection := doc.Find("div")
	
	// Verify method chaining works
	result := dom.SetAttr(selection, "class", "foo")
	assert.NotNil(t, result)
	assert.Equal(t, selection, result) // Should return the same selection
}

func TestSetAttr_MultipleAttributes(t *testing.T) {
	// Test setting multiple attributes sequentially
	html := `<div>Content</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	selection := doc.Find("div")
	
	// Set multiple attributes using method chaining
	dom.SetAttr(selection, "class", "test-class")
	dom.SetAttr(selection, "id", "test-id")
	dom.SetAttr(selection, "data-value", "test-data")
	
	// Verify all attributes were set
	class, exists := selection.Attr("class")
	assert.True(t, exists)
	assert.Equal(t, "test-class", class)
	
	id, exists := selection.Attr("id")
	assert.True(t, exists)
	assert.Equal(t, "test-id", id)
	
	dataValue, exists := selection.Attr("data-value")
	assert.True(t, exists)
	assert.Equal(t, "test-data", dataValue)
}

func TestSetAttr_OverwriteExistingAttribute(t *testing.T) {
	// Test overwriting an existing attribute value
	html := `<div class="old-class" id="existing">Content</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	selection := doc.Find("div")
	
	// Overwrite the existing class
	dom.SetAttr(selection, "class", "new-class")
	
	// Verify the class was overwritten
	class, exists := selection.Attr("class")
	assert.True(t, exists)
	assert.Equal(t, "new-class", class)
	
	// Verify other attributes remain unchanged
	id, exists := selection.Attr("id")
	assert.True(t, exists)
	assert.Equal(t, "existing", id)
}

func TestSetAttr_EmptyValue(t *testing.T) {
	// Test setting an attribute to an empty value
	html := `<div>Content</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	selection := doc.Find("div")
	
	// Set attribute to empty string
	dom.SetAttr(selection, "class", "")
	
	// Verify the attribute exists but is empty
	class, exists := selection.Attr("class")
	assert.True(t, exists)
	assert.Equal(t, "", class)
}

func TestSetAttr_SpecialCharacters(t *testing.T) {
	// Test setting attributes with special characters
	html := `<div>Content</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	selection := doc.Find("div")
	
	// Test various special characters and values
	testCases := []struct {
		attr string
		val  string
	}{
		{"data-test", "value with spaces"},
		{"class", "class-with-hyphens"},
		{"id", "id_with_underscores"},
		{"data-json", `{"key": "value"}`},
		{"title", "Title with 'quotes' and \"double quotes\""},
	}
	
	for _, tc := range testCases {
		dom.SetAttr(selection, tc.attr, tc.val)
		
		value, exists := selection.Attr(tc.attr)
		assert.True(t, exists, "Attribute %s should exist", tc.attr)
		assert.Equal(t, tc.val, value, "Attribute %s should have correct value", tc.attr)
	}
}

func TestSetAttr_EmptySelection(t *testing.T) {
	// Test behavior with empty selection (no matching elements)
	html := `<div>Content</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	selection := doc.Find("span") // Non-existent element
	
	// This should not panic and should return the selection
	result := dom.SetAttr(selection, "class", "foo")
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.Length()) // Should still be empty
}