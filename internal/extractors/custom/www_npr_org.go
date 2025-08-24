// ABOUTME: NPR custom extractor with storytitle, byline__name, and storytext content with bucketwrap transforms
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.npr.org/index.js WwwNprOrgExtractor

package custom

// GetNPRExtractor returns the custom extractor for www.npr.org
func GetNPRExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.npr.org",

		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1",
				".storytitle",
			},
		},

		Author: &FieldExtractor{
			Selectors: []interface{}{
				"p.byline__name.byline__name--block",
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`.dateblock time[datetime]`, "datetime"},
				[]string{`meta[name="date"]`, "value"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:image"]`, "value"},
				[]string{`meta[name="twitter:image:src"]`, "value"},
			},
		},

		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					".storytext",
				},
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
