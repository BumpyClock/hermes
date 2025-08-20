// ABOUTME: Types and interfaces for extractor system
// ABOUTME: Core type definitions for selector processing and field extraction

package extractors

import (
	"github.com/PuerkitoBio/goquery"
)

// SelectOptions contains parameters for field selection
type SelectOptions struct {
	Doc            *goquery.Document
	Type           string
	ExtractionOpts interface{}
	ExtractHTML    bool
	URL            string
}

// TransformFunc is a function type for DOM transformations
type TransformFunc func(*goquery.Selection, *goquery.Document) interface{}

// ExtractorOptions contains options for field extraction
type ExtractorOptions struct {
	StripUnlikelyCandidates bool
	WeightNodes             bool
	CleanConditionally      bool
	URL                     string
	Content                 string
	Title                   string
}

// SelectorEntry represents a parsed selector with metadata
type SelectorEntry struct {
	Selector        string
	Attribute       string
	TransformFunc   func(string) string
	IsMultiSelector bool
	IsAttribute     bool
}