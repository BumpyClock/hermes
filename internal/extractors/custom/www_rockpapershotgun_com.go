// ABOUTME: Rock Paper Shotgun custom extractor for gaming news and reviews
// ABOUTME: Handles modern website structure with semantic selectors and robust fallbacks

package custom

// WwwRockpapershotgunComExtractor provides the custom extraction rules for www.rockpapershotgun.com.
var WwwRockpapershotgunComExtractor = &CustomExtractor{
	Domain: "www.rockpapershotgun.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.title"}, // Primary title selector
			{Selector: "h1"},       // Generic fallback
			{Selector: "meta[property=\"og:title\"]", Attribute: "content"}, // Meta fallback
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".byline .author a"},                           // Primary author link
			{Selector: ".byline .author"},                             // Author span without link
			{Selector: "meta[name=\"author\"]", Attribute: "content"}, // Meta fallback from JSON-LD
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "time", Attribute: "datetime"},                                     // Primary datetime attribute
			{Selector: "meta[property=\"article:published_time\"]", Attribute: "content"}, // Meta fallback
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "p.strapline"}, // Rock Paper Shotgun's subtitle/deck
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[property=\"og:image\"]", Attribute: "content"}, // Primary image from meta
			{Selector: ".headline_image"},                                   // Direct image selector fallback
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".article_body_content.article-styling"}, // Primary content container
			{".article_body_content"},                 // Without styling class fallback
			{".article-content"},                      // Generic article content
			{"article .article_body"},                 // Article body fallback
		},

		// Clean selectors - remove unwanted elements
		Clean: []string{
			// Advertisement and promotional content
			".inlinead",
			".desktop_mpu",
			".mpu_container",
			".advert_container",
			".leaderboard_container",
			".injection_placeholder",     // Ad injection points
			"span.injection_placeholder", // Ad injection spans
			"[data-position]",            // Elements with ad positioning data

			// Navigation and UI elements
			".read-next",
			".article_footer",
			".comments__link",
			".load-comments",
			".smart-slot", // Newsletter signup
			".sign-in-buttons",

			// Metadata that doesn't belong in content
			".byline",
			".metadata",
			".avatar",
			".published_at",
			".tagged_with",
			".author-inline",

			// Interactive elements
			"button",
			"form",

			// Social sharing
			".social-sign-in-button",

			// Related content
			".tagged_with_item",

			// Comments section
			".comments-bubble",
		},
	},
}

// GetWwwRockpapershotgunComExtractor returns the Rock Paper Shotgun custom extractor.
func GetWwwRockpapershotgunComExtractor() *CustomExtractor {
	return WwwRockpapershotgunComExtractor
}
