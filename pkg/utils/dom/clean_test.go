package dom_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BumpyClock/hermes/pkg/utils/dom"
)

func TestCleanAttributes(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		removed  []string // attributes that should be removed
		kept     []string // attributes that should be kept
	}{
		{
			name: "removes style and align",
			html: `<div style="color: red;" align="center" class="content" id="main">Content</div>`,
			removed: []string{"style", "align"},
			kept:    []string{"class", "id"},
		},
		{
			name: "keeps whitelisted attributes",
			html: `<img src="image.jpg" srcset="image@2x.jpg 2x" alt="Image" width="100" height="200" onclick="alert()">`,
			removed: []string{"onclick"},
			kept:    []string{"src", "srcset", "alt", "width", "height"},
		},
		{
			name: "handles links properly",
			html: `<a href="http://example.com" target="_blank" rel="noopener" onclick="track()">Link</a>`,
			removed: []string{"target", "onclick"},
			kept:    []string{"href"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			result := dom.CleanAttributes(doc)

			// Find the element we're testing (skip html, head, body)
			element := result.Find("div, img, a").First()
			if element.Length() == 0 {
				// Fallback to any element with attributes
				result.Find("*").Each(func(i int, el *goquery.Selection) {
					if i > 2 && element.Length() == 0 { // Skip html, head, body
						element = el
					}
				})
			}
			
			// Check removed attributes
			for _, attr := range tt.removed {
				_, exists := element.Attr(attr)
				assert.False(t, exists, "Attribute %s should be removed", attr)
			}

			// Check kept attributes
			for _, attr := range tt.kept {
				_, exists := element.Attr(attr)
				assert.True(t, exists, "Attribute %s should be kept", attr)
			}
		})
	}
}

func TestCleanHeaders(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		remaining int
		removed   []string
	}{
		{
			name: "removes short headers",
			html: `<html><body>
				<h2>Good Header with Substantial Content</h2>
				<h3>Hi</h3>
				<h4>A</h4>
				<h5>Another Good Header</h5>
			</body></html>`,
			remaining: 2,
			removed:   []string{"Hi", "A"},
		},
		{
			name: "removes navigation headers",
			html: `<html><body>
				<h2 class="nav-header">Navigation</h2>
				<h3 class="content-header">Article Title</h3>
				<h4 id="sidebar-title">Sidebar</h4>
			</body></html>`,
			remaining: 1,
			removed:   []string{"Navigation", "Sidebar"},
		},
		{
			name: "keeps good content headers",
			html: `<html><body>
				<h2 class="article-title">Main Article Title</h2>
				<h3 class="section-header">Section Header</h3>
				<h4>Subsection Header</h4>
			</body></html>`,
			remaining: 3,
			removed:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			result := dom.CleanHeadersWithoutTitle(doc)

			// Count remaining headers
			headers := result.Find("h2, h3, h4, h5, h6")
			assert.Equal(t, tt.remaining, headers.Length(), "Should have expected number of headers")

			// Check that removed content is gone
			for _, removedText := range tt.removed {
				found := false
				headers.Each(func(i int, h *goquery.Selection) {
					if strings.Contains(h.Text(), removedText) {
						found = true
					}
				})
				assert.False(t, found, "Removed text should not be found: %s", removedText)
			}
		})
	}
}

func TestCleanTags(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		removed   []string // text that should be removed
		kept      []string // text that should be kept
	}{
		{
			name: "removes high link density elements with low weight",
			html: `<html><body>
				<div>
					<p>Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor.</p>
					<div score="10">
						<a href="#">Link 1</a> <a href="#">Link 2</a> <a href="#">Link 3</a>
						<a href="#">Link 4</a> some more text to make it longer than 75 chars for the test condition
					</div>
				</div>
			</body></html>`,
			removed: []string{"Link 1", "Link 2"},
			kept:    []string{"Lorem ipsum"},
		},
		{
			name: "removes negative scored short content",
			html: `<html><body>
				<div class="sidebar">Short sidebar text</div>
				<div class="footer">Footer content</div>
				<div class="content">
					This is substantial content that should be preserved because it has enough
					text and doesn't match negative patterns.
				</div>
			</body></html>`,
			removed: []string{"Short sidebar", "Footer content"},
			kept:    []string{"substantial content"},
		},
		{
			name: "keeps positive scored content",
			html: `<html><body>
				<div class="article sidebar">Article content in sidebar class</div>
				<div class="sidebar">Pure sidebar content</div>
			</body></html>`,
			removed: []string{"Pure sidebar"},
			kept:    []string{"Article content"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			result := dom.CleanTags(doc)

			bodyText := result.Find("body").Text()

			// Check removed content
			for _, removedText := range tt.removed {
				assert.NotContains(t, bodyText, removedText, "Should not contain removed text")
			}

			// Check kept content
			for _, keptText := range tt.kept {
				assert.Contains(t, bodyText, keptText, "Should contain kept text")
			}
		})
	}
}

// TestCleanTagsFormDetection tests the form detection logic exactly like JavaScript
func TestCleanTagsFormDetection(t *testing.T) {
	// Based on JavaScript test: "removes a node with too many inputs"
	html := `<html><body>
		<div>
			<p>What do you think?</p>
			<p>What do you think?</p>
			<p>What do you think?</p>
			<p>What do you think?</p>
			<p>What do you think?</p>
			<p>What do you think?</p>
			<p>What do you think?</p>
			<div>
				<p>What is your name?</p>
				<input type="text"></input>
				<p>What is your name?</p>
				<input type="text"></input>
				<p>What is your name?</p>
				<input type="text"></input>
			</div>
			<p>What do you think?</p>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	result := dom.CleanTags(doc)

	// The inner div with too many inputs should be removed
	bodyText := result.Find("body").Text()
	assert.NotContains(t, bodyText, "What is your name?", "Form with too many inputs should be removed")
	assert.Contains(t, bodyText, "What do you think?", "Regular content should be kept")
	
	// Should have removed the form div but kept the content
	innerDivs := result.Find("div div")
	assert.Equal(t, 0, innerDivs.Length(), "Inner div with form should be removed")
}

// TestCleanTagsShortContentNoImages tests removal of short content without images
func TestCleanTagsShortContentNoImages(t *testing.T) {
	// Based on JavaScript test: "removes a div with no images and very little text"
	html := `<html><body>
		<div>
			<p>What do you think?</p>
			<div>
				<p>Keep this one</p>
				<img src="asdf" />
			</div>
			<div>
				<p>Lose this one</p>
			</div>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	result := dom.CleanTags(doc)

	bodyText := result.Find("body").Text()
	
	// Short content without image should be removed
	assert.NotContains(t, bodyText, "Lose this one", "Short content without image should be removed")
	
	// Content with image should be kept
	assert.Contains(t, bodyText, "Keep this one", "Content with image should be kept")
	assert.Contains(t, bodyText, "What do you think?", "Main content should be kept")
	
	// Should still have the image
	images := result.Find("img")
	assert.Equal(t, 1, images.Length(), "Image should be preserved")
}

// TestCleanTagsLinkDensity tests the link density removal logic
func TestCleanTagsLinkDensity(t *testing.T) {
	// Based on JavaScript test: "removes a node with a link density that is too high"
	html := `<html><body>
		<div score="0">
			<p>Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu.</p>
			<ul>
				<li>Keep this one</li>
				<li>Keep this one</li>
				<li>Keep this one</li>
				<li>Keep this one</li>
				<li>Keep this one</li>
				<li>Keep this one</li>
				<li>Keep this one</li>
			</ul>
			<ul score="20">
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
			</ul>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	result := dom.CleanTags(doc)

	bodyText := result.Find("body").Text()
	
	// High link density ul should be removed
	assert.NotContains(t, bodyText, "Lose this one", "High link density list should be removed")
	
	// Low link density content should be kept
	assert.Contains(t, bodyText, "Keep this one", "Low link density list should be kept")
	assert.Contains(t, bodyText, "Lorem ipsum", "Main content should be kept")
}

// TestCleanTagsColonException tests the colon exception for lists
func TestCleanTagsColonException(t *testing.T) {
	// Based on JavaScript test: "keeps node with a good score but link density > 0.5 if preceding text ends in colon"
	html := `<html><body>
		<div score="40">
			<p>Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu.</p>
			<p>Now read these links: </p>
			<ul score="30">
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
			</ul>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	result := dom.CleanTags(doc)

	bodyText := result.Find("body").Text()
	
	// With colon in previous sibling, high link density list should be KEPT
	assert.Contains(t, bodyText, "Lose this one", "List after colon should be kept despite high link density")
	assert.Contains(t, bodyText, "Now read these links:", "Colon text should be kept")
	assert.Contains(t, bodyText, "Lorem ipsum", "Main content should be kept")
}

// TestCleanTagsEntryContentAsset tests the entry-content-asset protection
func TestCleanTagsEntryContentAsset(t *testing.T) {
	// Based on JavaScript test: "keeps anything with a class of entry-content-asset"
	html := `<html><body>
		<div score="100">
			<p>Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu.</p>
			<ul score="20" class="entry-content-asset">
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
				<li><a href="#">Lose this one</a></li>
			</ul>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	result := dom.CleanTags(doc)

	bodyText := result.Find("body").Text()
	
	// entry-content-asset should be kept despite high link density
	assert.Contains(t, bodyText, "Lose this one", "entry-content-asset should be kept")
	assert.Contains(t, bodyText, "Lorem ipsum", "Main content should be kept")
	
	// Should still have the entry-content-asset class
	assetElements := result.Find(".entry-content-asset")
	assert.Equal(t, 1, assetElements.Length(), "entry-content-asset element should be preserved")
}

// TestCleanTagsNegativeScore tests removal of negative scored elements
func TestCleanTagsNegativeScore(t *testing.T) {
	// Based on JavaScript test: "drops a matching node with a negative score"
	html := `<html><body>
		<div score="5">
			<p>What do you think?</p>
			<p>
				<ul score="-10">
					<li>Foo</li>
					<li>Bar</li>
				</ul>
			</p>
			<p>What do you think?</p>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	result := dom.CleanTags(doc)

	bodyText := result.Find("body").Text()
	
	// Negative scored elements should be removed
	assert.NotContains(t, bodyText, "Foo", "Negative scored element should be removed")
	assert.NotContains(t, bodyText, "Bar", "Negative scored element should be removed")
	
	// Positive content should be kept
	assert.Contains(t, bodyText, "What do you think?", "Positive content should be kept")
}

// TestCleanTagsScriptRemoval tests removal of elements with too many scripts
func TestCleanTagsScriptRemoval(t *testing.T) {
	html := `<html><body>
		<div>
			<p>Good content with substantial text that should be preserved</p>
		</div>
		<div>
			<script>alert('test');</script>
			<p>Short text</p>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	result := dom.CleanTags(doc)

	bodyText := result.Find("body").Text()
	
	// Element with script and short content should be removed
	assert.NotContains(t, bodyText, "Short text", "Element with script and short content should be removed")
	
	// Good content should be kept
	assert.Contains(t, bodyText, "Good content with substantial text", "Good content should be kept")
}

func TestRemoveEmpty(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		remaining int
	}{
		{
			name: "removes empty paragraphs",
			html: `<html><body>
				<p>Good paragraph</p>
				<p></p>
				<p>   </p>
				<p>&nbsp;</p>
				<p>Another good paragraph</p>
			</body></html>`,
			remaining: 2,
		},
		{
			name: "keeps paragraphs with content",
			html: `<html><body>
				<p>Paragraph with text</p>
				<p><strong>Paragraph with formatting</strong></p>
				<p><img src="image.jpg"></p>
			</body></html>`,
			remaining: 3,
		},
		{
			name: "handles mixed content",
			html: `<html><body>
				<p>Content paragraph</p>
				<p></p>
				<div></div>
				<p>Another content paragraph</p>
			</body></html>`,
			remaining: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			result := dom.RemoveEmpty(doc)

			paragraphs := result.Find("p")
			assert.Equal(t, tt.remaining, paragraphs.Length(), "Should have expected number of paragraphs")
		})
	}
}

func TestStripJunkTags(t *testing.T) {
	html := `<html><head>
		<title>Page Title</title>
		<script>alert('test');</script>
		<style>body { color: red; }</style>
		<link rel="stylesheet" href="style.css">
	</head><body>
		<p>Content paragraph</p>
		<noscript>No script content</noscript>
		<hr>
		<iframe src="embed.html"></iframe>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	result := dom.StripJunkTags(doc)

	// Check that junk tags are removed
	assert.Equal(t, 0, result.Find("script").Length())
	assert.Equal(t, 0, result.Find("style").Length())
	assert.Equal(t, 0, result.Find("noscript").Length())
	assert.Equal(t, 0, result.Find("hr").Length())
	assert.Equal(t, 0, result.Find("iframe").Length())
	assert.Equal(t, 0, result.Find("link").Length())
	assert.Equal(t, 0, result.Find("title").Length())

	// Check that content is preserved
	assert.Equal(t, 1, result.Find("p").Length())
	assert.Contains(t, result.Find("p").Text(), "Content paragraph")
}

func TestMarkToKeep(t *testing.T) {
	html := `<html><body>
		<iframe src="https://www.youtube.com/embed/abc123">YouTube</iframe>
		<iframe src="https://player.vimeo.com/video/123456">Vimeo</iframe>
		<iframe src="https://example.com/other">Other</iframe>
		<div class="content">Regular content</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	result := dom.MarkToKeep(doc)

	// Check that important iframes are marked
	youtubeFrame := result.Find(`iframe[src^="https://www.youtube.com"]`)
	assert.True(t, youtubeFrame.HasClass(dom.KEEP_CLASS), "YouTube iframe should be marked to keep")

	vimeoFrame := result.Find(`iframe[src^="https://player.vimeo.com"]`)
	assert.True(t, vimeoFrame.HasClass(dom.KEEP_CLASS), "Vimeo iframe should be marked to keep")

	// Check that other iframe is not marked
	otherFrame := result.Find(`iframe[src^="https://example.com"]`)
	assert.False(t, otherFrame.HasClass(dom.KEEP_CLASS), "Other iframe should not be marked to keep")
}

func TestCleanImages(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		remaining int
		removed   []string
	}{
		{
			name: "removes spacer images",
			html: `<html><body>
				<img src="spacer.gif" alt="Spacer">
				<img src="transparent.png" alt="Transparent">
				<img src="photo.jpg" alt="Real photo">
				<img src="blank.gif" alt="Blank">
			</body></html>`,
			remaining: 1,
			removed:   []string{"spacer.gif", "transparent.png", "blank.gif"},
		},
		{
			name: "removes tiny images",
			html: `<html><body>
				<img src="pixel.gif" width="1" height="1" alt="Tracking pixel">
				<img src="normal.jpg" width="100" height="200" alt="Normal image">
				<img src="zero.gif" width="0" height="0" alt="Zero size">
			</body></html>`,
			remaining: 1,
			removed:   []string{"pixel.gif", "zero.gif"},
		},
		{
			name: "removes ad images based on class",
			html: `<html><body>
				<img src="ad.jpg" class="advertisement" alt="Ad">
				<img src="content.jpg" class="article-image" alt="Content">
				<img src="sidebar.jpg" class="sidebar-ad" alt="Sidebar ad">
			</body></html>`,
			remaining: 1,
			removed:   []string{"ad.jpg", "sidebar.jpg"},
		},
		{
			name: "removes images without src",
			html: `<html><body>
				<img alt="No source">
				<img src="good.jpg" alt="Good image">
			</body></html>`,
			remaining: 1,
			removed:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			result := dom.CleanImages(doc)

			images := result.Find("img")
			assert.Equal(t, tt.remaining, images.Length(), "Should have expected number of images")

			// Check that removed images are gone
			for _, removedSrc := range tt.removed {
				found := false
				images.Each(func(i int, img *goquery.Selection) {
					if src, exists := img.Attr("src"); exists && strings.Contains(src, removedSrc) {
						found = true
					}
				})
				assert.False(t, found, "Removed image should not be found: %s", removedSrc)
			}
		})
	}
}

func TestCleaningPipeline(t *testing.T) {
	// Test the full cleaning pipeline
	html := `<html><head>
		<title>Test Page</title>
		<script>alert('test');</script>
		<style>body { color: red; }</style>
	</head><body>
		<div class="header navigation" style="background: blue;">
			<h2 id="nav-title">Nav</h2>
			<ul class="nav-menu">
				<li><a href="#">Link 1</a></li>
				<li><a href="#">Link 2</a></li>
			</ul>
		</div>
		<div class="article-content main">
			<h2 class="article-title">Good Article Title</h2>
			<p>This is substantial article content that should be preserved.</p>
			<p></p>
			<img src="spacer.gif" alt="Spacer">
			<img src="article-photo.jpg" alt="Article photo" width="400" height="300">
		</div>
		<div class="sidebar">
			<h3>AD</h3>
			<div class="advertisement">Ad content</div>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	// Apply the full cleaning pipeline
	result := dom.StripJunkTags(doc)
	result = dom.StripUnlikelyCandidates(result)
	result = dom.CleanAttributes(result)
	result = dom.CleanHeadersWithoutTitle(result)
	result = dom.CleanTags(result)
	result = dom.RemoveEmpty(result)
	result = dom.CleanImages(result)

	// Check results
	assert.Equal(t, 0, result.Find("script, style, title").Length(), "Junk tags should be removed")
	assert.Equal(t, 0, result.Find(".header, .sidebar").Length(), "Unlikely candidates should be removed")
	assert.Equal(t, 0, result.Find("[style]").Length(), "Style attributes should be removed")
	assert.Equal(t, 1, result.Find("h2").Length(), "Should keep one good header")
	assert.Equal(t, 1, result.Find("p").Length(), "Should keep one good paragraph")
	assert.Equal(t, 1, result.Find("img").Length(), "Should keep one good image")

	// Check that good content remains
	bodyText := result.Find("body").Text()
	assert.Contains(t, bodyText, "Good Article Title")
	assert.Contains(t, bodyText, "substantial article content")
}

func BenchmarkCleaningFunctions(b *testing.B) {
	html := `<html><head>
		<script>alert('test');</script>
		<style>body { color: red; }</style>
	</head><body>
		<div class="header" style="background: blue;" align="center">
			<h2>Navigation</h2>
			<ul><li><a href="#">Link</a></li></ul>
		</div>
		<div class="content">
			<h2>Article Title</h2>
			<p>Content paragraph</p>
			<p></p>
			<img src="spacer.gif">
		</div>
	</body></html>`

	b.Run("CleanAttributes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
			dom.CleanAttributes(doc)
		}
	})

	b.Run("CleanHeaders", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
			dom.CleanHeadersWithoutTitle(doc)
		}
	})

	b.Run("CleanTags", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
			dom.CleanTags(doc)
		}
	})

	b.Run("FullCleaningPipeline", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
			doc = dom.StripJunkTags(doc)
			doc = dom.CleanAttributes(doc)
			doc = dom.CleanHeadersWithoutTitle(doc)
			doc = dom.CleanTags(doc)
			doc = dom.RemoveEmpty(doc)
			doc = dom.CleanImages(doc)
		}
	})
}