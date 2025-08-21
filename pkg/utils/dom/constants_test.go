package dom_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/BumpyClock/hermes/pkg/utils/dom"
)

func TestConstants_RegexPatterns(t *testing.T) {
	tests := []struct {
		name    string
		regex   string
		input   string
		matches bool
	}{
		{
			name:    "SPACER_RE matches transparent",
			input:   "transparent.gif",
			matches: true,
		},
		{
			name:    "SPACER_RE matches spacer",
			input:   "spacer.png",
			matches: true,
		},
		{
			name:    "SPACER_RE matches blank",
			input:   "blank.jpg",
			matches: true,
		},
		{
			name:    "SPACER_RE case insensitive",
			input:   "TRANSPARENT.GIF",
			matches: true,
		},
		{
			name:    "SPACER_RE doesn't match normal image",
			input:   "photo.jpg",
			matches: false,
		},
		{
			name:    "POSITIVE_SCORE_RE matches article",
			input:   "article-content",
			matches: true,
		},
		{
			name:    "POSITIVE_SCORE_RE matches content",
			input:   "main-content",
			matches: true,
		},
		{
			name:    "NEGATIVE_SCORE_RE matches sidebar",
			input:   "sidebar-widget",
			matches: true,
		},
		{
			name:    "NEGATIVE_SCORE_RE matches footer",
			input:   "page-footer",
			matches: true,
		},
		{
			name:    "BLOCK_LEVEL_TAGS_RE matches div",
			input:   "div",
			matches: true,
		},
		{
			name:    "BLOCK_LEVEL_TAGS_RE matches article",
			input:   "article",
			matches: true,
		},
		{
			name:    "BLOCK_LEVEL_TAGS_RE doesn't match span",
			input:   "span",
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matches bool
			switch tt.name {
			case "SPACER_RE matches transparent", "SPACER_RE matches spacer", "SPACER_RE matches blank", "SPACER_RE case insensitive", "SPACER_RE doesn't match normal image":
				matches = dom.SPACER_RE.MatchString(tt.input)
			case "POSITIVE_SCORE_RE matches article", "POSITIVE_SCORE_RE matches content":
				matches = dom.POSITIVE_SCORE_RE.MatchString(tt.input)
			case "NEGATIVE_SCORE_RE matches sidebar", "NEGATIVE_SCORE_RE matches footer":
				matches = dom.NEGATIVE_SCORE_RE.MatchString(tt.input)
			case "BLOCK_LEVEL_TAGS_RE matches div", "BLOCK_LEVEL_TAGS_RE matches article", "BLOCK_LEVEL_TAGS_RE doesn't match span":
				matches = dom.BLOCK_LEVEL_TAGS_RE.MatchString(tt.input)
			}
			
			assert.Equal(t, tt.matches, matches)
		})
	}
}

func TestConstants_Lists(t *testing.T) {
	// Test that important constants are properly defined
	assert.NotEmpty(t, dom.STRIP_OUTPUT_TAGS)
	assert.Contains(t, dom.STRIP_OUTPUT_TAGS, "script")
	assert.Contains(t, dom.STRIP_OUTPUT_TAGS, "style")
	
	assert.NotEmpty(t, dom.WHITELIST_ATTRS)
	assert.Contains(t, dom.WHITELIST_ATTRS, "src")
	assert.Contains(t, dom.WHITELIST_ATTRS, "href")
	
	assert.NotEmpty(t, dom.BLOCK_LEVEL_TAGS)
	assert.Contains(t, dom.BLOCK_LEVEL_TAGS, "div")
	assert.Contains(t, dom.BLOCK_LEVEL_TAGS, "p")
	
	assert.Equal(t, "mercury-parser-keep", dom.KEEP_CLASS)
}

func TestConstants_CandidatesRegex(t *testing.T) {
	// Test blacklist patterns
	blacklistCases := []string{
		"sidebar-content",
		"nav-menu",
		"ad-banner",
		"footer-links",
		"comment-section",
	}
	
	for _, testCase := range blacklistCases {
		assert.True(t, dom.CANDIDATES_BLACKLIST.MatchString(testCase), "Should match blacklist: %s", testCase)
	}
	
	// Test whitelist patterns
	whitelistCases := []string{
		"article-content",
		"main-content",
		"entry-content",
		"post-body",
	}
	
	for _, testCase := range whitelistCases {
		assert.True(t, dom.CANDIDATES_WHITELIST.MatchString(testCase), "Should match whitelist: %s", testCase)
	}
}

func TestConstants_HelperFunctions(t *testing.T) {
	// Test helper functions
	removeSelectors := dom.GetRemoveAttrSelectors()
	assert.NotEmpty(t, removeSelectors)
	assert.Contains(t, removeSelectors, "[style]")
	assert.Contains(t, removeSelectors, "[align]")
	
	emptySelectors := dom.GetRemoveEmptySelectors()
	assert.NotEmpty(t, emptySelectors)
	assert.Contains(t, emptySelectors, "p:empty")
}