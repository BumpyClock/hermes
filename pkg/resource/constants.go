package resource

import (
	"regexp"
	"time"
)

// Request headers that match the JavaScript version
var REQUEST_HEADERS = map[string]string{
	"User-Agent": "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36",
}

// The number of milliseconds to attempt to fetch a resource before timing out
const FETCH_TIMEOUT = 10 * time.Second

// Content types that we do not extract content from
var BAD_CONTENT_TYPES = []string{
	"audio/mpeg",
	"image/gif",
	"image/jpeg", 
	"image/jpg",
}

// Regular expression to match bad content types
var BAD_CONTENT_TYPES_RE = regexp.MustCompile(`^(` + joinContentTypes() + `)$`)

// Use this setting as the maximum size an article can be
// for us to attempt parsing. Defaults to 5 MB.
const MAX_CONTENT_LENGTH = 5242880

// Regular expressions for image and link detection
var (
	IS_LINK_RE   = regexp.MustCompile(`https?://`)
	IS_IMAGE_RE  = regexp.MustCompile(`\.(png|gif|jpe?g)$`)
	IS_SRCSET_RE = regexp.MustCompile(`\.(png|gif|jpe?g)(\?\S+)?(\s*[\d.]+[wx])`)
)

// Tags to remove during initial DOM cleanup
const TAGS_TO_REMOVE = "script,style,form"

// Default encoding constants
const DEFAULT_ENCODING = "utf-8"

var ENCODING_RE = regexp.MustCompile(`charset=([\w-]+)\b`)

// joinContentTypes creates a regex-safe string of content types
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