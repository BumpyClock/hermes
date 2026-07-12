package dom_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BumpyClock/hermes/internal/utils/dom"
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

	selection := doc.Find("span")
	attrs := dom.GetAttrs(selection)

	assert.Empty(t, attrs)
}
