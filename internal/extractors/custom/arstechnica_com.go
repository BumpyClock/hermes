// ABOUTME: Ars Technica custom extractor with h2 paragraph transform handling
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/arstechnica.com/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// ArstechnicaComExtractor provides the custom extraction rules for arstechnica.com
// JavaScript equivalent: export const ArstechnicaComExtractor = { ... }.
var ArstechnicaComExtractor = &CustomExtractor{
	Domain: "arstechnica.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{{Selector: "article h1"}, {Selector: "h1"}, {Selector: "title"}},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"author\"]", Attribute: "value"},
			{Selector: "*[rel=\"author\"] *[itemprop=\"name\"]"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".byline time", Attribute: "datetime"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h2[itemprop=\"description\"]"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"article"},
			{"div[itemprop=\"articleBody\"]"},
		},

		// Transform functions for Ars Technica-specific content
		Transforms: map[string]TransformFunction{
			// Some pages have an element h2 that is significant, and that the parser will
			// remove if not following a paragraph. Adding this empty paragraph fixes it, and
			// the empty paragraph will be removed anyway.
			"h2": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					selection.BeforeHtml("<p></p>")
					return nil
				},
			},
		},

		// Clean selectors - remove unwanted elements
		Clean: []string{
			"article > header",
			// Remove enlarge links and separators inside image captions.
			"figcaption .enlarge-link",
			"figcaption .sep",

			// I could not transform the video into usable elements, so I
			// removed them.
			"figure.video",

			// Image galleries that do not work.
			".gallery",

			"aside",
			".sidebar",
		},
	},
}

// GetArstechnicaComExtractor returns the Ars Technica custom extractor.
func GetArstechnicaComExtractor() *CustomExtractor {
	return ArstechnicaComExtractor
}
