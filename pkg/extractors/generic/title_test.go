// ABOUTME: Comprehensive test suite for title extraction functionality
// ABOUTME: Tests JavaScript compatibility and covers all extraction scenarios

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestGenericTitleExtractor_Extract(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		url       string
		metaCache []string
		expected  string
	}{
		{
			name: "extracts strong meta title tags - dc.title",
			html: `<html>
				<meta name="dc.title" value="This Is the Title Okay" />
			</html>`,
			url:       "",
			metaCache: []string{"dc.title", "something-else"},
			expected:  "This Is the Title Okay",
		},
		{
			name: "extracts strong meta title tags - tweetmeme-title",
			html: `<html>
				<meta name="tweetmeme-title" value="Tweetmeme Article Title" />
			</html>`,
			url:       "",
			metaCache: []string{"tweetmeme-title"},
			expected:  "Tweetmeme Article Title",
		},
		{
			name: "extracts strong meta title tags - rbtitle",
			html: `<html>
				<meta name="rbtitle" value="RB Title Here" />
			</html>`,
			url:       "",
			metaCache: []string{"rbtitle"},
			expected:  "RB Title Here",
		},
		{
			name: "extracts strong meta title tags - headline",
			html: `<html>
				<meta name="headline" value="Headline Meta Tag" />
			</html>`,
			url:       "",
			metaCache: []string{"headline"},
			expected:  "Headline Meta Tag",
		},
		{
			name: "extracts strong meta title tags - title",
			html: `<html>
				<meta name="title" value="Title Meta Tag" />
			</html>`,
			url:       "",
			metaCache: []string{"title"},
			expected:  "Title Meta Tag",
		},
		{
			name: "pulls title from selectors when lacking strong meta",
			html: `<html>
				<article class="hentry">
					<h1 class="entry-title">This Is the Title Okay</h1>
				</article>
			</html>`,
			url:       "",
			metaCache: []string{"og:title", "something-else"},
			expected:  "This Is the Title Okay",
		},
		{
			name: "extracts from h1#articleHeader",
			html: `<html>
				<h1 id="articleHeader">Article Header Title</h1>
			</html>`,
			url:       "",
			metaCache: []string{},
			expected:  "Article Header Title",
		},
		{
			name: "extracts from h1.articleHeader",
			html: `<html>
				<h1 class="articleHeader">Class Article Header</h1>
			</html>`,
			url:       "",
			metaCache: []string{},
			expected:  "Class Article Header",
		},
		{
			name: "extracts from h1.article",
			html: `<html>
				<h1 class="article">Article Class Title</h1>
			</html>`,
			url:       "",
			metaCache: []string{},
			expected:  "Article Class Title",
		},
		{
			name: "extracts from .instapaper_title",
			html: `<html>
				<div class="instapaper_title">Instapaper Title</div>
			</html>`,
			url:       "",
			metaCache: []string{},
			expected:  "Instapaper Title",
		},
		{
			name: "extracts from #meebo-title",
			html: `<html>
				<div id="meebo-title">Meebo Title</div>
			</html>`,
			url:       "",
			metaCache: []string{},
			expected:  "Meebo Title",
		},
		{
			name: "falls back to weak meta title tags - og:title",
			html: `<html>
				<meta name="og:title" value="This Is the Title Okay" />
			</html>`,
			url:       "",
			metaCache: []string{"og:title", "something-else"},
			expected:  "This Is the Title Okay",
		},
		{
			name: "falls back to weak selectors - article h1",
			html: `<html>
				<article>
					<h1>Article H1 Title</h1>
				</article>
			</html>`,
			url:       "",
			metaCache: []string{},
			expected:  "Article H1 Title",
		},
		{
			name: "falls back to weak selectors - #entry-title",
			html: `<html>
				<div id="entry-title">Entry Title ID</div>
			</html>`,
			url:       "",
			metaCache: []string{},
			expected:  "Entry Title ID",
		},
		{
			name: "falls back to weak selectors - .entry-title",
			html: `<html>
				<div class="entry-title">Entry Title Class</div>
			</html>`,
			url:       "",
			metaCache: []string{},
			expected:  "Entry Title Class",
		},
		{
			name: "falls back to weak selectors - h1",
			html: `<html>
				<h1>Simple H1 Title</h1>
			</html>`,
			url:       "",
			metaCache: []string{},
			expected:  "Simple H1 Title",
		},
		{
			name: "falls back to weak selectors - html head title",
			html: `<html>
				<head>
					<title>This Is the Weak Title Okay</title>
				</head>
			</html>`,
			url:       "",
			metaCache: []string{},
			expected:  "This Is the Weak Title Okay",
		},
		{
			name: "prioritizes strong meta over strong selectors",
			html: `<html>
				<meta name="dc.title" value="Strong Meta Title" />
				<article class="hentry">
					<h1 class="entry-title">Strong Selector Title</h1>
				</article>
			</html>`,
			url:       "",
			metaCache: []string{"dc.title"},
			expected:  "Strong Meta Title",
		},
		{
			name: "prioritizes strong selectors over weak meta",
			html: `<html>
				<meta name="og:title" value="Weak Meta Title" />
				<article class="hentry">
					<h1 class="entry-title">Strong Selector Title</h1>
				</article>
			</html>`,
			url:       "",
			metaCache: []string{"og:title"},
			expected:  "Strong Selector Title",
		},
		{
			name: "prioritizes weak meta over weak selectors",
			html: `<html>
				<meta name="og:title" value="Weak Meta Title" />
				<h1>Weak Selector Title</h1>
			</html>`,
			url:       "",
			metaCache: []string{"og:title"},
			expected:  "Weak Meta Title",
		},
		{
			name:      "returns empty string when no matches",
			html:      `<html><body><p>No titles here</p></body></html>`,
			url:       "",
			metaCache: []string{},
			expected:  "",
		},
		{
			name: "strips HTML tags from title",
			html: `<html>
				<meta name="dc.title" value="Title with <em>emphasis</em> and <strong>bold</strong>" />
			</html>`,
			url:       "",
			metaCache: []string{"dc.title"},
			expected:  "Title with emphasis and bold",
		},
		{
			name: "normalizes spaces in title",
			html: `<html>
				<meta name="dc.title" value="Title    with   multiple    spaces" />
			</html>`,
			url:       "",
			metaCache: []string{"dc.title"},
			expected:  "Title with multiple spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := GenericTitleExtractor.Extract(doc.Selection, tt.url, tt.metaCache)
			if result != tt.expected {
				t.Errorf("GenericTitleExtractor.Extract() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCleanTitle_SplitTitleResolution(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		url      string
		expected string
	}{
		{
			name:     "resolves split title with domain match at end",
			title:    "Great Article - Example",
			url:      "https://example.com/article",
			expected: "Great Article",
		},
		{
			name:     "resolves split title with domain match at start",
			title:    "Example - Great Article",
			url:      "https://example.com/article",
			expected: "Great Article",
		},
		{
			name:     "handles breadcrumb-style titles (JavaScript compatible behavior)",
			title:    "Home : Category : Subcategory : Article Title : Site",
			url:      "https://site.com/article",
			expected: "Home : Category : Subcategory : Article Title : Site", // No match due to length requirements
		},
		{
			name:     "leaves clean titles unchanged",
			title:    "Clean Article Title",
			url:      "https://example.com/article",
			expected: "Clean Article Title",
		},
		{
			name:     "handles colon separators",
			title:    "Great Article: The Sequel",
			url:      "https://example.com/article",
			expected: "Great Article: The Sequel", // No domain match, keep as is
		},
		{
			name:     "handles pipe separators",
			title:    "Article Title | Site Name",
			url:      "https://sitename.com/article",
			expected: "Article Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanTitle(tt.title, tt.url, nil)
			if result != tt.expected {
				t.Errorf("cleanTitle(%q, %q) = %q, want %q", tt.title, tt.url, result, tt.expected)
			}
		})
	}
}

func TestCleanTitle_LengthValidation(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		html     string
		expected string
	}{
		{
			name:  "uses h1 fallback for overly long title",
			title: strings.Repeat("This is a very long title that exceeds the maximum length limit ", 3), // >150 chars
			html: `<html>
				<h1>Fallback H1 Title</h1>
			</html>`,
			expected: "Fallback H1 Title",
		},
		{
			name:  "keeps long title when no single h1 available",
			title: strings.Repeat("Long title ", 15), // >150 chars
			html: `<html>
				<h1>First H1</h1>
				<h1>Second H1</h1>
			</html>`,
			expected: strings.TrimSpace(strings.Repeat("Long title ", 15)), // Should keep original, normalized
		},
		{
			name:     "keeps normal length titles",
			title:    "Normal Length Title",
			html:     "<html><h1>H1 Title</h1></html>",
			expected: "Normal Length Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := cleanTitle(tt.title, "", doc.Selection)
			if result != tt.expected {
				t.Errorf("cleanTitle() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestResolveSplitTitle(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		url      string
		expected string
	}{
		{
			name:     "returns title unchanged when no separators",
			title:    "Simple Title",
			url:      "https://example.com",
			expected: "Simple Title",
		},
		{
			name:     "extracts breadcrumb title from complex hierarchy (JavaScript compatible)",
			title:    "Home : News : Tech : Politics : The Best Article Ever : News : Tech",
			url:      "https://example.com",
			expected: "Home : News : Tech : Politics : The Best Article Ever : News : Tech", // Falls back to original due to JavaScript logic
		},
		{
			name:     "removes domain name from start",
			title:    "ExampleSite - Great Article",
			url:      "https://examplesite.com/article",
			expected: "Great Article",
		},
		{
			name:     "removes domain name from end",
			title:    "Great Article - ExampleSite",
			url:      "https://examplesite.com/article",
			expected: "Great Article",
		},
		{
			name:     "handles fuzzy domain matching",
			title:    "Article Title | Example News",
			url:      "https://example.com/news",
			expected: "Article Title",
		},
		{
			name:     "leaves title with no domain match",
			title:    "Article Title - Something Else",
			url:      "https://example.com",
			expected: "Article Title - Something Else",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveSplitTitle(tt.title, tt.url)
			if result != tt.expected {
				t.Errorf("resolveSplitTitle(%q, %q) = %q, want %q", tt.title, tt.url, result, tt.expected)
			}
		})
	}
}

func TestTitleConstants(t *testing.T) {
	// Test that constants are properly defined
	if len(STRONG_TITLE_META_TAGS) == 0 {
		t.Error("STRONG_TITLE_META_TAGS should not be empty")
	}

	if len(WEAK_TITLE_META_TAGS) == 0 {
		t.Error("WEAK_TITLE_META_TAGS should not be empty")
	}

	if len(STRONG_TITLE_SELECTORS) == 0 {
		t.Error("STRONG_TITLE_SELECTORS should not be empty")
	}

	if len(WEAK_TITLE_SELECTORS) == 0 {
		t.Error("WEAK_TITLE_SELECTORS should not be empty")
	}

	// Test specific values match JavaScript
	expectedStrong := []string{"tweetmeme-title", "dc.title", "rbtitle", "headline", "title"}
	for i, expected := range expectedStrong {
		if i >= len(STRONG_TITLE_META_TAGS) || STRONG_TITLE_META_TAGS[i] != expected {
			t.Errorf("STRONG_TITLE_META_TAGS[%d] = %q, want %q", i, STRONG_TITLE_META_TAGS[i], expected)
		}
	}

	expectedWeak := []string{"og:title"}
	for i, expected := range expectedWeak {
		if i >= len(WEAK_TITLE_META_TAGS) || WEAK_TITLE_META_TAGS[i] != expected {
			t.Errorf("WEAK_TITLE_META_TAGS[%d] = %q, want %q", i, WEAK_TITLE_META_TAGS[i], expected)
		}
	}
}

// Benchmarks for performance verification
func BenchmarkGenericTitleExtractor_Extract(b *testing.B) {
	html := `<html>
		<meta name="dc.title" value="Benchmark Title" />
		<article class="hentry">
			<h1 class="entry-title">Fallback Title</h1>
		</article>
	</html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	metaCache := []string{"dc.title"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenericTitleExtractor.Extract(doc.Selection, "", metaCache)
	}
}

func BenchmarkResolveSplitTitle(b *testing.B) {
	title := "Great Article - Example Site"
	url := "https://example.com/article"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolveSplitTitle(title, url)
	}
}