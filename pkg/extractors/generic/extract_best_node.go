// ABOUTME: Main orchestrator for content extraction connecting stripping, conversion, scoring and candidate selection
// ABOUTME: Implements JavaScript extractBestNode function with 100% compatibility

package generic

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/hermes/pkg/utils/dom"
)

// ExtractBestNodeOptions represents configuration options for content extraction
type ExtractBestNodeOptions struct {
	StripUnlikelyCandidates bool
	WeightNodes             bool
}

// ExtractBestNode extracts the content most likely to be article text using a variety of scoring techniques.
//
// The function orchestrates the complete extraction pipeline:
// 1. Optionally strips unlikely candidates (comments, ads, etc.)
// 2. Converts elements to paragraphs for better scoring
// 3. Scores all content based on various signals
// 4. Finds and returns the top candidate element
//
// This is a direct port of the JavaScript extractBestNode function with 100% compatibility.
//
// Parameters:
//   - doc: A goquery Document representing the DOM to extract from
//   - opts: ExtractBestNodeOptions with configuration flags
//     - StripUnlikelyCandidates: If true, remove elements that match exclusion criteria
//     - WeightNodes: If true, use classNames and IDs to determine node worthiness
//
// Returns:
//   - *goquery.Selection: The top candidate element, or nil if no suitable content found
func ExtractBestNode(doc *goquery.Document, opts ExtractBestNodeOptions) *goquery.Selection {
	// Step 1: Conditionally strip unlikely candidates
	if opts.StripUnlikelyCandidates {
		doc = dom.StripUnlikelyCandidates(doc)
	}

	// Step 2: Convert elements to paragraphs for better scoring
	doc = dom.ConvertToParagraphs(doc)

	// Step 3: Score all content using the scoring system
	dom.ScoreContent(doc, opts.WeightNodes)

	// Step 4: Find and return the top candidate
	topCandidate := dom.FindTopCandidate(doc)

	return topCandidate
}