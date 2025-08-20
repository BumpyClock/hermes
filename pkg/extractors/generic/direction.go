// ABOUTME: Direction detection extractor for RTL/LTR text analysis using Unicode blocks
// ABOUTME: 1:1 port of JavaScript string-direction library with exact behavioral compatibility

package generic

import (
	"errors"
	"regexp"
	"strings"
)

// Direction constants matching JavaScript string-direction library exactly
const (
	LTRMark = "\u200e" // Left-to-right mark
	RTLMark = "\u200f" // Right-to-left mark
	LTR     = "ltr"    // Left to right direction content
	RTL     = "rtl"    // Right to left direction content  
	BIDI    = "bidi"   // Both directions - bidirectional content
	NODI    = ""       // No direction - empty string for no detectable direction
)

// RTLScriptRange represents a Unicode block range for RTL scripts
type RTLScriptRange struct {
	From int // Starting Unicode code point
	To   int // Ending Unicode code point
}

// rtlScriptRanges defines Unicode blocks for right-to-left scripts
// Matches JavaScript string-direction library rtlSciriptRanges exactly
var rtlScriptRanges = map[string]RTLScriptRange{
	"Hebrew":   {0x0590, 0x05FF}, // 0590-05FF
	"Arabic":   {0x0600, 0x06FF}, // 0600-06FF  
	"NKo":      {0x07C0, 0x07FF}, // 07C0-07FF
	"Syriac":   {0x0700, 0x074F}, // 0700-074F
	"Thaana":   {0x0780, 0x07BF}, // 0780-07BF
	"Tifinagh": {0x2D30, 0x2D7F}, // 2D30-2D7F
}

// Regex for stripping whitespace and non-directional characters
// Matches JavaScript: /[\s\n\0\f\t\v\'\"\-0-9\+\?\!]+/gm
var stripNonDirectionalRegex = regexp.MustCompile(`[\s\n\x00\f\t\v'"\-0-9+?!]+`)

// GetDirection analyzes string direction and returns 'ltr', 'rtl', 'bidi', or ''
// Direct port of JavaScript stringDirection.getDirection() function
func GetDirection(input interface{}) (string, error) {
	// Type checking - matches JavaScript behavior exactly
	if input == nil {
		return "", errors.New("TypeError missing argument")
	}
	
	str, ok := input.(string)
	if !ok {
		return "", errors.New("TypeError getDirection expects strings")
	}
	
	// Empty string returns no direction - matches JavaScript
	if str == "" {
		return NODI, nil
	}
	
	// Check for explicit direction marks first - matches JavaScript priority
	hasLTRMark := strings.Contains(str, LTRMark)
	hasRTLMark := strings.Contains(str, RTLMark)
	
	if hasLTRMark && hasRTLMark {
		return BIDI, nil
	}
	if hasLTRMark {
		return LTR, nil
	}
	if hasRTLMark {
		return RTL, nil  
	}
	
	// Analyze character-level direction
	hasRTL := hasDirectionCharacters(str, RTL)
	hasLTRChars := hasDirectionCharacters(str, LTR)
	
	// Return direction based on character analysis - matches JavaScript logic
	if hasRTL && hasLTRChars {
		return BIDI, nil
	}
	if hasLTRChars {
		return LTR, nil
	}  
	if hasRTL {
		return RTL, nil
	}
	
	return NODI, nil
}

// hasDirectionCharacters determines if string has RTL or LTR characters
// Direct port of JavaScript hasDirectionCharacters() function
func hasDirectionCharacters(str string, direction string) bool {
	hasRTL := false
	hasLTRChars := false
	
	// Check for digits - matches JavaScript logic exactly
	hasDigit := strings.ContainsAny(str, "0123456789")
	
	// Remove whitespace and non-directional characters - matches JavaScript regex
	cleanStr := stripNonDirectionalRegex.ReplaceAllString(str, "")
	
	// Loop through each character - matches JavaScript character-by-character analysis
	for _, char := range cleanStr {
		charIsRTL := false
		
		// Test character against all RTL script ranges - matches JavaScript logic
		for _, scriptRange := range rtlScriptRanges {
			if isInScriptRange(char, scriptRange.From, scriptRange.To) {
				hasRTL = true
				charIsRTL = true
				break
			}
		}
		
		// If character is NOT RTL, it's LTR - matches JavaScript logic exactly
		if !charIsRTL {
			hasLTRChars = true
		}
	}
	
	// Return based on requested direction - matches JavaScript conditions
	if direction == RTL {
		return hasRTL
	}
	if direction == LTR {
		// JavaScript: return hasLtr || (!hasRtl && hasDigit)  
		// Digits count as LTR if no RTL characters exist
		return hasLTRChars || (!hasRTL && hasDigit)
	}
	
	return false
}

// isInScriptRange checks if character is within Unicode block range
// Direct port of JavaScript isInScriptRange() function  
func isInScriptRange(char rune, from, to int) bool {
	charCode := int(char)
	// JavaScript: charCode > fromCode && charCode < toCode
	// Note: JavaScript uses exclusive bounds, Go implementation matches exactly
	return charCode > from && charCode < to
}

// DirectionExtractor extracts text direction from title field only
// Matches JavaScript: direction: ({ title }) => stringDirection.getDirection(title)
func DirectionExtractor(params ExtractorParams) (string, error) {
	direction, err := GetDirection(params.Title)
	if err != nil {
		// Return empty direction on error to match JavaScript behavior
		return NODI, nil
	}
	return direction, nil
}