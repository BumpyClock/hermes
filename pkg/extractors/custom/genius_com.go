// ABOUTME: Genius.com custom extractor with lyrics, annotation support, and JSON metadata extraction
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/genius.com/index.js

package custom

import (
	"encoding/json"
)

// GeniusCustomExtractor provides the custom extraction rules for genius.com
// JavaScript equivalent: export const GeniusComExtractor = { ... }
var GeniusCustomExtractor = &CustomExtractor{
	Domain: "genius.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"h2 a",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".lyrics",
			},
		},
		
		// No transforms needed for Genius
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - empty for Genius
		Clean: []string{},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			// Complex JSON extraction with transform function
			[]interface{}{
				"meta[itemprop=page_data]",
				"value",
				transformGeniusDateFromJSON,
			},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			// Complex JSON extraction with transform function  
			[]interface{}{
				"meta[itemprop=page_data]",
				"value",
				transformGeniusImageFromJSON,
			},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			// enter selectors - empty in JavaScript
		},
	},
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// transformGeniusDateFromJSON extracts release date from Genius JSON metadata
// JavaScript equivalent: res => { const json = JSON.parse(res); return json.song.release_date; }
func transformGeniusDateFromJSON(jsonStr string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return ""
	}
	
	// Navigate: json.song.release_date
	if song, ok := data["song"].(map[string]interface{}); ok {
		if releaseDate, ok := song["release_date"].(string); ok {
			return releaseDate
		}
	}
	
	return ""
}

// transformGeniusImageFromJSON extracts album cover art URL from Genius JSON metadata
// JavaScript equivalent: res => { const json = JSON.parse(res); return json.song.album.cover_art_url; }
func transformGeniusImageFromJSON(jsonStr string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return ""
	}
	
	// Navigate: json.song.album.cover_art_url
	if song, ok := data["song"].(map[string]interface{}); ok {
		if album, ok := song["album"].(map[string]interface{}); ok {
			if coverArtURL, ok := album["cover_art_url"].(string); ok {
				return coverArtURL
			}
		}
	}
	
	return ""
}

// GetGeniusExtractor returns the Genius custom extractor
func GetGeniusExtractor() *CustomExtractor {
	return GeniusCustomExtractor
}