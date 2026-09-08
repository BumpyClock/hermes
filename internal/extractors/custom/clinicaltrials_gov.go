// ABOUTME: ClinicalTrials.gov extractor for government clinical trial database
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/clinicaltrials.gov/index.js

package custom

// ClinicaltrialsGovExtractor provides the custom extraction rules for clinicaltrials.gov
// JavaScript equivalent: export const ClinicaltrialsGovExtractor = { ... }.
var ClinicaltrialsGovExtractor = &CustomExtractor{
	Domain: "clinicaltrials.gov",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.tr-solo_record"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "div#sponsor.tr-info-text"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			// JavaScript: selectors: ['div:has(> span.term[data-term="Last Update Posted"])']
			{Selector: `div:has(> span.term[data-term="Last Update Posted"])`},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div#tab-body"},
		},

		// Clean selectors
		Clean: []string{
			".usa-alert> img",
		},
	},
}

// GetClinicaltrialsGovExtractor returns the ClinicalTrials.gov custom extractor.
func GetClinicaltrialsGovExtractor() *CustomExtractor {
	return ClinicaltrialsGovExtractor
}
