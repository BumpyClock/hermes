package text

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/markusmobius/go-dateparser"
)

var (
	htmlTagDateRE      = regexp.MustCompile(`<[^>]*>`)
	spaceDateRE        = regexp.MustCompile(`\s+`)
	nonPrintableDateRE = regexp.MustCompile(`[^\x20-\x7E]`)
)

// ParseDate attempts to parse a date string using various methods.
func ParseDate(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, fmt.Errorf("empty date string")
	}

	// Clean the date string
	dateStr = cleanDateString(dateStr)
	if dateStr == "" {
		return nil, fmt.Errorf("date string became empty after cleaning")
	}

	// Try go-dateparser first (most flexible)
	cfg := &dateparser.Configuration{
		CurrentTime:   time.Now(),
		StrictParsing: false,
	}

	if parsedTime, err := dateparser.Parse(cfg, dateStr); err == nil {
		return &parsedTime.Time, nil
	}

	// Try standard time.Parse with common formats (fallback)
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"01-02-2006",
		"1-2-2006",
		"January 2, 2006",
		"Jan 2, 2006",
		"2 January 2006",
		"2 Jan 2006",
		"02 Jan 2006",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05+07:00",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("unable to parse date: %s", dateStr)
}

// cleanDateString performs basic cleaning on date strings.
func cleanDateString(dateStr string) string {
	// Trim whitespace
	dateStr = strings.TrimSpace(dateStr)

	// Remove common prefixes
	prefixes := []string{"Published:", "Updated:", "Date:", "Posted:", "By "}
	for _, prefix := range prefixes {
		if strings.HasPrefix(dateStr, prefix) {
			dateStr = strings.TrimSpace(dateStr[len(prefix):])
		}
	}

	// Remove HTML tags if any
	dateStr = htmlTagDateRE.ReplaceAllString(dateStr, "")

	// Remove extra whitespace
	dateStr = spaceDateRE.ReplaceAllString(dateStr, " ")

	// Remove non-printable characters
	dateStr = nonPrintableDateRE.ReplaceAllString(dateStr, "")

	return strings.TrimSpace(dateStr)
}
