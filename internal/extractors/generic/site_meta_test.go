// ABOUTME: Regression tests ensuring site-level metadata extraction honors normalized meta tags
// ABOUTME: Verifies site name/title/image still resolve after meta tag normalization to name/value

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

// helper to build a goquery document from HTML string.
func newDoc(html string, t *testing.T) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse html: %v", err)
	}
	return doc
}

func TestGenericSiteMetadataWithNormalizedMetaTags(t *testing.T) {
	html := `
	<!doctype html>
	<html>
	<head>
		<meta name="og:site_name" value="The Verge">
		<meta name="og:title" value="The Verge">
		<meta name="og:image" value="https://platform.theverge.com/wp-content/uploads/sites/2/2024/10/the_verge_social_share.png">
	</head>
	<body></body>
	</html>`

	doc := newDoc(html, t)
	selection := doc.Selection

	siteName := (&GenericSiteNameExtractor{}).Extract(selection, "https://www.theverge.com", nil)
	assert.Equal(t, "The Verge", siteName, "site_name should be read from normalized name/value meta tag")

	siteTitle := (&GenericSiteTitleExtractor{}).Extract(selection, "https://www.theverge.com", nil)
	assert.Equal(t, "The Verge", siteTitle, "site_title should be read from normalized name/value meta tag")

	siteImage := (&GenericSiteImageExtractor{}).Extract(selection, "https://www.theverge.com", nil)
	assert.Equal(t, "https://platform.theverge.com/wp-content/uploads/sites/2/2024/10/the_verge_social_share.png", siteImage, "site_image should be read from normalized name/value meta tag")
}
