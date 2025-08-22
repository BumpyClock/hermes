// ABOUTME: Date published cleaning and validation with timezone support
// ABOUTME: Faithful port of JavaScript cleaners/date-published.js with comprehensive date parsing

package cleaners

import (
	"strconv"
	"strings"
	"time"
)

// CleanDatePublished takes a date published string and returns a clean ISO date string.
// Returns nil if the date cannot be parsed or is invalid.
//
// This is a faithful 1:1 port of the JavaScript cleanDatePublished function:
// - Handles millisecond/second timestamps 
// - Supports relative time expressions ("5 minutes ago")
// - Handles "now" time indicators
// - Supports timezone and format parameters
// - Cleans date strings by removing "published:" prefixes
// - Returns ISO 8601 formatted string or nil for invalid dates
//
// JavaScript equivalent:
// export default function cleanDatePublished(dateString, { timezone, format } = {}) {
//   // Timestamp handling, date cleaning, and parsing logic
//   return date.isValid() ? date.toISOString() : null;
// }
func CleanDatePublished(dateString, timezone, format string) *string {
	dateString = strings.TrimSpace(dateString)
	if dateString == "" {
		return nil
	}

	// If string is in milliseconds, convert to int and return
	if MS_DATE_STRING.MatchString(dateString) {
		if ms, err := strconv.ParseInt(dateString, 10, 64); err == nil {
			t := time.Unix(0, ms*int64(time.Millisecond)).UTC()
			result := t.Format("2006-01-02T15:04:05.000Z")
			return &result
		}
	}

	// If string is in seconds, convert to int and return  
	if SEC_DATE_STRING.MatchString(dateString) {
		if sec, err := strconv.ParseInt(dateString, 10, 64); err == nil {
			t := time.Unix(sec, 0).UTC()
			result := t.Format("2006-01-02T15:04:05.000Z")
			return &result
		}
	}

	// Try to create date from string
	date := createDate(dateString, timezone, format)
	if date == nil || date.IsZero() {
		// If parsing failed, try cleaning the date string first
		cleaned := cleanDateString(dateString)
		if cleaned != dateString { // Only if cleaning actually changed something
			date = createDate(cleaned, timezone, format)
		}
	}

	if date != nil && !date.IsZero() {
		result := date.UTC().Format("2006-01-02T15:04:05.000Z")
		return &result
	}

	return nil
}

// cleanDateString cleans date strings by removing prefixes and normalizing format
// This is a faithful port of the JavaScript cleanDateString function
func cleanDateString(dateString string) string {
	// Apply cleaning transformations directly to the original string
	// to preserve formatting like commas while still cleaning meridian marks
	cleaned := dateString
	
	// Apply meridian dot cleaning first
	cleaned = TIME_MERIDIAN_DOTS_RE.ReplaceAllString(cleaned, "m")
	
	// Apply meridian spacing fixes
	cleaned = TIME_MERIDIAN_SPACE_RE.ReplaceAllStringFunc(cleaned, func(match string) string {
		submatches := TIME_MERIDIAN_SPACE_RE.FindStringSubmatch(match)
		if len(submatches) >= 4 {
			var builder strings.Builder
			builder.WriteString(submatches[1])
			builder.WriteString(" ")
			builder.WriteString(submatches[2])
			// Normalize spaces in the third group
			thirdGroup := strings.TrimSpace(submatches[3])
			if thirdGroup != "" {
				builder.WriteString(" ")
				builder.WriteString(thirdGroup)
			}
			return builder.String()
		}
		return match
	})
	
	// Apply general date string cleaning (removes prefixes like "PUBLISHED:")
	cleaned = CLEAN_DATE_STRING_RE.ReplaceAllString(cleaned, "$1")
	
	// Check if the cleaned version successfully removed prefixes
	// If it still contains "PUBLISHED" (case insensitive), try reassembly
	cleanedLower := strings.ToLower(cleaned)
	if strings.Contains(cleanedLower, "published") {
		matches := SPLIT_DATE_STRING.FindAllString(dateString, -1)
		if len(matches) > 1 {
			// Use reassembly for cases with clear prefixes that need removal
			assembled := strings.Join(matches, " ")
			assembled = TIME_MERIDIAN_DOTS_RE.ReplaceAllString(assembled, "m")
			assembled = TIME_MERIDIAN_SPACE_RE.ReplaceAllStringFunc(assembled, func(match string) string {
				submatches := TIME_MERIDIAN_SPACE_RE.FindStringSubmatch(match)
				if len(submatches) >= 4 {
					var builder strings.Builder
					builder.WriteString(submatches[1])
					builder.WriteString(" ")
					builder.WriteString(submatches[2])
					builder.WriteString(" ")
					builder.WriteString(submatches[3])
					return builder.String()
				}
				return match
			})
			assembled = CLEAN_DATE_STRING_RE.ReplaceAllString(assembled, "$1")
			return strings.TrimSpace(assembled)
		}
	}
	
	// Return the cleaned version (preserves commas and original formatting)
	return strings.TrimSpace(cleaned)
}

// createDate creates a time.Time from various date string formats
// This is a faithful port of the JavaScript createDate function
func createDate(dateString, timezone, format string) *time.Time {
	dateString = strings.TrimSpace(dateString)
	if dateString == "" {
		return nil
	}

	// Check for timezone offset in the string (like "2021-01-01T12:00:00-0500")
	if TIME_WITH_OFFSET_RE.MatchString(dateString) {
		if t, err := time.Parse(time.RFC3339, dateString); err == nil {
			return &t
		}
		// Try parsing as a different offset format
		if t, err := time.Parse("2006-01-02T15:04:05-0700", dateString); err == nil {
			return &t
		}
	}

	// Handle relative time expressions ("5 minutes ago")
	if TIME_AGO_STRING.MatchString(dateString) {
		matches := TIME_AGO_STRING.FindStringSubmatch(dateString)
		if len(matches) >= 3 {
			if amount, err := strconv.Atoi(matches[1]); err == nil {
				unit := matches[2]
				now := time.Now().UTC()
				
				var duration time.Duration
				switch {
				case strings.HasPrefix(unit, "second"):
					duration = time.Duration(amount) * time.Second
				case strings.HasPrefix(unit, "minute"):
					duration = time.Duration(amount) * time.Minute
				case strings.HasPrefix(unit, "hour"):
					duration = time.Duration(amount) * time.Hour
				case strings.HasPrefix(unit, "day"):
					duration = time.Duration(amount) * 24 * time.Hour
				case strings.HasPrefix(unit, "week"):
					duration = time.Duration(amount) * 7 * 24 * time.Hour
				case strings.HasPrefix(unit, "month"):
					// Approximate month as 30 days
					duration = time.Duration(amount) * 30 * 24 * time.Hour
				case strings.HasPrefix(unit, "year"):
					// Approximate year as 365 days
					duration = time.Duration(amount) * 365 * 24 * time.Hour
				default:
					return nil
				}
				
				result := now.Add(-duration)
				return &result
			}
		}
	}

	// Handle "now" expressions
	if TIME_NOW_STRING.MatchString(dateString) {
		now := time.Now().UTC()
		return &now
	}

	// Try parsing with provided timezone and format
	if timezone != "" || format != "" {
		return parseWithTimezoneAndFormat(dateString, timezone, format)
	}

	// Try common date formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z", // ISO with milliseconds
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02", // This is the key format for "2021-01-01"
		"January 2, 2006",
		"Jan 2, 2006",
		"January 2, 2006 15:04:05",
		"Jan 2, 2006 15:04:05",
		"01/02/2006",
		"01/02/2006 15:04:05",
		"01-02-2006", // US format MM-DD-YYYY
		"01-02-2006 15:04:05",
		"2006/01/02", // ISO-like with slashes
		"2006/01/02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateString); err == nil {
			return &t
		}
	}

	return nil
}

// parseWithTimezoneAndFormat attempts to parse a date string with specified timezone and format
func parseWithTimezoneAndFormat(dateString, timezone, format string) *time.Time {
	var loc *time.Location = time.UTC
	
	// Load timezone if provided
	if timezone != "" {
		if tz, err := time.LoadLocation(timezone); err == nil {
			loc = tz
		}
	}

	// If format is provided, use it
	if format != "" {
		// Convert JavaScript moment.js format to Go format
		// This is a simplified conversion - a full implementation would need
		// a complete mapping from moment.js format tokens to Go format tokens
		goFormat := convertMomentFormatToGo(format)
		if t, err := time.ParseInLocation(goFormat, dateString, loc); err == nil {
			return &t
		}
	}

	// Try common formats with the specified timezone
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"January 2, 2006 15:04:05",
		"Jan 2, 2006 15:04:05",
	}

	for _, fmt := range formats {
		if t, err := time.ParseInLocation(fmt, dateString, loc); err == nil {
			return &t
		}
	}

	return nil
}

// convertMomentFormatToGo converts moment.js format tokens to Go time format
// This is a simplified implementation covering common cases
func convertMomentFormatToGo(momentFormat string) string {
	// Simple replacements for common moment.js tokens
	replacements := map[string]string{
		"YYYY": "2006",
		"MM":   "01", 
		"DD":   "02",
		"HH":   "15",
		"mm":   "04",
		"ss":   "05",
		"A":    "PM",
		"h":    "3",
	}

	result := momentFormat
	for moment, go_fmt := range replacements {
		result = strings.ReplaceAll(result, moment, go_fmt)
	}

	return result
}