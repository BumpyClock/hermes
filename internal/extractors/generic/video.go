package generic

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// VideoMetadata contains structured video metadata.
type VideoMetadata struct {
	URL       string `json:"url,omitempty"`
	Type      string `json:"type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Duration  int    `json:"duration,omitempty"`
	SecureURL string `json:"secure_url,omitempty"`
}

// GenericVideoExtractor extracts video metadata from Open Graph and other meta tags.
type GenericVideoExtractor struct{}

// Extract extracts video metadata from the page.
func (extractor *GenericVideoExtractor) Extract(selection *goquery.Selection, pageURL string, metaCache []string) *VideoMetadata {
	video := &VideoMetadata{}
	hasVideoData := false

	// Extract Open Graph video URL
	if url := firstMetaValue(selection, []string{"og:video"}, nil); url != "" {
		video.URL = url
		hasVideoData = true
	}

	// Extract secure video URL (HTTPS)
	if secureURL := firstMetaValue(selection, []string{"og:video:secure_url"}, nil); secureURL != "" {
		video.SecureURL = secureURL
		hasVideoData = true
	}

	// Extract video type/MIME type
	if videoType := firstMetaValue(selection, []string{"og:video:type"}, nil); videoType != "" {
		video.Type = videoType
		hasVideoData = true
	}

	// Extract video width
	if widthStr := firstMetaValue(selection, []string{"og:video:width"}, nil); widthStr != "" {
		if width, err := strconv.Atoi(widthStr); err == nil && width > 0 {
			video.Width = width
			hasVideoData = true
		}
	}

	// Extract video height
	if heightStr := firstMetaValue(selection, []string{"og:video:height"}, nil); heightStr != "" {
		if height, err := strconv.Atoi(heightStr); err == nil && height > 0 {
			video.Height = height
			hasVideoData = true
		}
	}

	// Extract video duration (in seconds)
	if durationStr := firstMetaValue(selection, []string{"og:video:duration"}, nil); durationStr != "" {
		if duration, err := strconv.Atoi(durationStr); err == nil && duration > 0 {
			video.Duration = duration
			hasVideoData = true
		}
	}

	// Try alternative meta tag formats if no Open Graph data found
	if !hasVideoData {
		// Try Twitter video tags
		if url := firstMetaValue(selection, []string{"twitter:player"}, nil); url != "" {
			video.URL = url
			hasVideoData = true

			// Get Twitter video dimensions
			if widthStr := firstMetaValue(selection, []string{"twitter:player:width"}, nil); widthStr != "" {
				if width, err := strconv.Atoi(widthStr); err == nil && width > 0 {
					video.Width = width
				}
			}
			if heightStr := firstMetaValue(selection, []string{"twitter:player:height"}, nil); heightStr != "" {
				if height, err := strconv.Atoi(heightStr); err == nil && height > 0 {
					video.Height = height
				}
			}
		}

		// Try JSON-LD structured data for VideoObject
		video = extractor.extractFromJSONLD(selection, video)
	}

	// Return nil if no video metadata was found
	if !hasVideoData && video.URL == "" {
		return nil
	}

	// Validate and clean up URLs
	if video.URL != "" {
		video.URL = extractor.normalizeURL(video.URL, pageURL)
	}
	if video.SecureURL != "" {
		video.SecureURL = extractor.normalizeURL(video.SecureURL, pageURL)
	}

	return video
}

// extractFromJSONLD attempts to extract video information from JSON-LD structured data.
func (extractor *GenericVideoExtractor) extractFromJSONLD(selection *goquery.Selection, video *VideoMetadata) *VideoMetadata {
	// This is a simplified implementation - a full implementation would parse JSON-LD
	// Looking for script[type="application/ld+json"] with VideoObject
	// For now, we'll skip this complex parsing and return the existing video metadata
	return video
}

// normalizeURL ensures the video URL is absolute.
func (extractor *GenericVideoExtractor) normalizeURL(videoURL, pageURL string) string {
	videoURL = strings.TrimSpace(videoURL)

	// Already absolute
	if strings.HasPrefix(videoURL, "http://") || strings.HasPrefix(videoURL, "https://") {
		return videoURL
	}

	// Protocol-relative
	if strings.HasPrefix(videoURL, "//") {
		return "https:" + videoURL
	}

	// Parse the base URL
	baseURL, err := url.Parse(pageURL)
	if err != nil {
		// If we can't parse the page URL, return the videoURL as-is
		return videoURL
	}

	// Parse the relative URL
	relativeURL, err := url.Parse(videoURL)
	if err != nil {
		// If we can't parse the videoURL, return it as-is
		return videoURL
	}

	// Resolve the relative URL against the base URL
	resolved := baseURL.ResolveReference(relativeURL)
	return resolved.String()
}

// ExtractVideoURL is a convenience function that returns just the primary video URL.
func (extractor *GenericVideoExtractor) ExtractVideoURL(selection *goquery.Selection, pageURL string, metaCache []string) string {
	video := extractor.Extract(selection, pageURL, metaCache)
	if video == nil {
		return ""
	}

	// Prefer secure URL if available
	if video.SecureURL != "" {
		return video.SecureURL
	}

	return video.URL
}
