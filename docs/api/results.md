# Results API Reference

This document covers the `Result` structure and related functionality for working with extracted content data.

## Table of Contents

- [Result Structure](#result-structure)
- [Field Descriptions](#field-descriptions)
- [Result Methods](#result-methods)
- [Output Formats](#output-formats)
- [Error Handling](#error-handling)
- [Result Processing](#result-processing)
- [JSON Serialization](#json-serialization)

## Result Structure

The `Result` struct contains all extracted content and metadata from a web page.

```go
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
    Description    string                `json:"description"`
    Language       string                `json:"language"`
    
    Error          bool                   `json:"error,omitempty"`
    Message        string                 `json:"message,omitempty"`
}
```

## Field Descriptions

### Core Content Fields

#### Title
```go
Title string `json:"title"`
```
The main headline or title of the article.

**Example:**
```go
result.Title // "Scientists Discover New Species in Deep Ocean"
```

#### Content
```go
Content string `json:"content"`
```
The main article content, formatted according to the specified `ContentType`.

**HTML Format:**
```html
<p>Scientists have discovered a new species...</p>
<h2>Research Details</h2>
<p>The research team, led by Dr. Smith...</p>
```

**Markdown Format:**
```markdown
Scientists have discovered a new species...

## Research Details

The research team, led by Dr. Smith...
```

#### Author
```go
Author string `json:"author"`
```
Article author(s). Multiple authors are comma-separated.

**Examples:**
```go
result.Author // "John Smith"
result.Author // "Jane Doe, Bob Johnson"
```

#### DatePublished
```go
DatePublished *time.Time `json:"date_published"`
```
Publication date and time. Uses Go's `time.Time` for precise date handling.

**Usage:**
```go
if result.DatePublished != nil {
    fmt.Printf("Published: %s\n", result.DatePublished.Format("January 2, 2006"))
    fmt.Printf("Unix timestamp: %d\n", result.DatePublished.Unix())
}
```

### Media Fields

#### LeadImageURL
```go
LeadImageURL string `json:"lead_image_url"`
```
URL of the main article image (hero image, featured image, etc.).

**Example:**
```go
result.LeadImageURL // "https://example.com/images/article-hero.jpg"
```

#### Dek
```go
Dek string `json:"dek"`
```
Article subtitle, deck, or subheadline providing additional context.

**Example:**
```go
result.Dek // "Breakthrough research reveals unexpected biodiversity"
```

### Navigation Fields

#### NextPageURL
```go
NextPageURL string `json:"next_page_url"`
```
URL of the next page for multi-page articles.

**Usage:**
```go
if result.NextPageURL != "" {
    fmt.Printf("Next page: %s\n", result.NextPageURL)
    // Can be used for pagination handling
}
```

#### URL
```go
URL string `json:"url"`
```
The canonical URL of the article.

#### Domain
```go
Domain string `json:"domain"`
```
The domain name extracted from the URL.

**Example:**
```go
result.Domain // "www.example.com"
```

### Content Analysis Fields

#### Excerpt
```go
Excerpt string `json:"excerpt"`
```
Short summary or excerpt of the article content (typically first 150-200 characters).

**Example:**
```go
result.Excerpt // "Scientists have discovered a new species of deep-sea fish..."
```

#### WordCount
```go
WordCount int `json:"word_count"`
```
Total word count of the extracted content.

**Usage:**
```go
fmt.Printf("Article length: %d words\n", result.WordCount)
readingTime := result.WordCount / 200 // Assume 200 WPM
fmt.Printf("Reading time: %d minutes\n", readingTime)
```

#### Direction
```go
Direction string `json:"direction"`
```
Text direction for internationalization support.

**Values:**
- `"ltr"` - Left-to-right (English, etc.)
- `"rtl"` - Right-to-left (Arabic, Hebrew, etc.)

### Pagination Fields

#### TotalPages
```go
TotalPages int `json:"total_pages"`
```
Total number of pages in multi-page articles.

#### RenderedPages
```go
RenderedPages int `json:"rendered_pages"`
```
Number of pages that were actually fetched and rendered.

**Usage:**
```go
if result.TotalPages > result.RenderedPages {
    fmt.Printf("Partial content: %d of %d pages\n", 
        result.RenderedPages, result.TotalPages)
}
```

### Site Metadata Fields

#### Description
```go
Description string `json:"description"`
```
Site-level description extracted from meta tags and structured data.

**Example:**
```go
result.Description // "NPR delivers breaking news and analysis on politics, business, science, and more."
```

**Extraction Sources:**
- `<meta name="description" content="..." />`
- `<meta property="og:description" content="..." />`
- `<meta name="twitter:description" content="..." />`
- JSON-LD structured data (`WebSite`, `Organization`, `NewsMediaOrganization`)

**Validation:**
- Filters out article-specific descriptions
- Removes descriptions containing URLs
- Validates minimum length (10 characters)
- Excludes spam or promotional content

#### Language
```go
Language string `json:"language"`
```
Content language code extracted from HTML attributes, meta tags, and structured data.

**Example:**
```go
result.Language // "en-US"
```

**Extraction Sources:**
- `<html lang="en-US">` or `<html xml:lang="en-US">`
- `<meta property="og:locale" content="en_US" />`
- `<meta name="content-language" content="en-US" />`
- `<meta http-equiv="Content-Language" content="en-US" />`
- JSON-LD structured data (`inLanguage`, `@language`, `contentLanguage`)

**Language Code Format:**
- Simple codes: `"en"`, `"fr"`, `"es"`
- Locale codes: `"en-US"`, `"fr-CA"`, `"pt-BR"`
- Normalized from underscore format: `"en_US"` → `"en-US"`
- Proper case handling: `"en-us"` → `"en-US"`

### Metadata Fields

#### ExtractorUsed
```go
ExtractorUsed string `json:"extractor_used,omitempty"`
```
Name of the extractor that was used for content extraction.

**Values:**
- `"custom:www.nytimes.com"` - Custom extractor
- `"generic"` - Generic fallback extractor

#### Extended
```go
Extended map[string]interface{} `json:"extended,omitempty"`
```
Custom fields extracted via extended field extractors.

**Example:**
```go
result.Extended["category"] // "Technology"
result.Extended["tags"]     // []string{"AI", "Machine Learning"}
result.Extended["reading_time"] // "5 min read"
```

### Error Fields

#### Error
```go
Error bool `json:"error,omitempty"`
```
Indicates whether an error occurred during extraction.

#### Message
```go
Message string `json:"message,omitempty"`
```
Error message if extraction failed.

## Result Methods

### IsError

```go
func (r *Result) IsError() bool
```

Checks if the result contains an error.

**Usage:**
```go
result, err := parser.Parse("https://example.com", nil)
if err != nil {
    log.Fatal("Network error:", err)
}

if result.IsError() {
    log.Printf("Extraction error: %s\n", result.Message)
    return
}

// Safe to use result
fmt.Println(result.Title)
```

### FormatMarkdown

```go
func (r *Result) FormatMarkdown() string
```

Formats the result as markdown with metadata header.

**Returns:**
```markdown
# Article Title

**Author:** John Doe
**Date:** 2024-01-15T10:30:00Z
**URL:** https://example.com/article
**Language:** en-US
**Description:** Site description from meta tags or structured data

Article content in markdown format...
```

**Usage:**
```go
markdown := result.FormatMarkdown()
err := os.WriteFile("article.md", []byte(markdown), 0644)
```

## Output Formats

### HTML Output

When `ContentType: "html"` is specified:

```go
result.Content // Clean HTML content
```

**Features:**
- Preserves HTML structure and formatting
- Includes semantic tags (h1-h6, p, ul, ol, blockquote, etc.)
- Absolute URLs for links and images
- Cleaned of advertisements and navigation

**Example:**
```html
<h2>Section Heading</h2>
<p>Article paragraph with <strong>emphasis</strong> and <a href="https://example.com">links</a>.</p>
<ul>
  <li>List item one</li>
  <li>List item two</li>
</ul>
<blockquote>
  <p>Important quote or excerpt</p>
</blockquote>
```

### Markdown Output

When `ContentType: "markdown"` is specified:

```go
result.Content // Markdown formatted content
```

**Features:**
- Clean markdown syntax
- Preserves formatting and structure
- Suitable for documentation systems
- Human-readable plain text

**Example:**
```markdown
## Section Heading

Article paragraph with **emphasis** and [links](https://example.com).

- List item one
- List item two

> Important quote or excerpt
```

### Text Output

When `ContentType: "text"` is specified:

```go
result.Content // Plain text content
```

**Features:**
- All HTML tags removed
- Basic formatting preserved (line breaks, paragraphs)
- Smallest file size
- Suitable for text analysis

**Example:**
```
Section Heading

Article paragraph with emphasis and links.

- List item one
- List item two

Important quote or excerpt
```

### JSON Output

Complete result structure serialized as JSON:

```json
{
  "title": "Article Title",
  "content": "Article content...",
  "author": "John Doe",
  "date_published": "2024-01-15T10:30:00Z",
  "lead_image_url": "https://example.com/image.jpg",
  "dek": "Article subtitle",
  "url": "https://example.com/article",
  "domain": "example.com",
  "excerpt": "Article excerpt...",
  "word_count": 1250,
  "direction": "ltr",
  "total_pages": 1,
  "rendered_pages": 1,
  "extractor_used": "custom:example.com",
  "description": "Example site's description from meta tags",
  "language": "en-US",
  "extended": {
    "category": "Technology",
    "tags": ["AI", "Science"]
  }
}
```

## Error Handling

### Error Result Structure

When extraction fails:

```go
result := &Result{
    Error: true,
    Message: "Failed to extract content: no title found",
    URL: "https://example.com/article",
    Domain: "example.com",
}
```

### Error Checking Patterns

#### Basic Error Checking

```go
result, err := parser.Parse(url, nil)
if err != nil {
    return fmt.Errorf("network error: %w", err)
}

if result.IsError() {
    return fmt.Errorf("extraction error: %s", result.Message)
}

// Use result safely
processResult(result)
```

#### Graceful Error Handling

```go
result, err := parser.Parse(url, nil)
if err != nil {
    log.Printf("Failed to fetch %s: %v", url, err)
    return nil
}

if result.IsError() {
    log.Printf("Failed to extract content from %s: %s", url, result.Message)
    return nil
}

// Check for minimum content requirements
if result.Title == "" || len(result.Content) < 100 {
    log.Printf("Insufficient content extracted from %s", url)
    return nil
}

return result
```

#### Batch Error Handling

```go
urls := []string{"https://example.com/1", "https://example.com/2"}
results := make([]*Result, 0, len(urls))
errors := make([]error, 0)

for _, url := range urls {
    result, err := parser.Parse(url, nil)
    if err != nil {
        errors = append(errors, fmt.Errorf("failed to process %s: %w", url, err))
        continue
    }
    
    if result.IsError() {
        errors = append(errors, fmt.Errorf("extraction failed for %s: %s", url, result.Message))
        continue
    }
    
    results = append(results, result)
}

fmt.Printf("Successfully processed %d of %d URLs\n", len(results), len(urls))
if len(errors) > 0 {
    fmt.Printf("Errors: %v\n", errors)
}
```

## Result Processing

### Content Analysis

```go
func AnalyzeResult(result *Result) {
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Author: %s\n", result.Author)
    fmt.Printf("Word count: %d\n", result.WordCount)
    
    if result.DatePublished != nil {
        age := time.Since(*result.DatePublished)
        fmt.Printf("Age: %.0f days\n", age.Hours()/24)
    }
    
    readingTime := result.WordCount / 200 // 200 WPM average
    fmt.Printf("Estimated reading time: %d minutes\n", readingTime)
    
    if result.Extended != nil {
        if category, ok := result.Extended["category"].(string); ok {
            fmt.Printf("Category: %s\n", category)
        }
        
        if tags, ok := result.Extended["tags"].([]string); ok {
            fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
        }
    }
}
```

### Content Validation

```go
func ValidateResult(result *Result) error {
    if result.IsError() {
        return errors.New(result.Message)
    }
    
    if result.Title == "" {
        return errors.New("no title extracted")
    }
    
    if len(result.Content) < 100 {
        return errors.New("insufficient content extracted")
    }
    
    if result.WordCount < 50 {
        return errors.New("content too short")
    }
    
    return nil
}
```

### Content Transformation

```go
func TransformResult(result *Result, format string) (string, error) {
    switch format {
    case "html":
        return result.Content, nil
        
    case "markdown":
        return result.FormatMarkdown(), nil
        
    case "text":
        // Convert HTML to plain text
        return html2text.FromString(result.Content)
        
    case "json":
        data, err := json.MarshalIndent(result, "", "  ")
        return string(data), err
        
    case "summary":
        // Create a summary
        return fmt.Sprintf("%s\n\nBy %s | %d words | %s\n\n%s",
            result.Title,
            result.Author,
            result.WordCount,
            result.DatePublished.Format("Jan 2, 2006"),
            result.Excerpt), nil
            
    default:
        return "", fmt.Errorf("unsupported format: %s", format)
    }
}
```

## JSON Serialization

### Custom JSON Marshaling

```go
func (r *Result) MarshalJSON() ([]byte, error) {
    type Alias Result
    return json.Marshal(&struct {
        *Alias
        DatePublishedString string `json:"date_published_string,omitempty"`
    }{
        Alias: (*Alias)(r),
        DatePublishedString: func() string {
            if r.DatePublished != nil {
                return r.DatePublished.Format(time.RFC3339)
            }
            return ""
        }(),
    })
}
```

### JSON Output Examples

**Minimal Result:**
```json
{
  "title": "Article Title",
  "content": "Article content...",
  "url": "https://example.com/article",
  "domain": "example.com",
  "word_count": 450
}
```

**Complete Result:**
```json
{
  "title": "Scientists Discover New Deep-Sea Species",
  "content": "<p>Scientists have discovered...</p>",
  "author": "Dr. Jane Smith, Prof. Bob Johnson",
  "date_published": "2024-01-15T10:30:00Z",
  "lead_image_url": "https://example.com/images/deep-sea.jpg",
  "dek": "Breakthrough research reveals unexpected biodiversity",
  "url": "https://example.com/science/deep-sea-discovery",
  "domain": "example.com",
  "excerpt": "Scientists have discovered a new species of deep-sea fish that challenges our understanding of marine biology...",
  "word_count": 1250,
  "direction": "ltr",
  "total_pages": 1,
  "rendered_pages": 1,
  "extractor_used": "custom:example.com",
  "description": "Leading scientific research and news from the world's top researchers",
  "language": "en-US",
  "extended": {
    "category": "Science",
    "tags": ["marine biology", "deep sea", "biodiversity"],
    "reading_time": "6 min read",
    "social_shares": 42
  }
}
```

**Error Result:**
```json
{
  "url": "https://example.com/not-found",
  "domain": "example.com",
  "error": true,
  "message": "Failed to extract content: 404 Not Found"
}
```