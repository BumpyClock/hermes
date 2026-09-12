// ABOUTME: NPR custom extractor with storytitle, byline__name, and storytext content with bucketwrap transforms
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.npr.org/index.js WwwNprOrgExtractor

package custom

// GetNPRExtractor returns the custom extractor for www.npr.org.
func GetNPRExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.npr.org",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1"},
				{Selector: ".storytitle"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "p.byline__name.byline__name--block"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `.dateblock time[datetime]`, Attribute: "datetime"},
				{Selector: `meta[name="date"]`, Attribute: "value"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
				{Selector: `meta[name="twitter:image:src"]`, Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".storytext"},
			},

			Transforms: map[string]TransformFunction{
				".bucketwrap.image": &StringTransform{
					TargetTag: "figure",
				},
				".bucketwrap.image .credit-caption": &StringTransform{
					TargetTag: "figcaption",
				},
			},

			Clean: []string{
				"div.enlarge_measure",
				"b.toggle-caption",
				"b.hide-caption",
				".ad-header",
				".ad-wrap",
				"aside.ad-wrap",
				"aside[id*='ad-']",
				"button",
				// Additional patterns for UI elements
				".bucketwrap .toggle-caption",
				".bucketwrap .hide-caption",
			},
		},
	}
}
