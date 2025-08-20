package parser

import (
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Parser is the main interface for content extraction
type Parser interface {
	Parse(url string, opts ParserOptions) (*Result, error)
	ParseHTML(html string, url string, opts ParserOptions) (*Result, error)
}

// ParserOptions configures the parser behavior
type ParserOptions struct {
	FetchAllPages   bool              // Fetch and merge multi-page articles
	Fallback        bool              // Use generic extractor as fallback
	ContentType     string            // Output format: "html", "markdown", "text"
	Headers         map[string]string // Custom HTTP headers
	CustomExtractor *CustomExtractor  // Custom extraction rules
	Extend          map[string]ExtractorFunc // Extended fields
}

// Result contains the extracted article data
type Result struct {
	Title          string                 `json:"title"`
	Content        string                 `json:"content"`
	Author         string                 `json:"author"`
	DatePublished  *time.Time            `json:"date_published"`
	LeadImageURL   string                `json:"lead_image_url"`
	Dek            string                `json:"dek"`
	NextPageURL    string                `json:"next_page_url"`
	URL            string                `json:"url"`
	Domain         string                `json:"domain"`
	Excerpt        string                `json:"excerpt"`
	WordCount      int                   `json:"word_count"`
	Direction      string                `json:"direction"`
	TotalPages     int                   `json:"total_pages"`
	RenderedPages  int                   `json:"rendered_pages"`
	Extended       map[string]interface{} `json:"extended,omitempty"`
	
	// Error handling fields for JS compatibility
	Error   bool   `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// Extractor defines the interface for content extractors
type Extractor interface {
	Extract(doc *goquery.Document, url string, opts ExtractorOptions) (*Result, error)
	GetDomain() string
}

// ExtractorOptions configures individual extractors
type ExtractorOptions struct {
	URL         string
	HTML        string
	MetaCache   map[string]string
	Fallback    bool
	ContentType string
}

// CustomExtractor defines site-specific extraction rules
type CustomExtractor struct {
	Domain        string
	Title         FieldExtractor
	Author        FieldExtractor
	Content       ContentExtractor
	DatePublished FieldExtractor
	LeadImageURL  FieldExtractor
	Dek           FieldExtractor
	NextPageURL   FieldExtractor
	Excerpt       FieldExtractor
	Extend        map[string]FieldExtractor
}

// FieldExtractor defines extraction rules for a specific field
type FieldExtractor struct {
	Selectors      SelectorList  // Type-safe CSS selectors (replaces []interface{})
	SelectorsLegacy []interface{} `json:"selectors,omitempty"` // Deprecated: use Selectors instead
	AllowMultiple  bool
	DefaultCleaner bool
}

// ContentExtractor extends FieldExtractor with cleaning options
type ContentExtractor struct {
	FieldExtractor
	Clean      []string                   // Selectors to remove
	Transforms map[string]TransformFunc   // Element transformations
}

// TransformFunc modifies extracted elements
type TransformFunc func(*goquery.Selection) string

// ExtractorFunc is a custom extraction function
type ExtractorFunc func(*goquery.Document, string) (interface{}, error)