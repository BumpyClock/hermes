# Extractors API Reference

Extractors are the core components responsible for extracting content from web pages. Hermes includes both generic extractors for general content extraction and custom extractors for site-specific optimization.

## Table of Contents

- [Overview](#overview)
- [Extractor Interface](#extractor-interface)
- [Custom Extractors](#custom-extractors)
- [Generic Extractors](#generic-extractors)
- [Field Extractors](#field-extractors)
- [Content Extractors](#content-extractors)
- [Extractor Registry](#extractor-registry)
- [Creating Custom Extractors](#creating-custom-extractors)

## Overview

The extractor system operates in a hierarchical manner:

1. **Custom Extractors**: Site-specific extractors optimized for particular domains
2. **Generic Extractors**: Fallback extractors that work on any HTML content
3. **Field Extractors**: Extract specific fields (title, author, date, etc.)

## Extractor Interface

### Base Extractor Interface

```go
type Extractor interface {
    Extract(doc *goquery.Document, url string, opts *ExtractorOptions) (*Result, error)
    GetDomain() string
}
```

### ExtractorOptions

```go
type ExtractorOptions struct {
    URL         string
    HTML        string
    MetaCache   map[string]string
    Fallback    bool
    ContentType string
}
```

Configuration for extractor operations.

#### Fields

- **URL** (string): Target URL being processed
- **HTML** (string): Raw HTML content
- **MetaCache** (map[string]string): Cached meta tag values
- **Fallback** (bool): Whether to use generic extractor as fallback
- **ContentType** (string): Desired output format

## Custom Extractors

Custom extractors provide site-specific extraction logic for optimal results on particular domains.

### CustomExtractor Structure

```go
type CustomExtractor struct {
    Domain           string                    `json:"domain"`
    SupportedDomains []string                  `json:"supportedDomains,omitempty"`
    Title            *FieldExtractor           `json:"title,omitempty"`
    Author           *FieldExtractor           `json:"author,omitempty"`
    Content          *ContentExtractor         `json:"content,omitempty"`
    DatePublished    *FieldExtractor           `json:"date_published,omitempty"`
    LeadImageURL     *FieldExtractor           `json:"lead_image_url,omitempty"`
    Dek              *FieldExtractor           `json:"dek,omitempty"`
    NextPageURL      *FieldExtractor           `json:"next_page_url,omitempty"`
    Excerpt          *FieldExtractor           `json:"excerpt,omitempty"`
    Extend           map[string]*FieldExtractor `json:"extend,omitempty"`
}
```

### Example: NY Times Extractor

```go
func GetNYTimesExtractor() *CustomExtractor {
    return &CustomExtractor{
        Domain: "www.nytimes.com",
        
        Title: &FieldExtractor{
            Selectors: []interface{}{
                `h1[data-testid="headline"]`,
                "h1.g-headline",
                `h1[itemprop="headline"]`,
                "h1.headline",
                "h1 .balancedHeadline",
            },
        },
        
        Author: &FieldExtractor{
            Selectors: []interface{}{
                []string{`meta[name="author"]`, "value"},
                ".g-byline",
                ".byline",
                []string{`meta[name="byl"]`, "value"},
            },
        },
        
        Content: &ContentExtractor{
            FieldExtractor: &FieldExtractor{
                Selectors: []interface{}{
                    "div.g-blocks",
                    `section[name="articleBody"]`,
                    "article#story",
                },
            },
            Clean: []string{
                ".ad",
                "header#story-header", 
                ".story-body-1 .lede.video",
                ".visually-hidden",
                "#newsletter-promo",
                ".promo",
                ".comments-button",
                ".hidden",
                ".comments",
                ".supplemental",
                ".nocontent",
                ".story-footer-links",
            },
        },
        
        DatePublished: &FieldExtractor{
            Selectors: []interface{}{
                []string{`meta[name="article:published_time"]`, "value"},
                []string{`meta[name="article:published"]`, "value"},
            },
        },
        
        LeadImageURL: &FieldExtractor{
            Selectors: []interface{}{
                []string{`meta[name="og:image"]`, "value"},
            },
        },
    }
}
```

### Built-in Custom Extractors

Hermes includes 150+ custom extractors for major publications. Examples include:

**News:**
- NY Times (`www.nytimes.com`)
- Washington Post (`www.washingtonpost.com`)
- CNN (`www.cnn.com`)
- The Guardian (`www.theguardian.com`)
- CBC (`www.cbc.ca`)

**Tech:**
- Ars Technica (`arstechnica.com`)
- The Verge (`www.theverge.com`)
- Wired (`www.wired.com`)

**Business:**
- Bloomberg (`www.bloomberg.com`)
- Reuters (`www.reuters.com`)

## Generic Extractors

Generic extractors provide fallback content extraction for any website using algorithmic content detection.

### GenericContentExtractor

```go
type GenericContentExtractor struct {
    DefaultOpts ExtractorOptions
}
```

Main generic extractor that implements the content scoring algorithm.

#### ExtractorOptions

```go
type ExtractorOptions struct {
    StripUnlikelyCandidates bool
    WeightNodes             bool
    CleanConditionally      bool
}
```

- **StripUnlikelyCandidates**: Remove elements unlikely to contain article content
- **WeightNodes**: Apply content scoring algorithm
- **CleanConditionally**: Apply conditional cleaning rules

#### Content Extraction Strategy

The generic extractor uses a cascading approach:

1. **Initial attempt** with strict options
2. **Fallback attempts** with progressively relaxed options
3. **Final attempt** with minimal restrictions

```go
func (e *GenericContentExtractor) Extract(params ExtractorParams, opts ExtractorOptions) string {
    // Try with current options
    node := e.GetContentNode(doc, params.Title, params.URL, opts)
    
    if NodeIsSufficient(node) {
        return e.CleanAndReturnNode(node, doc)
    }
    
    // Cascade through options, disabling them one by one
    if opts.StripUnlikelyCandidates {
        opts.StripUnlikelyCandidates = false
        // Retry...
    }
    
    if opts.WeightNodes {
        opts.WeightNodes = false
        // Retry...
    }
    
    if opts.CleanConditionally {
        opts.CleanConditionally = false
        // Retry...
    }
    
    return e.CleanAndReturnNode(node, doc)
}
```

### Content Scoring Algorithm

The generic extractor uses sophisticated scoring to identify article content:

- **Node scoring** based on content density and text length
- **Sibling merging** to combine related content blocks
- **Link density analysis** to avoid navigation-heavy content
- **Candidate selection** from top-scoring content nodes

## Field Extractors

Field extractors define how to extract specific fields from a document.

### FieldExtractor Structure

```go
type FieldExtractor struct {
    Selectors      []interface{} `json:"selectors"`      // CSS selectors or [selector, attr] pairs
    AllowMultiple  bool          `json:"allowMultiple"`  // Allow multiple values
    DefaultCleaner bool          `json:"defaultCleaner"` // Apply default field cleaner
    Format         string        `json:"format"`         // Date format (for date fields)
    Timezone       string        `json:"timezone"`       // Timezone (for date fields)
}
```

### Selector Types

Field extractors support multiple selector formats:

#### CSS Selector (String)

```go
Selectors: []interface{}{
    "h1.headline",
    "#article-title",
    ".post-title",
}
```

#### Attribute Extraction (Array)

```go
Selectors: []interface{}{
    []string{"meta[name='author']", "content"},
    []string{"time[datetime]", "datetime"},
    []string{"img.lead-image", "src"},
}
```

### Field-Specific Extractors

#### Title Extractor

```go
Title: &FieldExtractor{
    Selectors: []interface{}{
        "h1.headline",
        "h1[data-testid='headline']",
        ".article-title h1",
        "meta[property='og:title']", "content"},
    },
}
```

#### Author Extractor

```go
Author: &FieldExtractor{
    Selectors: []interface{}{
        []string{"meta[name='author']", "content"},
        ".byline .author",
        ".article-author",
        "[rel='author']",
    },
    AllowMultiple: true,
}
```

#### Date Extractor

```go
DatePublished: &FieldExtractor{
    Selectors: []interface{}{
        []string{"time[datetime]", "datetime"},
        []string{"meta[property='article:published_time']", "content"},
        ".publish-date",
    },
    Format: "2006-01-02T15:04:05Z07:00",
}
```

## Content Extractors

Content extractors extend field extractors with cleaning and transformation capabilities.

### ContentExtractor Structure

```go
type ContentExtractor struct {
    *FieldExtractor
    Clean          []string                       `json:"clean"`          // Selectors to remove
    Transforms     map[string]TransformFunction   `json:"transforms"`     // Element transformations
    DefaultCleaner bool                          `json:"defaultCleaner"` // Apply default content cleaner
}
```

### Transform Functions

Transform functions modify elements during extraction:

#### String Transform

```go
Transforms: map[string]TransformFunction{
    "noscript": &StringTransform{TargetTag: "div"},
    "blockquote cite": &StringTransform{TargetTag: "p"},
}
```

#### Function Transform

```go
Transforms: map[string]TransformFunction{
    "img.lazy": &FunctionTransform{
        Fn: func(node *goquery.Selection) error {
            src := node.AttrOr("data-src", "")
            if src != "" {
                node.SetAttr("src", src)
                node.RemoveAttr("data-src")
            }
            return nil
        },
    },
}
```

### Content Cleaning

Content extractors can specify elements to remove:

```go
Clean: []string{
    ".advertisement",
    ".related-articles",
    ".social-share",
    ".newsletter-signup",
    "script",
    "style",
    ".comments",
}
```

## Extractor Registry

The extractor registry manages all available extractors.

### ExtractorRegistry

```go
type ExtractorRegistry struct {
    extractors map[string]*CustomExtractor
}
```

### Registry Methods

#### Register

```go
func (r *ExtractorRegistry) Register(extractor *CustomExtractor)
```

Register a custom extractor.

#### Get

```go
func (r *ExtractorRegistry) Get(domain string) (*CustomExtractor, bool)
```

Retrieve an extractor by domain.

#### List

```go
func (r *ExtractorRegistry) List() []string
```

List all registered domains.

### Usage Example

```go
registry := custom.NewExtractorRegistry()

// Register custom extractor
extractor := &custom.CustomExtractor{
    Domain: "example.com",
    Title: &custom.FieldExtractor{
        Selectors: []interface{}{"h1.title"},
    },
}
registry.Register(extractor)

// Get extractor
if ext, exists := registry.Get("example.com"); exists {
    // Use extractor
}

// List all domains
domains := registry.List()
fmt.Printf("Supported domains: %v\n", domains)
```

## Creating Custom Extractors

### Step 1: Define the Extractor

```go
func GetCustomExtractor() *custom.CustomExtractor {
    return &custom.CustomExtractor{
        Domain: "example.com",
        
        Title: &custom.FieldExtractor{
            Selectors: []interface{}{
                "h1.article-title",
                ".headline h1",
            },
        },
        
        Author: &custom.FieldExtractor{
            Selectors: []interface{}{
                ".byline .author-name",
                []string{"meta[name='author']", "content"},
            },
        },
        
        Content: &custom.ContentExtractor{
            FieldExtractor: &custom.FieldExtractor{
                Selectors: []interface{}{
                    ".article-content",
                    ".post-body",
                },
            },
            Clean: []string{
                ".ads",
                ".related",
                ".social-share",
            },
        },
        
        DatePublished: &custom.FieldExtractor{
            Selectors: []interface{}{
                []string{"time.published", "datetime"},
                []string{"meta[property='article:published_time']", "content"},
            },
        },
    }
}
```

### Step 2: Register the Extractor

```go
func init() {
    registry := GetGlobalRegistry()
    registry.Register(GetCustomExtractor())
}
```

### Step 3: Test the Extractor

```go
func TestCustomExtractor(t *testing.T) {
    parser := parser.New()
    
    result, err := parser.Parse("https://example.com/article", &parser.ParserOptions{
        CustomExtractor: GetCustomExtractor(),
    })
    
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Title)
    assert.NotEmpty(t, result.Content)
}
```

## Site Metadata Extractors

Hermes includes specialized extractors for site-level metadata that provide information about the website or publication itself, rather than individual articles.

### GenericDescriptionExtractor

Extracts site-level descriptions from meta tags and JSON-LD structured data.

```go
type GenericDescriptionExtractor struct{}

func (extractor *GenericDescriptionExtractor) Extract(selection *goquery.Selection, pageURL string, metaCache []string) string
```

#### Extraction Strategy

The description extractor uses a priority-based approach:

1. **Meta tags** (highest priority)
   - `<meta name="description" content="..." />`
   - `<meta property="og:description" content="..." />`
   - `<meta name="twitter:description" content="..." />`
   - `<meta name="dc.description" content="..." />`

2. **JSON-LD structured data**
   - `WebSite` schema description
   - `Organization` schema description
   - `NewsMediaOrganization` schema description
   - Publisher description from Article schema

#### Content Validation

The extractor includes validation to ensure quality:

```go
func (extractor *GenericDescriptionExtractor) isValidDescription(description string) bool
```

- **Length validation**: 10-500 characters
- **URL detection**: Rejects descriptions containing URLs
- **Article-specific filtering**: Excludes descriptions starting with:
  - "In this article"
  - "This article"
  - "Read more about"
  - "Continue reading"
  - "Full story:"

#### Usage Example

```go
descriptionExtractor := &GenericDescriptionExtractor{}
description := descriptionExtractor.Extract(doc.Selection, "https://example.com", []string{})

if description != "" {
    fmt.Printf("Site description: %s\n", description)
}
```

### GenericLanguageExtractor

Extracts content language from HTML attributes, meta tags, and JSON-LD structured data.

```go
type GenericLanguageExtractor struct{}

func (extractor *GenericLanguageExtractor) Extract(selection *goquery.Selection, pageURL string, metaCache []string) string
```

#### Extraction Strategy

The language extractor checks multiple sources in priority order:

1. **HTML attributes** (highest priority)
   - `<html lang="en-US">`
   - `<html xml:lang="en-US">`

2. **Meta tags**
   - `<meta property="og:locale" content="en_US" />`
   - `<meta name="content-language" content="en-US" />`
   - `<meta http-equiv="Content-Language" content="en-US" />`
   - `<meta name="dc.language" content="en" />`

3. **JSON-LD structured data**
   - `inLanguage` property
   - `@language` property
   - `contentLanguage` property (Article schema)

#### Language Code Normalization

The extractor normalizes language codes to standard formats:

```go
func (extractor *GenericLanguageExtractor) normalizeLanguageCode(code string) string
```

**Normalization Rules:**
- Convert underscores to hyphens: `"en_US"` → `"en-US"`
- Proper case handling: `"en-us"` → `"en-US"`
- Special Chinese variants: `"zh-hans"` → `"zh-Hans"`
- Simple codes remain lowercase: `"en"`, `"fr"`, `"es"`

#### Language Validation

```go
func (extractor *GenericLanguageExtractor) isValidLanguageCode(code string) bool
```

Validates language codes against common patterns:
- Simple codes: 2 letters (e.g., `"en"`, `"fr"`)
- Locale codes: language-region format (e.g., `"en-US"`, `"pt-BR"`)
- Underscore variants: `"en_US"` (normalized to hyphen format)

#### Usage Example

```go
languageExtractor := &GenericLanguageExtractor{}
language := languageExtractor.Extract(doc.Selection, "https://example.com", []string{})

if language != "" {
    fmt.Printf("Content language: %s\n", language)
}
```

#### Common Language Codes

The extractor recognizes standard ISO language codes:

**Simple Codes:**
- `"en"` - English
- `"es"` - Spanish  
- `"fr"` - French
- `"de"` - German
- `"it"` - Italian
- `"pt"` - Portuguese
- `"ru"` - Russian
- `"ja"` - Japanese
- `"ko"` - Korean
- `"zh"` - Chinese
- `"ar"` - Arabic

**Locale Codes:**
- `"en-US"` - English (United States)
- `"en-GB"` - English (United Kingdom)
- `"es-ES"` - Spanish (Spain)
- `"es-MX"` - Spanish (Mexico)
- `"fr-FR"` - French (France)
- `"fr-CA"` - French (Canada)
- `"pt-BR"` - Portuguese (Brazil)
- `"zh-CN"` - Chinese (Simplified)
- `"zh-TW"` - Chinese (Traditional)

### Integration with Parser

Both extractors are automatically used during the extraction process:

```go
// In extractAllFields function
go func() {
    defer wg.Done()
    descriptionExtractor := &generic.GenericDescriptionExtractor{}
    if description := descriptionExtractor.Extract(doc.Selection, targetURL, metaCache); description != "" {
        mu.Lock()
        result.Description = description
        mu.Unlock()
    }
}()

go func() {
    defer wg.Done()
    languageExtractor := &generic.GenericLanguageExtractor{}
    if language := languageExtractor.Extract(doc.Selection, targetURL, metaCache); language != "" {
        mu.Lock()
        result.Language = language
        mu.Unlock()
    }
}()
```

## Advanced Features

### Extended Fields

Add custom fields to extraction:

```go
Extend: map[string]*custom.FieldExtractor{
    "category": {
        Selectors: []interface{}{".category a"},
        AllowMultiple: true,
    },
    "tags": {
        Selectors: []interface{}{".tags .tag"},
        AllowMultiple: true,
    },
    "reading_time": {
        Selectors: []interface{}{".reading-time"},
    },
}
```

### Conditional Extraction

Use transforms for conditional logic:

```go
Transforms: map[string]TransformFunction{
    "img": &FunctionTransform{
        Fn: func(node *goquery.Selection) error {
            // Only keep images with alt text
            if alt, exists := node.Attr("alt"); !exists || alt == "" {
                node.Remove()
            }
            return nil
        },
    },
}
```

### Multi-Domain Support

Support multiple domains with one extractor:

```go
extractor := &custom.CustomExtractor{
    Domain: "example.com",
    SupportedDomains: []string{
        "example.com",
        "www.example.com",
        "blog.example.com",
        "news.example.com",
    },
    // ... field definitions
}
```
