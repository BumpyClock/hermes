// ABOUTME: Text processing constants including regex patterns for URL parsing
// ABOUTME: Contains PAGE_IN_HREF_RE pattern for extracting page numbers from URLs

package text

import "regexp"

// PAGE_IN_HREF_RE is a regular expression that looks to try to find the page digit within a URL, if it exists.
// This matches the JavaScript regex: /(page|paging|(p(a|g|ag)?(e|enum|ewanted|ing|ination)))?(=|\/)([0-9]{1,3})/i
//
// Matches:
//   page=1
//   pg=1
//   p=1
//   paging=12
//   pag=7
//   pagination/1
//   paging/88
//   pa/83
//   p/11
//
// Does not match:
//   pg=102
//   page:2
var PAGE_IN_HREF_RE = regexp.MustCompile(`(?i)(page|paging|(p(a|g|ag)?(e|enum|ewanted|ing|ination)))?(=|/)([0-9]{1,3})`)

// HAS_ALPHA_RE matches strings containing alphabetic characters
var HAS_ALPHA_RE = regexp.MustCompile(`(?i)[a-z]`)

// IS_ALPHA_RE matches strings containing only alphabetic characters
var IS_ALPHA_RE = regexp.MustCompile(`(?i)^[a-z]+$`)

// IS_DIGIT_RE matches strings containing only digits
var IS_DIGIT_RE = regexp.MustCompile(`^[0-9]+$`)

// ENCODING_RE matches charset declarations in HTML meta tags  
// Matches both quoted and unquoted charset values like the JavaScript version
var ENCODING_RE = regexp.MustCompile(`(?i)charset=['"]?([\w-]+)['"]?`)

// DEFAULT_ENCODING is the fallback encoding when none is detected
const DEFAULT_ENCODING = "utf-8"