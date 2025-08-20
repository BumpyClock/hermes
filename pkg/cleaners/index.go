// ABOUTME: Index file for cleaners package providing registry and interface for all field cleaners
// ABOUTME: Provides a clean API for content cleaning that can be used independently or with extractors

package cleaners

import "github.com/PuerkitoBio/goquery"

// CleanerOptions represents unified options for all cleaner types
type CleanerOptions struct {
	URL            string
	Title          string
	Content        string
	Excerpt        string
	DefaultCleaner bool
}

// FieldCleaner represents the interface for field-specific cleaners
type FieldCleaner interface {
	// Clean cleans a field value (string, []string, etc.)
	Clean(value interface{}, opts CleanerOptions) interface{}
	
	// CleanSelection cleans a goquery selection (for HTML content)
	CleanSelection(selection *goquery.Selection, doc *goquery.Document, opts CleanerOptions) *goquery.Selection
}

// ContentCleaner implements FieldCleaner for content fields
type ContentCleaner struct{}

func (c *ContentCleaner) Clean(value interface{}, opts CleanerOptions) interface{} {
	// For content, we expect HTML strings
	return value
}

func (c *ContentCleaner) CleanSelection(selection *goquery.Selection, doc *goquery.Document, opts CleanerOptions) *goquery.Selection {
	contentOpts := ContentCleanOptions{
		CleanConditionally: true,
		Title:              opts.Title,
		URL:                opts.URL,
		DefaultCleaner:     &opts.DefaultCleaner,
	}
	return ExtractCleanNode(selection, doc, contentOpts)
}

// Registry of all available cleaners
var cleanerRegistry = map[string]FieldCleaner{
	"content": &ContentCleaner{},
	// Additional cleaners will be added here as they're implemented
}

// GetCleaner retrieves a cleaner by field type
func GetCleaner(fieldType string) FieldCleaner {
	return cleanerRegistry[fieldType]
}

// RegisterCleaner registers a new cleaner for a field type
func RegisterCleaner(fieldType string, cleaner FieldCleaner) {
	cleanerRegistry[fieldType] = cleaner
}

// Re-export the main content cleaning function and options for external use
// This allows other packages to import and use the content cleaner independently

// ExtractCleanNode is the main content cleaning function
// It can be used as a standalone utility or integrated with content extractors
var ExtractCleanNodeFunc = ExtractCleanNode

// ContentCleanOptions represents the configuration options for content cleaning
type ContentCleanOptionsStruct = ContentCleanOptions