// ABOUTME: Constants and regex patterns for content cleaning functions
// ABOUTME: Faithful port of JavaScript cleaners/constants.js with all regex patterns preserved

package cleaners

import "regexp"

// CLEAN AUTHOR CONSTANTS

// CLEAN_AUTHOR_RE matches "by" prefixes in author strings
// Matches the JavaScript regex: /^\s*(posted |written )?by\s*:?\s*(.*)/i
// Note: In JavaScript, .* does NOT match newlines by default, so we use [^\r\n]*.
var CLEAN_AUTHOR_RE = regexp.MustCompile(`(?i)^\s*(posted |written )?by\s*:?\s*([^\r\n]*)`)

// CLEAN TITLE CONSTANTS

// TITLE_SPLITTERS_RE matches title separating characters
// Matches the JavaScript regex: /(: | - | \| )/g.
var TITLE_SPLITTERS_RE = regexp.MustCompile(`(: | - | \| )`)

// DOMAIN_ENDINGS_RE matches common domain endings
// Matches the JavaScript regex: /.com$|.net$|.org$|.co.uk$/g.
var DOMAIN_ENDINGS_RE = regexp.MustCompile(`\.com$|\.net$|\.org$|\.co\.uk$`)

// CLEAN CONTENT CONSTANTS

// EMPTY_HTML_RE matches HTML that only contains whitespace and/or br tags
// Used to detect and remove empty paragraph elements during content cleaning
// Supports: Unicode whitespace (\p{Zs}), NBSP, case-insensitive br tags with attributes.
var EMPTY_HTML_RE = regexp.MustCompile(`(?i)^([\s\p{Zs}\x{00A0}]|<br[^>]*\/?>)*$`)
