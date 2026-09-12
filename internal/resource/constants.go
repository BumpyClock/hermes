package resource

import (
	"net/http"
	"regexp"
	"time"
)

// DefaultUserAgent matches current stable Chrome's reduced User-Agent format.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36"

// Request headers that match the JavaScript version.
var REQUEST_HEADERS = map[string]string{
	"User-Agent": DefaultUserAgent,
}

// Standard HTTP headers for web content fetching.
var STANDARD_HEADERS = map[string]string{
	"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	"Accept-Language":           "en-US,en;q=0.5",
	"DNT":                       "1",
	"Connection":                "keep-alive",
	"Upgrade-Insecure-Requests": "1",
}

// The number of milliseconds to attempt to fetch a resource before timing out.
const FETCH_TIMEOUT = 10 * time.Second

// Content types that we do not extract content from.
var BAD_CONTENT_TYPES = []string{
	"audio/mpeg",
	"image/gif",
	"image/jpeg",
	"image/jpg",
}

// Regular expression to match bad content types.
var BAD_CONTENT_TYPES_RE = regexp.MustCompile(`^(` + joinContentTypes() + `)$`)

// Use this setting as the maximum size an article can be
// for us to attempt parsing. Defaults to 5 MB.
const MAX_CONTENT_LENGTH = 5242880

// Maximum document size for processing (10 MB).
const MAX_DOCUMENT_SIZE = 10485760

// Maximum number of DOM elements to process.
const MAX_DOM_ELEMENTS = 50000

// Regular expressions for image and link detection.
var (
	IS_LINK_RE   = regexp.MustCompile(`https?://`)
	IS_IMAGE_RE  = regexp.MustCompile(`\.(png|gif|jpe?g)$`)
	IS_SRCSET_RE = regexp.MustCompile(`\.(png|gif|jpe?g)(\?\S+)?(\s*[\d.]+[wx])`)
)

// Tags to remove during initial DOM cleanup.
const TAGS_TO_REMOVE = "script,style,form"

// joinContentTypes creates a regex-safe string of content types.
func joinContentTypes() string {
	result := ""
	for i, ct := range BAD_CONTENT_TYPES {
		if i > 0 {
			result += "|"
		}
		result += regexp.QuoteMeta(ct)
	}
	return result
}

// MergeHeaders creates a complete header map by merging default and custom headers.
func MergeHeaders(customHeaders map[string]string) map[string]string {
	return mergeHeaders(REQUEST_HEADERS, STANDARD_HEADERS, customHeaders)
}

func mergeHeaders(headerSets ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, headers := range headerSets {
		for name, value := range headers {
			merged[http.CanonicalHeaderKey(name)] = value
		}
	}
	return merged
}
