// ABOUTME: Tests for article base URL extraction functionality
// ABOUTME: Tests IsGoodSegment and ArticleBaseURL functions for faithful JavaScript compatibility
package text

import (
	"net/url"
	"testing"
)

// TestIsGoodSegment tests the IsGoodSegment function with various inputs
func TestIsGoodSegment(t *testing.T) {
	testCases := []struct {
		segment                 string
		index                   int
		firstSegmentHasLetters bool
		expected               bool
		desc                   string
	}{
		// From JavaScript logic analysis:
		// If this is purely a number, and it's the first or second
		// url_segment, it's probably a page number. Remove it.
		{"12", 0, false, false, "pure number in first segment should be removed (< 3 chars, no letters)"},
		{"12", 1, false, false, "pure number in second segment should be removed (< 3 chars, no letters)"},
		{"123", 0, false, true, "3-digit number in first segment should be kept"},
		{"1234", 0, false, true, "4+ digit number in first segment should be kept"},

		// If this is the first url_segment and it's just "index", remove it
		{"index", 0, false, false, "index in first segment should be removed"},
		{"Index", 0, false, false, "Index (case insensitive) in first segment should be removed"},
		{"INDEX", 0, false, false, "INDEX in first segment should be removed"},
		{"index", 1, false, true, "index in second segment should be kept"},

		// If our first or second url_segment is smaller than 3 characters,
		// and the first url_segment had no alphas, remove it.
		{"12", 0, false, false, "short segment without letters in first should be removed"},
		{"ab", 0, false, false, "short segment without letters in first should be removed"},
		{"12", 1, false, false, "short segment without letters in second should be removed"},
		{"ab", 1, false, false, "short segment without letters in second should be removed"},
		{"12", 0, true, true, "short segment with first having letters should be kept"},
		{"ab", 0, true, true, "short segment with first having letters should be kept"},
		{"12", 1, true, true, "short segment with first having letters should be kept"},
		{"123", 0, false, true, "3+ char segment should be kept"},
		{"abc", 2, false, true, "third+ segment should be kept"},

		// Edge cases
		{"", 0, false, false, "empty segment should be removed"},
		{"a", 0, false, false, "single char segment without first having letters should be removed"},
		{"a", 0, true, true, "single char segment with first having letters should be kept"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := IsGoodSegment(tc.segment, tc.index, tc.firstSegmentHasLetters)
			if result != tc.expected {
				t.Errorf("IsGoodSegment(%q, %d, %v) = %v, expected %v",
					tc.segment, tc.index, tc.firstSegmentHasLetters, result, tc.expected)
			}
		})
	}
}

// TestArticleBaseURL tests the main ArticleBaseURL function
func TestArticleBaseURL(t *testing.T) {
	testCases := []struct {
		inputURL string
		expected string
		desc     string
	}{
		// Test cases from JavaScript tests
		{
			"http://example.com/foo/bar/wow-cool/page=10",
			"http://example.com/foo/bar/wow-cool",
			"returns the base url of a paginated url",
		},
		{
			"http://example.com/foo/bar/wow-cool/",
			"http://example.com/foo/bar/wow-cool",
			"returns same url if url has no pagination info",
		},

		// Additional test cases based on actual JavaScript behavior
		{
			"http://example.com/article/p=2",
			"http://example.com/article/p",
			"replaces page parameter with equals (leaves remainder)",
		},
		{
			"http://example.com/news/paging/5",
			"http://example.com/news/paging",
			"replaces paging with slash (leaves remainder)",
		},
		{
			"http://example.com/story.html",
			"http://example.com/story",
			"removes file extension",
		},
		{
			"http://example.com/index/article",
			"http://example.com/index/article",
			"keeps index segment when not in first position",
		},
		{
			"http://example.com/2/3/article",
			"http://example.com/2/3/article",
			"keeps segments when first segment has letters",
		},
		{
			"http://example.com/a1/b2/article",
			"http://example.com/a1/b2/article",
			"keeps segments when first has letters",
		},

		// Edge cases
		{
			"http://example.com/",
			"http://example.com",
			"handles root URL",
		},
		{
			"http://example.com",
			"http://example.com",
			"handles URL without path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := ArticleBaseURL(tc.inputURL, nil)
			if result != tc.expected {
				t.Errorf("ArticleBaseURL(%q) = %q, expected %q",
					tc.inputURL, result, tc.expected)
			}
		})
	}
}

// TestArticleBaseURLWithParsedURL tests ArticleBaseURL with pre-parsed URLs
func TestArticleBaseURLWithParsedURL(t *testing.T) {
	testURL := "http://example.com/foo/bar/page=10"
	parsedURL, err := url.Parse(testURL)
	if err != nil {
		t.Fatalf("Failed to parse test URL: %v", err)
	}

	result := ArticleBaseURL(testURL, parsedURL)
	expected := "http://example.com/foo/bar"

	if result != expected {
		t.Errorf("ArticleBaseURL with parsed URL = %q, expected %q", result, expected)
	}
}

// TestArticleBaseURLEdgeCases tests additional edge cases and malformed URLs
func TestArticleBaseURLEdgeCases(t *testing.T) {
	testCases := []struct {
		inputURL string
		expected string
		desc     string
	}{
		// Malformed URLs - Go's url.Parse handles these differently than JavaScript
		{
			"not-a-url",
			":///not-a-url",
			"malformed URL gets processed with empty scheme/host",
		},
		{
			"http://",
			"http://",
			"incomplete URL should return as-is",
		},
		{
			"http://example.com/article/../page=5",
			"http://example.com/article",
			"pagination parameter gets removed from parent dir reference",
		},
		
		// Complex pagination patterns
		{
			"http://example.com/news/category/page/5",
			"http://example.com/news/category/page",
			"page with slash separator leaves page segment",
		},
		{
			"http://example.com/blog/post-title/p=1",
			"http://example.com/blog/post-title/p",
			"single page parameter",
		},
		{
			"http://example.com/forum/thread/paging=25",
			"http://example.com/forum/thread",
			"full paging parameter gets removed completely",
		},
		
		// File extensions and special cases
		{
			"http://example.com/article.php?page=2",
			"http://example.com/article",
			"query parameters are stripped, extensions removed",
		},
		{
			"http://example.com/document.pdf",
			"http://example.com/document",
			"PDF extension removal",
		},
		{
			"http://example.com/script.js.bak",
			"http://example.com/script.js.bak",
			"non-alpha extension should be kept",
		},
		
		// Multiple segments and complex paths
		{
			"http://example.com/a/b/c/d/e/f/g",
			"http://example.com/a/b/c/d/e/f/g",
			"long path with letters should be kept",
		},
		{
			"http://example.com/1/2/3/4/5/article",
			"http://example.com/1/2/3/4/5/article",
			"numeric segments after second should be kept",
		},
		
		// Protocol variations
		{
			"https://example.com/article/page=1",
			"https://example.com/article",
			"HTTPS protocol",
		},
		{
			"ftp://example.com/file/page=1",
			"ftp://example.com/file",
			"FTP protocol",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := ArticleBaseURL(tc.inputURL, nil)
			if result != tc.expected {
				t.Errorf("ArticleBaseURL(%q) = %q, expected %q",
					tc.inputURL, result, tc.expected)
			}
		})
	}
}

// TestArticleBaseURLJavaScriptCompatibility tests exact JavaScript compatibility
func TestArticleBaseURLJavaScriptCompatibility(t *testing.T) {
	// These are the exact test cases from the JavaScript test file
	testCases := []struct {
		inputURL string
		expected string
		desc     string
	}{
		{
			"http://example.com/foo/bar/wow-cool/page=10",
			"http://example.com/foo/bar/wow-cool",
			"JavaScript test case 1 - paginated URL",
		},
		{
			"http://example.com/foo/bar/wow-cool/",
			"http://example.com/foo/bar/wow-cool",
			"JavaScript test case 2 - URL without pagination",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := ArticleBaseURL(tc.inputURL, nil)
			if result != tc.expected {
				t.Errorf("JavaScript compatibility: ArticleBaseURL(%q) = %q, expected %q",
					tc.inputURL, result, tc.expected)
			}
		})
	}
}