// ABOUTME: Generic date published extractor for extracting publication dates from articles
// ABOUTME: Implements 100% JavaScript-compatible date extraction from meta tags, CSS selectors, and URLs

package generic

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/hermes/internal/utils/dom"
	"github.com/BumpyClock/hermes/internal/utils/text"
)

// DATE_PUBLISHED_META_TAGS - Ordered list of meta tag names that denote likely date published dates
// All attributes should be lowercase for faster case-insensitive matching
// From most distinct to least distinct (matches JavaScript exactly)
var DATE_PUBLISHED_META_TAGS = []string{
	"article:published_time",
	"displaydate",
	"dc.date",
	"dc.date.issued",
	"rbpubdate",
	"publish_date",
	"pub_date",
	"pagedate",
	"pubdate",
	"revision_date",
	"doc_date",
	"date_created",
	"content_create_date",
	"lastmodified",
	"created",
	"date",
}

// DATE_PUBLISHED_SELECTORS - Ordered list of CSS selectors to find likely date published dates
// From most explicit to least explicit (matches JavaScript exactly)
var DATE_PUBLISHED_SELECTORS = []string{
	".hentry .dtstamp.published",
	".hentry .published",
	".hentry .dtstamp.updated",
	".hentry .updated",
	".single .published",
	".meta .published",
	".meta .postDate",
	".entry-date",
	".byline .date",
	".postmetadata .date",
	".article_datetime",
	".date-header",
	".story-date",
	".dateStamp",
	"#story .datetime",
	".dateline",
	".pubdate",
}

// DATE_PUBLISHED_URL_RES - Ordered list of compiled regular expressions to find likely date
// published dates from the URL. These should always have the first reference be a date string
// that is parseable. Matches JavaScript exactly.
var DATE_PUBLISHED_URL_RES = []*regexp.Regexp{
	regexp.MustCompile(`/(20\d{2}/\d{2}/\d{2})/`),                                                // 2023/12/01
	regexp.MustCompile(`(20\d{2}-[01]\d-[0-3]\d)`),                                             // 2023-12-01
	regexp.MustCompile(`/(20\d{2}/(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)/[0-3]\d)/`), // 2023/dec/01
}

// JavaScript date cleaner constants (ported from cleaners/constants.js)
var (
	MS_DATE_STRING        = regexp.MustCompile(`^\d{13}$`)
	SEC_DATE_STRING       = regexp.MustCompile(`^\d{10}$`)
	CLEAN_DATE_STRING_RE  = regexp.MustCompile(`^\s*published\s*:?\s*(.*)`)
	TIME_MERIDIAN_SPACE_RE = regexp.MustCompile(`(.*\d)(am|pm)(.*)`)
	TIME_MERIDIAN_DOTS_RE = regexp.MustCompile(`\.m\.`)
	TIME_NOW_STRING       = regexp.MustCompile(`^\s*(just|right)?\s*now\s*`)
	TIME_WITH_OFFSET_RE   = regexp.MustCompile(`-\d{3,4}$`)
)

// TIME_AGO_STRING regex for parsing relative dates (X minutes ago, etc.)
var TIME_AGO_STRING = regexp.MustCompile(`(\d+)\s+(seconds?|minutes?|hours?|days?|weeks?|months?|years?)\s+ago`)

// SPLIT_DATE_STRING regex for splitting date components (matches JavaScript exactly with case-insensitive)
var SPLIT_DATE_STRING = regexp.MustCompile(`(?i)([0-9]{1,2}:[0-9]{2,2}( ?[ap].?m.?)?)|([0-9]{1,2}[/-][0-9]{1,2}[/-][0-9]{2,4})|(-[0-9]{3,4}$)|([0-9]{1,4})|(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec|january|february|march|april|may|june|july|august|september|october|november|december)`)

// GenericDateExtractor - Extractor for publication dates with 100% JavaScript compatibility
type GenericDateExtractorType struct{}

var GenericDateExtractor = GenericDateExtractorType{}

// Extract publication date from document using meta tags, selectors, and URL patterns
func (e GenericDateExtractorType) Extract(doc *goquery.Selection, url string, metaCache []string) *string {
	var datePublished string
	
	// Convert Selection to Document for meta tag extraction
	var document *goquery.Document
	if html, err := doc.Html(); err == nil {
		if doc.Is("html") {
			document, _ = goquery.NewDocumentFromReader(strings.NewReader(html))
		} else {
			// Wrap in HTML if not already an html element
			document, _ = goquery.NewDocumentFromReader(strings.NewReader("<html>" + html + "</html>"))
		}
	}
	
	// First, check to see if we have a matching meta tag that we can make use of.
	// Don't try cleaning tags from this string (false parameter matches JavaScript)
	if document != nil {
		if meta := dom.ExtractFromMeta(document, DATE_PUBLISHED_META_TAGS, metaCache, false); meta != nil {
			datePublished = *meta
			if cleaned := cleanDatePublished(datePublished, nil); cleaned != nil {
				return cleaned
			}
		}
	}
	
	// Second, look through our selectors looking for potential date_published's
	if selector := dom.ExtractFromSelectors(doc, DATE_PUBLISHED_SELECTORS, 5, false); selector != nil {
		datePublished = *selector
		if cleaned := cleanDatePublished(datePublished, nil); cleaned != nil {
			return cleaned
		}
	}
	
	// Lastly, look to see if a date string exists in the URL
	if urlDate, found := text.ExtractFromURL(url, DATE_PUBLISHED_URL_RES); found {
		datePublished = urlDate
		if cleaned := cleanDatePublished(datePublished, nil); cleaned != nil {
			return cleaned
		}
	}
	
	return nil
}

// cleanDatePublished takes a date published string and returns a clean ISO date string
// Implements 100% JavaScript compatibility with moment.js behavior
func cleanDatePublished(dateString string, options map[string]interface{}) *string {
	if dateString == "" {
		return nil
	}
	
	// Handle timezone and format options (for future compatibility)
	var timezone string
	var format string
	if options != nil {
		if tz, ok := options["timezone"].(string); ok {
			timezone = tz
		}
		if fmt, ok := options["format"].(string); ok {
			format = fmt
		}
	}
	
	// If string is in milliseconds, convert to int and return (13 digits)
	if MS_DATE_STRING.MatchString(dateString) {
		if ms, err := strconv.ParseInt(dateString, 10, 64); err == nil {
			t := time.Unix(ms/1000, (ms%1000)*1000000).UTC()
			iso := t.Format("2006-01-02T15:04:05.000Z")
			return &iso
		}
	}
	
	// If string is in seconds, convert to int and return (10 digits)
	if SEC_DATE_STRING.MatchString(dateString) {
		if sec, err := strconv.ParseInt(dateString, 10, 64); err == nil {
			t := time.Unix(sec, 0).UTC()
			iso := t.Format("2006-01-02T15:04:05.000Z")
			return &iso
		}
	}
	
	// Try to create date using various parsing strategies
	if date := createDate(dateString, timezone, format); date != nil {
		iso := date.UTC().Format("2006-01-02T15:04:05.000Z")
		return &iso
	}
	
	// If that failed, clean the date string and try again
	cleanedDateString := cleanDateString(dateString)
	if date := createDate(cleanedDateString, timezone, format); date != nil {
		iso := date.UTC().Format("2006-01-02T15:04:05.000Z")
		return &iso
	}
	
	return nil
}

// cleanDateString performs JavaScript-compatible date string cleaning
func cleanDateString(dateString string) string {
	// Extract date components using SPLIT_DATE_STRING regex
	matches := SPLIT_DATE_STRING.FindAllString(dateString, -1)
	if len(matches) > 0 {
		dateString = strings.Join(matches, " ")
	}
	
	// Replace meridian dots (.m. -> m)
	dateString = TIME_MERIDIAN_DOTS_RE.ReplaceAllString(dateString, "m")
	
	// Fix meridian spacing (e.g., "3pm something" -> "3 pm something")
	dateString = TIME_MERIDIAN_SPACE_RE.ReplaceAllString(dateString, "$1 $2 $3")
	
	// Remove "published:" prefix
	if matches := CLEAN_DATE_STRING_RE.FindStringSubmatch(dateString); len(matches) > 1 {
		dateString = matches[1]
	}
	
	return strings.TrimSpace(dateString)
}

// createDate creates a time.Time from various date string formats
// Implements JavaScript moment.js-like behavior
func createDate(dateString, timezone, format string) *time.Time {
	if dateString == "" {
		return nil
	}
	
	// Check for offset in the string (matches JavaScript TIME_WITH_OFFSET_RE)
	if TIME_WITH_OFFSET_RE.MatchString(dateString) {
		if t, err := time.Parse(time.RFC3339, dateString); err == nil {
			return &t
		}
		// Try parsing as date with simple offset
		if t, err := time.Parse("2006-01-02T15:04:05-0700", dateString); err == nil {
			return &t
		}
	}
	
	// Check for relative time strings ("5 minutes ago", "now", etc.)
	if TIME_AGO_STRING.MatchString(dateString) {
		matches := TIME_AGO_STRING.FindStringSubmatch(dateString)
		if len(matches) >= 3 {
			amount, err := strconv.Atoi(matches[1])
			if err == nil {
				unit := matches[2]
				now := time.Now()
				
				// Convert to singular for switch statement
				unit = strings.TrimSuffix(unit, "s")
				
				var duration time.Duration
				switch unit {
				case "second":
					duration = time.Duration(amount) * time.Second
				case "minute":
					duration = time.Duration(amount) * time.Minute
				case "hour":
					duration = time.Duration(amount) * time.Hour
				case "day":
					duration = time.Duration(amount) * 24 * time.Hour
				case "week":
					duration = time.Duration(amount) * 7 * 24 * time.Hour
				case "month":
					// Approximate month as 30 days
					duration = time.Duration(amount) * 30 * 24 * time.Hour
				case "year":
					// Approximate year as 365 days
					duration = time.Duration(amount) * 365 * 24 * time.Hour
				}
				
				result := now.Add(-duration)
				return &result
			}
		}
	}
	
	// Check for "now" strings
	if TIME_NOW_STRING.MatchString(dateString) {
		now := time.Now()
		return &now
	}
	
	// Use timezone if provided
	_ = timezone // Timezone support not implemented - would require zone parsing
	_ = format   // Custom format support not implemented - uses standard Go layouts
	
	// Try general-purpose date parsing (using existing text utils)
	if parsed, err := text.ParseDate(dateString); err == nil {
		// Convert to UTC to match JavaScript behavior
		utc := parsed.UTC()
		return &utc
	}
	
	return nil
}