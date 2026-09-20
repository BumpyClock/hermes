// Package extractors defines shared, immutable extraction rule types.
package extractors

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

// DefinitionExtractor is a complete site rule loaded from an external definition snapshot.
type DefinitionExtractor struct {
	Domain        string
	Title         *FieldExtractor
	Author        *FieldExtractor
	Content       *ContentExtractor
	DatePublished *FieldExtractor
	LeadImageURL  *FieldExtractor
}

// FieldExtractor defines ordered text, attribute, or bounded text-capture alternatives.
type FieldExtractor struct {
	Selectors []SelectorEntry
}

// ContentExtractor defines selected content and its cleanup and transform program.
type ContentExtractor struct {
	Selectors             []ContentSelectorGroup
	Clean                 []string
	Preserve              []string
	DisableDefaultCleaner bool
	OrderedTransforms     ContentTransformProgram
}

// ContentTransformProgram copies selected content and executes an immutable YAML program.
type ContentTransformProgram interface {
	CopyAndExecute(context.Context, *goquery.Selection, string) (*goquery.Selection, error)
}

// ContentSelectorGroup combines CSS selectors in source order without duplicate elements.
type ContentSelectorGroup []string

// SelectorEntry selects text or an attribute from the first matching element.
type SelectorEntry struct {
	Selector  string
	Attribute string
	Capture   *TextCapture
}

// TextCapture extracts one bounded RE2 capture from selected element text.
type TextCapture struct {
	Pattern                             *regexp.Regexp
	Group, MaxBytes, MaxNodes, MaxDepth int
}

// ErrMetadataCaptureLimit identifies a bounded metadata text capture failure.
var ErrMetadataCaptureLimit = errors.New("metadata text capture resource limit exceeded")

// MetadataCaptureError identifies a selector capture that could not be evaluated safely.
type MetadataCaptureError struct {
	Selector string
	Err      error
}

func (e *MetadataCaptureError) Error() string {
	return fmt.Sprintf("metadata text capture %q: %v", e.Selector, e.Err)
}

func (e *MetadataCaptureError) Unwrap() error { return e.Err }
