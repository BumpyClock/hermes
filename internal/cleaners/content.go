// ABOUTME: Content cleaning pipeline that transforms extracted article content into clean, readable output
// ABOUTME: Direct port of JavaScript extractCleanNode function with 100% compatibility for all cleaning stages

package cleaners

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"

	"github.com/BumpyClock/hermes/internal/utils/dom"
)

// ContentCleanOptions represents configuration options for content cleaning.
type ContentCleanOptions struct {
	CleanConditionally bool
	Title              string
	URL                string
	DefaultCleaner     *bool // Use pointer to distinguish between unset and explicitly false
	Preserve           []string
	PreserveMatchers   []goquery.Matcher
}

var (
	contentImages       = cascadia.MustCompile("img")
	contentImageSources = cascadia.MustCompile("img[src]")
	contentLinks        = cascadia.MustCompile("a[href], link[href]")
	contentAnchors      = cascadia.MustCompile("a")
	contentKeep         = cascadia.MustCompile(".hermes-parser-keep")
	contentJunk         = cascadia.MustCompile("script, style, link, meta, noscript, template, title, head, object, embed, applet")
	contentHOnes        = cascadia.MustCompile("h1")
	contentHeaders      = cascadia.MustCompile("h1, h2, h3, h4, h5, h6")
	contentParagraphs   = cascadia.MustCompile("p")
	contentMedia        = cascadia.MustCompile("iframe, video, audio, embed, object")
	contentEmpty        = cascadia.MustCompile("p, div, span")
	contentAll          = cascadia.MustCompile("*")
	contentKeepMedia    = []goquery.Matcher{
		cascadia.MustCompile("iframe[src*='youtube.com']"),
		cascadia.MustCompile("iframe[src*='www.youtube.com']"),
		cascadia.MustCompile("iframe[src*='youtu.be']"),
		cascadia.MustCompile("iframe[src*='vimeo.com']"),
		cascadia.MustCompile("iframe[src*='player.vimeo.com']"),
		cascadia.MustCompile("object[data*='youtube.com']"),
		cascadia.MustCompile("object[data*='vimeo.com']"),
		cascadia.MustCompile("embed[src*='youtube.com']"),
		cascadia.MustCompile("embed[src*='vimeo.com']"),
	}
	contentConditional = []goquery.Matcher{
		cascadia.MustCompile("div"),
		cascadia.MustCompile("section"),
		cascadia.MustCompile("header"),
		cascadia.MustCompile("footer"),
		cascadia.MustCompile("aside"),
		cascadia.MustCompile("nav"),
	}
)

// ExtractCleanNode cleans article content, returning a new, cleaned node
// Direct port of JavaScript extractCleanNode function with identical cleaning pipeline:
//
// 1. rewriteTopLevel - Convert HTML/BODY tags to DIV to avoid complications
// 2. cleanImages - Remove small/spacer images (if defaultCleaner enabled)
// 3. makeLinksAbsolute - Convert relative URLs to absolute URLs
// 4. markToKeep - Mark video iframes and important elements for preservation
// 5. stripJunkTags - Remove script, style, title and other junk tags
// 6. cleanHOnes - Remove or convert H1 tags based on count
// 7. cleanHeaders - Clean headers that match article title
// 8. cleanTags - Remove low-quality tags with high link density (if defaultCleaner enabled)
// 9. removeEmpty - Remove empty paragraph and other empty elements
// 10. cleanAttributes - Remove unnecessary attributes
//
// This function matches the JavaScript implementation exactly, including:
// - Same cleaning order and logic
// - Same conditional cleaning based on options
// - Same default behaviors for aggressive vs conservative cleaning.
func ExtractCleanNode(article *goquery.Selection, doc *goquery.Document, opts ContentCleanOptions) *goquery.Selection {
	if article == nil || article.Length() == 0 {
		return article
	}

	// Set default for DefaultCleaner to match JavaScript behavior
	// In JavaScript, defaultCleaner defaults to true if not specified
	defaultCleaner := true
	if opts.DefaultCleaner != nil {
		defaultCleaner = *opts.DefaultCleaner
	}
	preserve := opts.PreserveMatchers
	if len(opts.Preserve) != 0 {
		preserve = append([]goquery.Matcher(nil), preserve...)
		for _, selector := range opts.Preserve {
			// Closest uses Match; Single retains goquery's invalid-selector behavior.
			preserve = append(preserve, goquery.Single(selector))
		}
	}

	// Apply cleaning functions in the exact same order as JavaScript:
	// Unlike the document-level cleaning in the generic extractor,
	// we need to apply these operations specifically to the article scope

	// 1. Rewrite the tag name to div if it's a top level node like body or
	// html to avoid later complications with multiple body tags.
	article = rewriteTopLevelSelection(article)

	// 2. Drop small images and spacer images
	// Only do this if defaultCleaner is set to true;
	// this can sometimes be too aggressive.
	if defaultCleaner {
		cleanImagesInSelection(article, preserve)
	}

	// 3. Make links absolute
	if opts.URL != "" {
		makeLinksAbsoluteInSelection(article, opts.URL)
	}

	// 4. Mark elements to keep that would normally be removed.
	// E.g., stripJunkTags will remove iframes, so we're going to mark
	// YouTube/Vimeo videos as elements we want to keep.
	markToKeepInSelection(article, opts.URL)

	// 5. Drop certain tags like <title>, etc
	// This is -mostly- for cleanliness, not security.
	stripJunkTagsInSelection(article)

	// 6. H1 tags are typically the article title, which should be extracted
	// by the title extractor instead. If there's less than 3 of them (<3),
	// strip them. Otherwise, turn 'em into H2s.
	cleanHOnesInSelection(article, preserve)

	// 7. Clean headers
	cleanHeadersInSelection(article, opts.Title, preserve)

	// 8. We used to clean UL's and OL's here, but it was leading to
	// too many in-article lists being removed. Consider a better
	// way to detect menus particularly and remove them.
	// Also optionally running, since it can be overly aggressive.
	if defaultCleaner {
		cleanTagsInSelection(article, opts.CleanConditionally, preserve)
	}

	// 9. Remove empty paragraph nodes
	removeEmptyInSelection(article, preserve)

	// 10. Remove unnecessary attributes
	cleanAttributesInSelection(article)

	return article
}

// Helper functions that operate on selections instead of whole documents
// These mirror the JavaScript behavior of scoped operations

func rewriteTopLevelSelection(selection *goquery.Selection) *goquery.Selection {
	// If the selection is body or html, convert to div
	selection.Each(func(i int, s *goquery.Selection) {
		tagName := goquery.NodeName(s)
		if tagName == "body" || tagName == "html" {
			dom.SetAttr(s, "data-original-tag", tagName)
			s.Get(0).Data = "div"
		}
	})
	return selection
}

func cleanImagesInSelection(selection *goquery.Selection, preserve []goquery.Matcher) {
	selection.FindMatcher(contentImages).Each(func(i int, img *goquery.Selection) {
		if preserved(img, preserve) {
			return
		}
		// Remove spacer images (small or named spacer/blank)
		src, _ := img.Attr("src")
		width, _ := img.Attr("width")
		height, _ := img.Attr("height")

		// Check if it's a spacer by name
		lowerSrc := strings.ToLower(src)
		if strings.Contains(lowerSrc, "spacer") ||
			strings.Contains(lowerSrc, "blank") ||
			strings.Contains(lowerSrc, "clear.gif") {
			img.Remove()
			return
		}

		// Check if it's small (likely spacer)
		if width != "" && height != "" {
			w, errW := strconv.Atoi(width)
			h, errH := strconv.Atoi(height)
			if errW == nil && errH == nil && (w <= 1 || h <= 1) {
				img.Remove()
				return
			}
		}

		// Remove images with very small dimensions in style
		style, _ := img.Attr("style")
		if strings.Contains(style, "width:1px") || strings.Contains(style, "height:1px") {
			img.Remove()
		}
	})
}

func makeLinksAbsoluteInSelection(selection *goquery.Selection, baseURL string) {
	if baseURL == "" {
		return
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return
	}

	selection.FindMatcher(contentLinks).Each(func(i int, link *goquery.Selection) {
		href, exists := link.Attr("href")
		if !exists {
			return
		}

		resolved, err := base.Parse(href)
		if err == nil {
			link.SetAttr("href", resolved.String())
		}
	})

	selection.FindMatcher(contentImageSources).Each(func(i int, img *goquery.Selection) {
		src, exists := img.Attr("src")
		if !exists {
			return
		}

		resolved, err := base.Parse(src)
		if err == nil {
			img.SetAttr("src", resolved.String())
		}

		// Handle srcset
		if srcset, exists := img.Attr("srcset"); exists {
			parts := strings.Split(srcset, ",")
			var newParts []string
			for _, part := range parts {
				part = strings.TrimSpace(part)
				spaceIdx := strings.IndexByte(part, ' ')
				var srcPart, descriptor string
				if spaceIdx >= 0 {
					srcPart = part[:spaceIdx]
					descriptor = part[spaceIdx:]
				} else {
					srcPart = part
				}

				if resolved, err := base.Parse(srcPart); err == nil {
					newParts = append(newParts, resolved.String()+descriptor)
				} else {
					newParts = append(newParts, part)
				}
			}
			img.SetAttr("srcset", strings.Join(newParts, ", "))
		}
	})
}

func markToKeepInSelection(selection *goquery.Selection, baseURL string) {
	for _, matcher := range contentKeepMedia {
		selection.FindMatcher(matcher).AddClass("hermes-parser-keep")
	}

	// If we have a base URL, also mark iframes from the same domain
	if baseURL != "" {
		if parsed, err := url.Parse(baseURL); err == nil {
			var domainSelectorBuilder strings.Builder
			domainSelectorBuilder.WriteString("iframe[src^=\"")
			domainSelectorBuilder.WriteString(parsed.Scheme)
			domainSelectorBuilder.WriteString("://")
			domainSelectorBuilder.WriteString(parsed.Host)
			domainSelectorBuilder.WriteString("\"]")
			domainSelector := domainSelectorBuilder.String()
			selection.Find(domainSelector).AddClass("hermes-parser-keep")
		}
	}
}

func stripJunkTagsInSelection(selection *goquery.Selection) {
	selection.FindMatcher(contentJunk).NotMatcher(contentKeep).Remove()
}

func cleanHOnesInSelection(selection *goquery.Selection, preserve []goquery.Matcher) {
	h1s := selection.FindMatcher(contentHOnes)

	if h1s.Length() < 3 {
		h1s.Each(func(_ int, h1 *goquery.Selection) {
			if !preserved(h1, preserve) {
				h1.Remove()
			}
		})
	} else {
		// Convert H1s to H2s if there are 3 or more
		h1s.Each(func(i int, h1 *goquery.Selection) {
			h1.Get(0).Data = "h2"
		})
	}
}

func cleanHeadersInSelection(selection *goquery.Selection, title string, preserve []goquery.Matcher) {
	headers := selection.FindMatcher(contentHeaders)

	headers.Each(func(i int, header *goquery.Selection) {
		if preserved(header, preserve) {
			return
		}
		headerText := strings.TrimSpace(header.Text())

		// Remove headers that appear before all paragraphs
		allParagraphs := selection.FindMatcher(contentParagraphs)
		if allParagraphs.Length() > 0 {
			prevParagraphs := header.PrevAll().FilterMatcher(contentParagraphs)
			if prevParagraphs.Length() == 0 {
				header.Remove()
				return
			}
		}

		// Remove headers that match the title exactly
		if title != "" && headerText == title {
			header.Remove()
			return
		}

		// Remove very short headers
		if len(headerText) < 3 {
			header.Remove()
		}
	})
}

func cleanTagsInSelection(selection *goquery.Selection, cleanConditionally bool, preserve []goquery.Matcher) {
	if !cleanConditionally {
		return // Skip conditional cleaning
	}

	for _, matcher := range contentConditional {
		selection.FindMatcher(matcher).Each(func(i int, elem *goquery.Selection) {
			if preserved(elem, preserve) {
				return
			}
			// Skip if marked to keep
			if elem.HasClass("hermes-parser-keep") {
				return
			}

			// Skip if it contains elements marked to keep
			if elem.FindMatcher(contentKeep).Length() > 0 {
				return
			}

			// Basic heuristic: remove if mostly links
			text := strings.TrimSpace(elem.Text())

			// Don't remove empty elements that contain important media (iframe, video, etc.)
			if len(text) == 0 {
				// Check if it contains media elements that should be preserved
				if elem.FindMatcher(contentMedia).Length() > 0 {
					return // Preserve containers with media
				}
				elem.Remove()
				return
			}

			links := elem.FindMatcher(contentAnchors)
			var linkTextBuilder strings.Builder
			links.Each(func(j int, link *goquery.Selection) {
				linkTextBuilder.WriteString(strings.TrimSpace(link.Text()))
				linkTextBuilder.WriteString(" ")
			})
			linkText := linkTextBuilder.String()

			// If more than 50% of text is links, likely navigation/junk
			if len(strings.TrimSpace(linkText)) > len(text)/2 {
				elem.Remove()
			}
		})
	}
}

func removeEmptyInSelection(selection *goquery.Selection, preserve []goquery.Matcher) {
	// Remove empty paragraphs and other elements
	selection.FindMatcher(contentEmpty).Each(func(i int, elem *goquery.Selection) {
		if preserved(elem, preserve) {
			return
		}
		text := strings.TrimSpace(elem.Text())
		// Remove if empty or only whitespace/br tags
		if text == "" {
			// Check if it only contains br tags or whitespace
			html, _ := elem.Html()
			cleanHTML := strings.TrimSpace(html)
			if cleanHTML == "" || EMPTY_HTML_RE.MatchString(cleanHTML) {
				elem.Remove()
			}
		}
	})
}

func preserved(selection *goquery.Selection, matchers []goquery.Matcher) bool {
	for _, matcher := range matchers {
		if selection.ClosestMatcher(matcher).Length() != 0 {
			return true
		}
	}
	return false
}

func cleanAttributesInSelection(selection *goquery.Selection) {
	selection.FindMatcher(contentAll).Each(func(i int, elem *goquery.Selection) {
		// Get all current attributes
		node := elem.Get(0)
		if node == nil {
			return
		}

		// Collect attributes to remove
		var attrsToRemove []string
		for _, attr := range node.Attr {
			switch attr.Key {
			case "href", "src", "alt", "title", "srcset", "data-content-score", "class":
			default:
				attrsToRemove = append(attrsToRemove, attr.Key)
			}
		}

		// Remove unwanted attributes
		for _, attrName := range attrsToRemove {
			elem.RemoveAttr(attrName)
		}
	})
}
