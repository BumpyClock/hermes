// ABOUTME: Lead image extraction with scoring and selection strategies for article images
// ABOUTME: Ports JavaScript GenericLeadImageUrlExtractor with meta tag extraction, content image scoring, and fallback selectors

package generic

import (
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Lead image URL meta tags in priority order (most distinct first)
var LEAD_IMAGE_URL_META_TAGS = []string{
	"og:image",
	"twitter:image",
	"image_src",
}

// Fallback selectors for lead image extraction
var LEAD_IMAGE_URL_SELECTORS = []string{
	"link[rel=image_src]",
}

// Positive hints that increase image score
var POSITIVE_LEAD_IMAGE_URL_HINTS = []string{
	"upload",
	"wp-content",
	"large",
	"photo",
	"wp-image",
}

// Negative hints that decrease image score
var NEGATIVE_LEAD_IMAGE_URL_HINTS = []string{
	"spacer", "sprite", "blank", "throbber", "gradient", "tile", "bg",
	"background", "icon", "social", "header", "hdr", "advert", "spinner",
	"loader", "loading", "default", "rating", "share", "facebook",
	"twitter", "theme", "promo", "ads", "wp-includes",
}

// Compiled regexes for URL scoring
var (
	POSITIVE_LEAD_IMAGE_URL_HINTS_RE = regexp.MustCompile("(?i)" + strings.Join(POSITIVE_LEAD_IMAGE_URL_HINTS, "|"))
	NEGATIVE_LEAD_IMAGE_URL_HINTS_RE = regexp.MustCompile("(?i)" + strings.Join(NEGATIVE_LEAD_IMAGE_URL_HINTS, "|"))
	GIF_RE                           = regexp.MustCompile(`(?i)\.gif(\?.*)?$`)
	JPG_RE                           = regexp.MustCompile(`(?i)\.jpe?g(\?.*)?$`)
	PHOTO_HINTS_RE                   = regexp.MustCompile(`(?i)figure|photo|image|caption`) // From constants.go
)

// ExtractorImageParams contains parameters for image extraction
type ExtractorImageParams struct {
	Doc       *goquery.Document
	Content   string
	MetaCache map[string]string
	HTML      string
}

// GenericLeadImageExtractor implements lead image extraction logic
type GenericLeadImageExtractor struct{}

// NewGenericLeadImageExtractor creates a new lead image extractor
func NewGenericLeadImageExtractor() *GenericLeadImageExtractor {
	return &GenericLeadImageExtractor{}
}

// Extract finds the lead image URL from the document using scoring and fallback strategies
// Matches JavaScript behavior: meta tags → content images → fallback selectors
func (e *GenericLeadImageExtractor) Extract(params ExtractorImageParams) *string {
	doc := params.Doc
	
	// JavaScript: if (!$.browser && $('head').length === 0) - handle headless HTML
	if doc.Find("head").Length() == 0 {
		// Prepend HTML to first element to ensure proper parsing
		doc.Find("*").First().PrependHtml(params.HTML)
	}

	// Check meta tags first (moving higher because of Open Graph/Twitter cards)
	if imageUrl := e.extractFromMetaTags(doc, params.MetaCache); imageUrl != nil {
		if cleanUrl := cleanImage(*imageUrl); cleanUrl != nil {
			return cleanUrl
		}
	}

	// Try to find the "best" image via content scoring
	if params.Content != "" {
		if imageUrl := e.extractFromContent(doc, params.Content); imageUrl != nil {
			if cleanUrl := cleanImage(*imageUrl); cleanUrl != nil {
				return cleanUrl
			}
		}
	}

	// Fallback to selector-based extraction
	if imageUrl := e.extractFromSelectors(doc); imageUrl != nil {
		if cleanUrl := cleanImage(*imageUrl); cleanUrl != nil {
			return cleanUrl
		}
	}

	return nil
}

// extractFromMetaTags extracts image URL from meta tags using priority order
// Handles both standard meta[name] and OpenGraph meta[property] tags
func (e *GenericLeadImageExtractor) extractFromMetaTags(doc *goquery.Document, metaCache map[string]string) *string {
	for _, metaName := range LEAD_IMAGE_URL_META_TAGS {
		// Try both name and property attributes for maximum compatibility
		selectors := []string{
			fmt.Sprintf("meta[name=\"%s\"]", metaName),
			fmt.Sprintf("meta[property=\"%s\"]", metaName),
		}
		
		for _, selector := range selectors {
			nodes := doc.Find(selector)
			if nodes.Length() == 0 {
				continue
			}
			
			// Check both content and value attributes
			var imageUrl string
			nodes.Each(func(i int, node *goquery.Selection) {
				if imageUrl != "" {
					return // Already found
				}
				
				// Try content attribute first (standard for OpenGraph)
				if content, exists := node.Attr("content"); exists && content != "" {
					imageUrl = content
					return
				}
				
				// Try value attribute (original JavaScript behavior)
				if value, exists := node.Attr("value"); exists && value != "" {
					imageUrl = value
					return
				}
			})
			
			if imageUrl != "" {
				return &imageUrl
			}
		}
	}
	
	return nil
}

// extractFromContent scores images in content and returns the highest scoring one
func (e *GenericLeadImageExtractor) extractFromContent(doc *goquery.Document, content string) *string {
	contentSelection := doc.Find(content)
	if contentSelection.Length() == 0 {
		// If content selector doesn't match, use the whole document
		contentSelection = doc.Selection
	}

	imgs := contentSelection.Find("img")
	if imgs.Length() == 0 {
		return nil
	}

	imgScores := make(map[string]int)
	imgArray := make([]interface{}, imgs.Length())

	imgs.Each(func(index int, img *goquery.Selection) {
		src, exists := img.Attr("src")
		if !exists || src == "" {
			return
		}

		score := 0
		score += scoreImageUrl(src)
		score += scoreAttr(img)
		score += scoreByParents(img)
		score += scoreBySibling(img)
		score += scoreByDimensions(img)
		score += int(scoreByPosition(imgArray, index))

		imgScores[src] = score
	})

	// Find the highest scoring image
	var topUrl string
	topScore := 0
	
	for url, score := range imgScores {
		if score > topScore {
			topUrl = url
			topScore = score
		}
	}

	if topScore > 0 {
		return &topUrl
	}

	return nil
}

// extractFromSelectors tries fallback selectors for image URLs
func (e *GenericLeadImageExtractor) extractFromSelectors(doc *goquery.Document) *string {
	for _, selector := range LEAD_IMAGE_URL_SELECTORS {
		node := doc.Find(selector).First()
		if node.Length() == 0 {
			continue
		}

		// Try src attribute
		if src, exists := node.Attr("src"); exists && src != "" {
			return &src
		}

		// Try href attribute
		if href, exists := node.Attr("href"); exists && href != "" {
			return &href
		}

		// Try value attribute
		if value, exists := node.Attr("value"); exists && value != "" {
			return &value
		}
	}

	return nil
}

// scoreImageUrl scores URLs based on hints and file extensions
func scoreImageUrl(url string) int {
	url = strings.TrimSpace(url)
	score := 0

	if POSITIVE_LEAD_IMAGE_URL_HINTS_RE.MatchString(url) {
		score += 20
	}

	if NEGATIVE_LEAD_IMAGE_URL_HINTS_RE.MatchString(url) {
		score -= 20
	}

	// GIFs are less desirable (but still common/popular)
	if GIF_RE.MatchString(url) {
		score -= 10
	}

	if JPG_RE.MatchString(url) {
		score += 10
	}

	// PNGs are neutral (no score change)
	return score
}

// scoreAttr gives bonus for alt attribute (non-presentational)
func scoreAttr(img *goquery.Selection) int {
	if _, exists := img.Attr("alt"); exists {
		return 5
	}
	return 0
}

// scoreByParents looks for figure-like containers and photo hints in parents
func scoreByParents(img *goquery.Selection) int {
	score := 0

	// Check for figure parent
	figParent := img.Parents().Filter("figure").First()
	if figParent.Length() == 1 {
		score += 25
	}

	// Check parent and grandparent for photo hints
	parent := img.Parent()
	var gParent *goquery.Selection
	if parent.Length() == 1 {
		gParent = parent.Parent()
	}

	// Check both parent and grandparent for photo hints
	if parent.Length() > 0 {
		sig := getSig(parent)
		if PHOTO_HINTS_RE.MatchString(sig) {
			score += 15
		}
	}

	if gParent != nil && gParent.Length() > 0 {
		sig := getSig(gParent)
		if PHOTO_HINTS_RE.MatchString(sig) {
			score += 15
		}
	}

	return score
}

// scoreBySibling checks for caption-like siblings
func scoreBySibling(img *goquery.Selection) int {
	score := 0
	sibling := img.Next()

	// Check for figcaption sibling
	if sibling.Length() > 0 {
		if sibling.Is("figcaption") {
			score += 25
		}

		sig := getSig(sibling)
		if PHOTO_HINTS_RE.MatchString(sig) {
			score += 15
		}
	}

	return score
}

// scoreByDimensions scores based on image dimensions
func scoreByDimensions(img *goquery.Selection) int {
	score := 0
	src, _ := img.Attr("src")

	widthStr, widthExists := img.Attr("width")
	heightStr, heightExists := img.Attr("height")

	if !widthExists || !heightExists {
		return 0
	}

	width, err1 := strconv.ParseFloat(widthStr, 64)
	height, err2 := strconv.ParseFloat(heightStr, 64)

	if err1 != nil || err2 != nil {
		return 0
	}

	// Penalty for skinny images
	if width <= 50 {
		score -= 50
	}

	// Penalty for short images
	if height <= 50 {
		score -= 50
	}

	// Area-based scoring (but not for sprites)
	if width > 0 && height > 0 && !strings.Contains(src, "sprite") {
		area := width * height
		if area < 5000 {
			// Smaller than 50 x 100
			score -= 100
		} else {
			score += int(math.Round(area / 1000))
		}
	}

	return score
}

// scoreByPosition gives bonus to images earlier in the content
func scoreByPosition(imgs []interface{}, index int) float64 {
	return float64(len(imgs))/2.0 - float64(index)
}

// getSig gets the signature (class + id) of an element for scoring
func getSig(node *goquery.Selection) string {
	class, _ := node.Attr("class")
	id, _ := node.Attr("id")
	return fmt.Sprintf("%s %s", class, id)
}

// cleanImage validates and cleans image URLs
func cleanImage(imageUrl string) *string {
	imageUrl = strings.TrimSpace(imageUrl)
	if imageUrl == "" {
		return nil
	}

	// Parse URL to validate
	_, err := url.Parse(imageUrl)
	if err != nil {
		return nil
	}

	// Basic validation - must be http/https
	if !strings.HasPrefix(imageUrl, "http://") && !strings.HasPrefix(imageUrl, "https://") {
		return nil
	}

	return &imageUrl
}