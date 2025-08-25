package parser

import (
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Parser is the main interface for content extraction
type Parser interface {
	Parse(url string, opts *ParserOptions) (*Result, error)
	ParseHTML(html string, url string, opts *ParserOptions) (*Result, error)
}

// ParserOptions configures the parser behavior
type ParserOptions struct {
	FetchAllPages        bool              // Fetch and merge multi-page articles
	Fallback             bool              // Use generic extractor as fallback
	ContentType          string            // Output format: "html", "markdown", "text"
	Headers              map[string]string         // Custom HTTP headers
	CustomExtractor      *CustomExtractor          // Custom extraction rules
	Extend               map[string]ExtractorFunc  // Extended fields
	HTTPClient           *http.Client              // HTTP client to use for requests
	AllowPrivateNetworks bool                      // Allow SSRF to private networks (default: false)
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
	ExtractorUsed  string                `json:"extractor_used,omitempty"`
	Extended       map[string]interface{} `json:"extended,omitempty"`
	
	// Site metadata fields
	SiteName       string                `json:"site_name"`
	SiteTitle      string                `json:"site_title"`
	SiteImage      string                `json:"site_image"`
	Favicon        string                `json:"favicon"`
	Description    string                `json:"description"`
	Language       string                `json:"language"`
	
	// Error handling fields for JS compatibility
	Error   bool   `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// Extractor defines the interface for content extractors
type Extractor interface {
	Extract(doc *goquery.Document, url string, opts *ExtractorOptions) (*Result, error)
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

// DefaultParserOptions returns default parser options
func DefaultParserOptions() *ParserOptions {
	return &ParserOptions{
		FetchAllPages: true,
		Fallback:      true,
		ContentType:   "html",
	}
}

// DefaultExtractorOptions returns default extractor options
func DefaultExtractorOptions() *ExtractorOptions {
	return &ExtractorOptions{
		Fallback:    true,
		ContentType: "html",
		MetaCache:   make(map[string]string),
	}
}

// FormatMarkdown formats the result as markdown with metadata header
func (r *Result) FormatMarkdown() string {
	var sb strings.Builder
	
	// Add title as H1
	if r.Title != "" {
		sb.WriteString("# ")
		sb.WriteString(r.Title)
		sb.WriteString("\n\n")
	}
	
	// Add site metadata section
	hasSiteMetadata := r.SiteName != "" || r.SiteTitle != "" || r.SiteImage != "" || r.Favicon != "" || r.Description != "" || r.Language != ""
	if hasSiteMetadata {
		sb.WriteString("## Site Information\n\n")
		
		if r.SiteName != "" {
			sb.WriteString("**Site:** ")
			sb.WriteString(r.SiteName)
			sb.WriteString("\n")
		}
		
		if r.SiteTitle != "" {
			sb.WriteString("**Site Title:** ")
			sb.WriteString(r.SiteTitle)
			sb.WriteString("\n")
		}
		
		if r.SiteImage != "" {
			sb.WriteString("**Site Image:** ")
			sb.WriteString(r.SiteImage)
			sb.WriteString("\n")
		}
		
		if r.Favicon != "" {
			sb.WriteString("**Favicon:** ")
			sb.WriteString(r.Favicon)
			sb.WriteString("\n")
		}
		
		if r.Description != "" {
			sb.WriteString("**Description:** ")
			sb.WriteString(r.Description)
			sb.WriteString("\n")
		}
		
		if r.Language != "" {
			sb.WriteString("**Language:** ")
			sb.WriteString(r.Language)
			sb.WriteString("\n")
		}
		
		sb.WriteString("\n")
	}
	
	// Add article metadata
	hasArticleMetadata := r.Author != "" || r.DatePublished != nil || r.URL != ""
	if hasArticleMetadata {
		sb.WriteString("## Article Information\n\n")
		
		if r.Author != "" {
			sb.WriteString("**Author:** ")
			sb.WriteString(r.Author)
			sb.WriteString("\n")
		}
		
		if r.DatePublished != nil {
			sb.WriteString("**Date:** ")
			sb.WriteString(r.DatePublished.Format(time.RFC3339))
			sb.WriteString("\n")
		}
		
		if r.URL != "" {
			sb.WriteString("**URL:** ")
			sb.WriteString(r.URL)
			sb.WriteString("\n")
		}
		
		sb.WriteString("\n")
	}
	
	// Add content section
	if r.Content != "" {
		sb.WriteString("## Content\n\n")
		sb.WriteString(r.Content)
	}
	
	return sb.String()
}

// PoolStats is deprecated - kept for backward compatibility
// Object pooling has been removed in favor of simplicity
type PoolStats struct {
	// All fields are deprecated and return zero values
	ResultsCreated   int64
	ResultsReused    int64
	BuffersCreated   int64
	BuffersReused    int64
	ParsersCreated   int64
	ParsersReused    int64
	LastReset        time.Time
}