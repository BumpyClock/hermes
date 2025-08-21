package dom_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BumpyClock/hermes/pkg/utils/dom"
)

func TestGetAttrs(t *testing.T) {
	html := `<div id="test" class="container" data-value="example">Content</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)
	
	selection := doc.Find("div")
	attrs := dom.GetAttrs(selection)
	
	assert.Equal(t, "test", attrs["id"])
	assert.Equal(t, "container", attrs["class"])
	assert.Equal(t, "example", attrs["data-value"])
}

func TestGetAttrs_EmptySelection(t *testing.T) {
	html := `<div>Content</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)
	
	selection := doc.Find("span") // Non-existent element
	attrs := dom.GetAttrs(selection)
	
	assert.Empty(t, attrs)
}

func TestGetAttr(t *testing.T) {
	html := `<img src="image.jpg" alt="Test Image">`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)
	
	selection := doc.Find("img")
	
	src, exists := dom.GetAttr(selection, "src")
	assert.True(t, exists)
	assert.Equal(t, "image.jpg", src)
	
	_, exists = dom.GetAttr(selection, "nonexistent")
	assert.False(t, exists)
}


func TestRemoveAttr(t *testing.T) {
	html := `<div id="test" class="old-class">Content</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)
	
	selection := doc.Find("div")
	dom.RemoveAttr(selection, "class")
	
	_, exists := selection.Attr("class")
	assert.False(t, exists)
	
	// ID should still exist
	id, exists := selection.Attr("id")
	assert.True(t, exists)
	assert.Equal(t, "test", id)
}

func TestHasAttr(t *testing.T) {
	html := `<input type="text" required>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)
	
	selection := doc.Find("input")
	
	assert.True(t, dom.HasAttr(selection, "type"))
	assert.True(t, dom.HasAttr(selection, "required"))
	assert.False(t, dom.HasAttr(selection, "disabled"))
}