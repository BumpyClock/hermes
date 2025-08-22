// ABOUTME: Constants and regex patterns for content cleaning functions
// ABOUTME: Faithful port of JavaScript cleaners/constants.js with all regex patterns preserved

package cleaners

import "regexp"

// CLEAN AUTHOR CONSTANTS

// CLEAN_AUTHOR_RE matches "by" prefixes in author strings
// Matches the JavaScript regex: /^\s*(posted |written )?by\s*:?\s*(.*)/i
// Note: In JavaScript, .* does NOT match newlines by default, so we use [^\r\n]*
var CLEAN_AUTHOR_RE = regexp.MustCompile(`(?i)^\s*(posted |written )?by\s*:?\s*([^\r\n]*)`)

// CLEAN DEK CONSTANTS

// TEXT_LINK_RE matches HTTP/HTTPS URLs in text
// Matches the JavaScript regex: /http(s)?:/i
var TEXT_LINK_RE = regexp.MustCompile(`(?i)http(s)?:`)

// DEK_META_TAGS is an ordered list of meta tag names for article deks
// From most distinct to least distinct
// NOTE: Currently empty as no meta tags provide consistent dek content
var DEK_META_TAGS = []string{}

// DEK_SELECTORS is an ordered list of CSS selectors for article deks
// From most explicit to least explicit
var DEK_SELECTORS = []string{".entry-summary"}

// CLEAN DATE PUBLISHED CONSTANTS

// MS_DATE_STRING matches 13-digit millisecond timestamps
// Matches the JavaScript regex: /^\d{13}$/i
var MS_DATE_STRING = regexp.MustCompile(`(?i)^\d{13}$`)

// SEC_DATE_STRING matches 10-digit second timestamps  
// Matches the JavaScript regex: /^\d{10}$/i
var SEC_DATE_STRING = regexp.MustCompile(`(?i)^\d{10}$`)

// CLEAN_DATE_STRING_RE matches "published:" prefixes in date strings
// Matches the JavaScript regex: /^\s*published\s*:?\s*(.*)/i
var CLEAN_DATE_STRING_RE = regexp.MustCompile(`(?i)^\s*published\s*:?\s*(.*)`)

// TIME_MERIDIAN_SPACE_RE matches time strings with AM/PM
// Matches the JavaScript regex: /(.*\d)(am|pm)(.*)/i
var TIME_MERIDIAN_SPACE_RE = regexp.MustCompile(`(?i)(.*\d)(a|p)(\s*m.*)`)

// TIME_MERIDIAN_DOTS_RE matches ".m." in time strings
// Matches the JavaScript regex: /\.m\./i
var TIME_MERIDIAN_DOTS_RE = regexp.MustCompile(`(?i)\.m\.`)

// TIME_NOW_STRING matches "now" time indicators
// Matches the JavaScript regex: /^\s*(just|right)?\s*now\s*/i
var TIME_NOW_STRING = regexp.MustCompile(`(?i)^\s*(just|right)?\s*now\s*`)

// TIME_AGO_STRING matches relative time expressions like "5 minutes ago"
// Dynamically built from timeUnits like JavaScript version
var TIME_AGO_STRING = regexp.MustCompile(`(?i)(\d+)\s+(seconds?|minutes?|hours?|days?|weeks?|months?|years?)\s+ago`)

// SPLIT_DATE_STRING matches various date/time components
// Complex regex built from multiple timestamp patterns
var SPLIT_DATE_STRING = regexp.MustCompile(`(?i)([0-9]{1,2}:[0-9]{2,2}( ?[ap]\.?m\.?)?)|([0-9]{1,2}[/-][0-9]{1,2}[/-][0-9]{2,4})|(-[0-9]{3,4}$)|([0-9]{1,4})|(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)`)

// TIME_WITH_OFFSET_RE checks if datetime string has timezone offset at end
// Matches the JavaScript regex: /-\d{3,4}$/ but also handles positive offsets
var TIME_WITH_OFFSET_RE = regexp.MustCompile(`[+-]\d{3,4}$`)

// CLEAN TITLE CONSTANTS

// TITLE_SPLITTERS_RE matches title separating characters
// Matches the JavaScript regex: /(: | - | \| )/g
var TITLE_SPLITTERS_RE = regexp.MustCompile(`(: | - | \| )`)

// DOMAIN_ENDINGS_RE matches common domain endings
// Matches the JavaScript regex: /.com$|.net$|.org$|.co.uk$/g
var DOMAIN_ENDINGS_RE = regexp.MustCompile(`\.com$|\.net$|\.org$|\.co\.uk$`)