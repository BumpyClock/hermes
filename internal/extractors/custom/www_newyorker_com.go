// ABOUTME: New Yorker custom extractor for long-form journalism with special typography
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.newyorker.com/index.js

package custom

// NewYorkerCustomExtractor provides the custom extraction rules for www.newyorker.com
// JavaScript equivalent: export const NewYorkerExtractor = { ... }.
var NewYorkerCustomExtractor = &CustomExtractor{
	Domain: "www.newyorker.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1[class^=\"content-header\"]"},
			{Selector: "h1[class^=\"ArticleHeader__hed\"]"},
			{Selector: "h1[class*=\"ContentHeaderHed\"]"},
			{Selector: "meta[name=\"og:title\"]", Attribute: "value"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "article header div[class^=\"BylinesWrapper\"]"},
			{Selector: "meta[name=\"article:author\"]", Attribute: "value"},
			{Selector: "div[class^=\"ArticleContributors\"] a[rel=\"author\"]"},
			{Selector: "article header div[class*=\"Byline__multipleContributors\"]"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".article__body"},
			{"article.article.main-content"},
			{"main[class^=\"Layout__content\"]"},
		},

		// Transform functions for New Yorker-specific content
		Transforms: map[string]TransformFunction{
			".caption__text":   &StringTransform{TargetTag: "figcaption"},
			".caption__credit": &StringTransform{TargetTag: "figcaption"},
		},

		// Clean selectors - remove unwanted elements
		Clean: []string{
			"footer[class^=\"ArticleFooter__footer\"]",
			"aside",
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
			{Selector: "time.content-header__publish-date"},
			{Selector: "meta[name=\"pubdate\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "div[class^=\"ContentHeaderDek\"]"},
			{Selector: "div.content-header__dek"},
			{Selector: "h2[class^=\"ArticleHeader__dek\"]"},
		},
	},
}

// GetNewYorkerExtractor returns the New Yorker custom extractor.
func GetNewYorkerExtractor() *CustomExtractor {
	return NewYorkerCustomExtractor
}
