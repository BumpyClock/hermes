package generic

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GenericThemeColorExtractor extracts the theme color from meta tags
type GenericThemeColorExtractor struct{}

// Extract extracts the theme color from the page
func (extractor *GenericThemeColorExtractor) Extract(selection *goquery.Selection, pageURL string, metaCache []string) string {
	// Look for meta name="theme-color" tag
	themeColor := selection.Find("meta[name=\"theme-color\"]").AttrOr("content", "")
	if themeColor != "" {
		return extractor.validateAndNormalizeColor(themeColor)
	}

	// Fallback: Look for meta name="msapplication-TileColor" (Microsoft specific)
	tileColor := selection.Find("meta[name=\"msapplication-TileColor\"]").AttrOr("content", "")
	if tileColor != "" {
		return extractor.validateAndNormalizeColor(tileColor)
	}

	// No theme color found
	return ""
}

// validateAndNormalizeColor validates and normalizes color values
func (extractor *GenericThemeColorExtractor) validateAndNormalizeColor(color string) string {
	color = strings.TrimSpace(color)
	if color == "" {
		return ""
	}

	// Convert to lowercase for consistency
	color = strings.ToLower(color)

	// Validate hex colors (#fff, #ffffff)
	hexPattern := regexp.MustCompile(`^#([a-f0-9]{3}|[a-f0-9]{6})$`)
	if hexPattern.MatchString(color) {
		return color
	}

	// Validate rgb/rgba colors
	rgbPattern := regexp.MustCompile(`^rgba?\(\s*\d+\s*,\s*\d+\s*,\s*\d+\s*(,\s*[0-9.]+)?\s*\)$`)
	if rgbPattern.MatchString(color) {
		return color
	}

	// Validate hsl/hsla colors
	hslPattern := regexp.MustCompile(`^hsla?\(\s*\d+\s*,\s*\d+%\s*,\s*\d+%\s*(,\s*[0-9.]+)?\s*\)$`)
	if hslPattern.MatchString(color) {
		return color
	}

	// Validate named colors (basic validation for common colors)
	namedColors := map[string]bool{
		"black": true, "white": true, "red": true, "green": true, "blue": true,
		"yellow": true, "cyan": true, "magenta": true, "gray": true, "grey": true,
		"orange": true, "purple": true, "pink": true, "brown": true, "lime": true,
		"navy": true, "teal": true, "olive": true, "maroon": true, "silver": true,
		"gold": true, "transparent": true,
	}

	if namedColors[color] {
		return color
	}

	// Invalid color format, return empty string
	return ""
}