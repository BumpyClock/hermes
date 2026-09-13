package parser

import (
	"net/http"
	"time"

	"github.com/BumpyClock/hermes/internal/definitions"
)

// ParserOptions configures the parser behavior.
type ParserOptions struct {
	Definitions           *definitions.Snapshot
	DefinitionsConfigured bool
	Fallback              bool              // Use generic extractor as fallback
	ContentType           string            // Output format: "html", "markdown", "text"
	Headers               map[string]string // Custom HTTP headers
	HTTPClient            *http.Client      // HTTP client to use for requests
	AllowPrivateNetworks  bool              // Allow SSRF to private networks (default: false)
}

// Result contains the extracted article data.
type Result struct {
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	Author        string     `json:"author"`
	DatePublished *time.Time `json:"date_published"`
	LeadImageURL  string     `json:"lead_image_url"`
	Dek           string     `json:"dek"`
	NextPageURL   string     `json:"next_page_url"`
	URL           string     `json:"url"`
	Domain        string     `json:"domain"`
	Excerpt       string     `json:"excerpt"`
	WordCount     int        `json:"word_count"`
	Direction     string     `json:"direction"`
	TotalPages    int        `json:"total_pages"`
	RenderedPages int        `json:"rendered_pages"`
	ExtractorUsed string     `json:"extractor_used,omitempty"`

	// Site metadata fields
	SiteName    string `json:"site_name"`
	SiteTitle   string `json:"site_title"`
	SiteImage   string `json:"site_image"`
	Favicon     string `json:"favicon"`
	Description string `json:"description"`
	Language    string `json:"language"`
	ThemeColor  string `json:"theme_color,omitempty"`

	// Video metadata fields
	VideoURL      string                 `json:"video_url,omitempty"`
	VideoMetadata map[string]interface{} `json:"video_metadata,omitempty"`
}

// DefaultParserOptions returns default parser options.
func DefaultParserOptions() *ParserOptions {
	return &ParserOptions{
		Fallback:    true,
		ContentType: "html",
		Headers:     make(map[string]string),
	}
}
