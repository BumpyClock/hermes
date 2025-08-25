// ABOUTME: New Yorker custom extractor for long-form journalism with special typography
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.newyorker.com/index.js

package custom

// NewYorkerCustomExtractor provides the custom extraction rules for www.newyorker.com
// JavaScript equivalent: export const NewYorkerExtractor = { ... }
var NewYorkerCustomExtractor = &CustomExtractor{
	Domain: "www.newyorker.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1[class^=\"content-header\"]",
			"h1[class^=\"ArticleHeader__hed\"]",
			"h1[class*=\"ContentHeaderHed\"]",
			[]string{"meta[name=\"og:title\"]", "value"},
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"article header div[class^=\"BylinesWrapper\"]",
			[]string{"meta[name=\"article:author\"]", "value"},
			"div[class^=\"ArticleContributors\"] a[rel=\"author\"]",
			"article header div[class*=\"Byline__multipleContributors\"]",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".article__body",
				"article.article.main-content",
				"main[class^=\"Layout__content\"]",
			},
		},
		
		// Transform functions for New Yorker-specific content
		Transforms: map[string]TransformFunction{
			".caption__text": &StringTransform{TargetTag: "figcaption"},
			".caption__credit": &StringTransform{TargetTag: "figcaption"},
		},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			"footer[class^=\"ArticleFooter__footer\"]",
			"aside",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
			"time.content-header__publish-date",
			[]string{"meta[name=\"pubdate\"]", "value"},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			"div[class^=\"ContentHeaderDek\"]",
			"div.content-header__dek",
			"h2[class^=\"ArticleHeader__dek\"]",
		},
	},
	
	// Next page URL and excerpt are null in original JavaScript
	NextPageURL: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{},
	},
}

// GetNewYorkerExtractor returns the New Yorker custom extractor
func GetNewYorkerExtractor() *CustomExtractor {
	return NewYorkerCustomExtractor
}