# Results API reference

`Result` contains article content, metadata, and methods that check field availability.

## Contents

- [Structure](#structure)
- [Field details](#field-details)
- [Methods](#methods)
- [JSON](#json)

## Structure

```go
type Result struct {
    // Core content fields
    URL           string     `json:"url"`
    Title         string     `json:"title"`
    Content       string     `json:"content"`
    Author        string     `json:"author,omitempty"`
    DatePublished *time.Time `json:"date_published,omitempty"`

    // Media and metadata
    LeadImageURL string `json:"lead_image_url,omitempty"`
    Dek          string `json:"dek,omitempty"`
    Domain       string `json:"domain"`
    Excerpt      string `json:"excerpt,omitempty"`

    // Content metrics
    WordCount     int    `json:"word_count"`
    Direction     string `json:"direction,omitempty"`
    TotalPages    int    `json:"total_pages,omitempty"`
    RenderedPages int    `json:"rendered_pages,omitempty"`

    // Site information
    SiteName    string `json:"site_name,omitempty"`
    SiteTitle   string `json:"site_title,omitempty"`
    SiteImage   string `json:"site_image,omitempty"`
    Description string `json:"description,omitempty"`
    Language    string `json:"language,omitempty"`
    ThemeColor  string `json:"theme_color,omitempty"`
    Favicon     string `json:"favicon,omitempty"`

    // Video metadata
    VideoURL      string                 `json:"video_url,omitempty"`
    VideoMetadata map[string]interface{} `json:"video_metadata,omitempty"`
}
```

## Field details

| Field | Meaning |
| --- | --- |
| `URL` | Article URL. |
| `Title` | Article headline. |
| `Content` | Extracted content body. `WithContentType` selects the format. |
| `Author` | Author names. |
| `DatePublished` | Publication date, or `nil` when unavailable. |
| `LeadImageURL` | Primary article image URL. |
| `Dek` | Article subtitle. |
| `Domain` | Domain of the URL. |
| `Excerpt` | Short excerpt or summary. |
| `WordCount` | Content word count. |
| `Direction` | Text direction, `"ltr"` or `"rtl"`. |
| `TotalPages`, `RenderedPages` | Pagination metadata. |
| `SiteName`, `Description`, `Language`, `ThemeColor`, `Favicon` | Site metadata. |
| `SiteTitle`, `SiteImage` | Fields present in `Result` but not populated by the public client's result conversion. |
| `VideoURL`, `VideoMetadata` | Video information, when available. |

## Methods

```go
func (r *Result) FormatMarkdown() string
func (r *Result) IsEmpty() bool
func (r *Result) HasAuthor() bool
func (r *Result) HasDate() bool
func (r *Result) HasImage() bool
```

`FormatMarkdown` returns a document with the title, available metadata, description or excerpt, and content. It does not convert `Content` to Markdown.

Use `WithContentType("markdown")` before you parse to obtain a Markdown content body.

`IsEmpty` returns `true` when both `Title` and `Content` are empty. `HasAuthor`, `HasDate`, and `HasImage` check `Author`, `DatePublished`, and `LeadImageURL`, respectively.

## JSON

Use `encoding/json` to serialize a result:

```go
data, err := json.MarshalIndent(res, "", "  ")
if err != nil { /* handle */ }
fmt.Println(string(data))
```

The default JSON encoder represents a non-nil `DatePublished` as an RFC 3339 string, with fractional seconds when necessary.
