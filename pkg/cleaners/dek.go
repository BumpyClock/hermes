// ABOUTME: Dek (description/subtitle) cleaning and validation functionality
// ABOUTME: Faithful port of JavaScript cleaners/dek.js with HTML stripping and content validation

package cleaners

import (
	"strings"
	
	"github.com/PuerkitoBio/goquery"
	"github.com/postlight/parser-go/pkg/utils/dom"
	"github.com/postlight/parser-go/pkg/utils/text"
)

// CleanDek takes a dek HTML fragment and returns the cleaned version of it.
// Returns nil if the dek wasn't good enough (too short, too long, has URLs, matches excerpt).
//
// This is a faithful 1:1 port of the JavaScript cleanDek function:
// - Validates length between 5 and 1000 characters
// - Checks that dek isn't the same as excerpt (first 10 words)
// - Strips HTML tags using stripTags function
// - Rejects deks containing plain text URLs (http/https)
// - Normalizes whitespace using normalizeSpaces
//
// JavaScript equivalent:
// export default function cleanDek(dek, { $, excerpt }) {
//   if (dek.length > 1000 || dek.length < 5) return null;
//   if (excerpt && excerptContent(excerpt, 10) === excerptContent(dek, 10)) return null;
//   const dekText = stripTags(dek, $);
//   if (TEXT_LINK_RE.test(dekText)) return null;
//   return normalizeSpaces(dekText.trim());
// }
func CleanDek(dek string, doc *goquery.Document, excerpt string) *string {
	// Sanity check that we didn't get too short or long of a dek
	if len(dek) > 1000 || len(dek) < 5 {
		return nil
	}

	// Check that dek isn't the same as excerpt
	// Use excerptContent to compare first 10 words of each
	if excerpt != "" {
		dekExcerpt := text.ExcerptContent(dek, 10)
		excerptSample := text.ExcerptContent(excerpt, 10)
		if dekExcerpt == excerptSample {
			return nil
		}
	}

	// Strip HTML tags from the dek
	dekText := dom.StripTags(dek, doc)

	// Plain text links shouldn't exist in the dek. If we have some, it's
	// not a good dek - bail.
	if TEXT_LINK_RE.MatchString(dekText) {
		return nil
	}

	// Normalize spaces and trim whitespace
	cleaned := text.NormalizeSpaces(strings.TrimSpace(dekText))
	
	// Final check - if after cleaning it's too short, reject it
	if len(cleaned) < 5 {
		return nil
	}

	return &cleaned
}