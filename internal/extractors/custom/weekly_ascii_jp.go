// ABOUTME: Weekly ASCII Japan tech magazine custom extractor with Japanese date format and timezone
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/weekly.ascii.jp/index.js

package custom

// WeeklyAsciiJpExtractor provides the custom extraction rules for weekly.ascii.jp
// JavaScript equivalent: export const WeeklyAsciiJpExtractor = { ... }.
var WeeklyAsciiJpExtractor = &CustomExtractor{
	Domain: "weekly.ascii.jp",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "article h1"},
			{Selector: "h1[itemprop=\"headline\"]"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "p.author"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "p.date"},
			{Selector: "meta[name=\"odate\"]", Attribute: "value"},
		},

		// format: 'YYYY年MM月DD日 HH:mm' in JavaScript - note: Go implementation handles format in date cleaner
		Format: "YYYY年MM月DD日 HH:mm",

		// timezone: 'Asia/Tokyo' in JavaScript - note: Go implementation handles timezone in date cleaner
		Timezone: "Asia/Tokyo",
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div#contents_detail"},
			{"div.article"},
		},
	},
}

// GetWeeklyAsciiJpExtractor returns the Weekly ASCII Japan custom extractor.
func GetWeeklyAsciiJpExtractor() *CustomExtractor {
	return WeeklyAsciiJpExtractor
}
