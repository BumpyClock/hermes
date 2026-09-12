package generic

import (
	"strconv"

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
	}

	// Return nil if no video metadata was found
	if !hasVideoData && video.URL == "" {
		return nil
	}

	// Validate and clean up URLs
	if video.URL != "" {
		video.URL = normalizeResourceURL(video.URL, pageURL)
	}
	if video.SecureURL != "" {
		video.SecureURL = normalizeResourceURL(video.SecureURL, pageURL)
	}

	return video
}
