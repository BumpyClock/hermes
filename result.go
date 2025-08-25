package hermes

import (
	"fmt"
	"strings"
	"time"
)

// Result contains the extracted content from a web page.
// All fields are read-only and represent the parsed article data.
type Result struct {
	// Core content fields
	URL           string     `json:"url"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	Author        string     `json:"author,omitempty"`
	DatePublished *time.Time `json:"date_published,omitempty"`
	
	// Media and metadata
	LeadImageURL  string `json:"lead_image_url,omitempty"`
	Dek           string `json:"dek,omitempty"`
	Domain        string `json:"domain"`
	Excerpt       string `json:"excerpt,omitempty"`
	
	// Content metrics
	WordCount     int    `json:"word_count"`
	Direction     string `json:"direction,omitempty"`
	TotalPages    int    `json:"total_pages,omitempty"`
	RenderedPages int    `json:"rendered_pages,omitempty"`
	
	// Site information
	SiteName    string `json:"site_name,omitempty"`
	Description string `json:"description,omitempty"`
	Language    string `json:"language,omitempty"`
}

// FormatMarkdown formats the result as Markdown with metadata header.
// This is useful for saving the content in a human-readable format.
//
// Example output:
//
//	# Article Title
//	
//	## Metadata
//	**Author:** John Doe
//	**Date:** 2024-01-01
//	**URL:** https://example.com/article
//	
//	## Content
//	Article content here...
func (r *Result) FormatMarkdown() string {
	var sb strings.Builder
	
	// Title
	if r.Title != "" {
		sb.WriteString("# ")
		sb.WriteString(r.Title)
		sb.WriteString("\n\n")
	}
	
	// Metadata section
	hasMetadata := r.Author != "" || r.DatePublished != nil || r.URL != "" || r.SiteName != ""
	if hasMetadata {
		sb.WriteString("## Metadata\n\n")
		
		if r.Author != "" {
			sb.WriteString("**Author:** ")
			sb.WriteString(r.Author)
			sb.WriteString("\n")
		}
		
		if r.DatePublished != nil {
			sb.WriteString("**Date:** ")
			sb.WriteString(r.DatePublished.Format("2006-01-02"))
			sb.WriteString("\n")
		}
		
		if r.URL != "" {
			sb.WriteString("**URL:** ")
			sb.WriteString(r.URL)
			sb.WriteString("\n")
		}
		
		if r.SiteName != "" {
			sb.WriteString("**Site:** ")
			sb.WriteString(r.SiteName)
			sb.WriteString("\n")
		}
		
		if r.Language != "" {
			sb.WriteString("**Language:** ")
			sb.WriteString(r.Language)
			sb.WriteString("\n")
		}
		
		if r.WordCount > 0 {
			sb.WriteString("**Word Count:** ")
			sb.WriteString(fmt.Sprintf("%d", r.WordCount))
			sb.WriteString("\n")
		}
		
		sb.WriteString("\n")
	}
	
	// Description/Excerpt
	if r.Description != "" {
		sb.WriteString("## Description\n\n")
		sb.WriteString(r.Description)
		sb.WriteString("\n\n")
	} else if r.Excerpt != "" {
		sb.WriteString("## Excerpt\n\n")
		sb.WriteString(r.Excerpt)
		sb.WriteString("\n\n")
	}
	
	// Main content
	if r.Content != "" {
		sb.WriteString("## Content\n\n")
		sb.WriteString(r.Content)
	}
	
	return sb.String()
}

// IsEmpty returns true if the result contains no meaningful content
func (r *Result) IsEmpty() bool {
	return r.Title == "" && r.Content == ""
}

// HasAuthor returns true if author information is available
func (r *Result) HasAuthor() bool {
	return r.Author != ""
}

// HasDate returns true if publication date is available
func (r *Result) HasDate() bool {
	return r.DatePublished != nil
}

// HasImage returns true if a lead image is available
func (r *Result) HasImage() bool {
	return r.LeadImageURL != ""
}