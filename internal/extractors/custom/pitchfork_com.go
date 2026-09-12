// ABOUTME: Pitchfork custom extractor for music reviews and embedded media
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/pitchfork.com/index.js

package custom

// PitchforkCustomExtractor provides the custom extraction rules for pitchfork.com
// JavaScript equivalent: export const PitchforkComExtractor = { ... }.
var PitchforkCustomExtractor = &CustomExtractor{
	Domain: "pitchfork.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:title\"]", Attribute: "value"},
			{Selector: "title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:author\"]", Attribute: "value"},
			{Selector: ".authors-detail__display-name"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "div[class^=\"InfoSliceWrapper-\"]"},
			{Selector: ".pub-date", Attribute: "datetime"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:description\"]", Attribute: "value"},
			{Selector: ".review-detail__abstract"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			{Selector: ".single-album-tombstone__art img", Attribute: "src"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.body__inner-container"},
			{".review-detail__text"},
		},
	},

	// Extended fields for music review scores
	Extend: map[string]*FieldExtractor{
		"score": {
			Selectors: []SelectorEntry{
				{Selector: "p[class*=\"Rating\"]"},
				{Selector: ".score"},
			},
		},
	},
}

// GetPitchforkExtractor returns the Pitchfork custom extractor.
func GetPitchforkExtractor() *CustomExtractor {
	return PitchforkCustomExtractor
}
