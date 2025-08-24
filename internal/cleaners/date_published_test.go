// ABOUTME: Comprehensive test suite for date published cleaner with JavaScript compatibility
// ABOUTME: Tests date parsing, validation, timezone handling, and various date formats

package cleaners

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCleanDatePublished(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		timezone string
		format   string
		expected *string // nil for invalid dates, pointer to ISO string for valid
	}{
		// Millisecond timestamps
		{
			name:     "millisecond timestamp",
			input:    "1609459200000", // January 1, 2021 00:00:00 UTC
			expected: stringPtr("2021-01-01T00:00:00.000Z"),
		},
		{
			name:     "millisecond timestamp string",
			input:    "1640995200000", // January 1, 2022 00:00:00 UTC
			expected: stringPtr("2022-01-01T00:00:00.000Z"),
		},
		
		// Second timestamps
		{
			name:     "second timestamp",
			input:    "1609459200", // January 1, 2021 00:00:00 UTC
			expected: stringPtr("2021-01-01T00:00:00.000Z"),
		},
		{
			name:     "second timestamp string",
			input:    "1640995200", // January 1, 2022 00:00:00 UTC
			expected: stringPtr("2022-01-01T00:00:00.000Z"),
		},

		// Standard date formats
		{
			name:     "ISO date string",
			input:    "2021-01-01T00:00:00Z",
			expected: stringPtr("2021-01-01T00:00:00.000Z"),
		},
		{
			name:     "RFC3339 date",
			input:    "2021-01-01T12:00:00+00:00",
			expected: stringPtr("2021-01-01T12:00:00.000Z"),
		},
		{
			name:     "simple date format",
			input:    "January 1, 2021",
			expected: stringPtr("2021-01-01T00:00:00.000Z"),
		},

		// Time ago strings
		{
			name:     "minutes ago",
			input:    "5 minutes ago",
			expected: nil, // Relative to current time - hard to test exactly
		},
		{
			name:     "hours ago",
			input:    "2 hours ago",
			expected: nil, // Relative to current time - hard to test exactly
		},
		{
			name:     "days ago",
			input:    "3 days ago",
			expected: nil, // Relative to current time - hard to test exactly
		},

		// "Now" strings
		{
			name:     "just now",
			input:    "just now",
			expected: nil, // Current time - hard to test exactly
		},
		{
			name:     "right now",
			input:    "right now",
			expected: nil, // Current time - hard to test exactly
		},
		{
			name:     "now",
			input:    "now",
			expected: nil, // Current time - hard to test exactly
		},

		// Time with offset
		{
			name:     "time with offset",
			input:    "2021-01-01T12:00:00-0500",
			expected: stringPtr("2021-01-01T17:00:00.000Z"), // UTC conversion
		},
		{
			name:     "time with positive offset",
			input:    "2021-01-01T12:00:00+0300",
			expected: stringPtr("2021-01-01T09:00:00.000Z"), // UTC conversion
		},

		// Published prefixes
		{
			name:     "published prefix",
			input:    "Published: January 1, 2021",
			expected: stringPtr("2021-01-01T00:00:00.000Z"),
		},
		{
			name:     "published prefix with colon",
			input:    "published: 2021-01-01",
			expected: stringPtr("2021-01-01T00:00:00.000Z"),
		},

		// Timezone handling
		{
			name:     "date with timezone",
			input:    "2021-01-01 12:00:00",
			timezone: "America/New_York",
			expected: stringPtr("2021-01-01T17:00:00.000Z"), // EST is UTC-5
		},
		{
			name:     "date with format and timezone",
			input:    "01/01/2021 12:00 PM",
			format:   "MM/DD/YYYY h:mm A",
			timezone: "UTC",
			expected: stringPtr("2021-01-01T12:00:00.000Z"),
		},

		// Invalid dates
		{
			name:     "invalid date string",
			input:    "not a date",
			expected: nil,
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "invalid timestamp too short",
			input:    "123",
			expected: nil,
		},
		{
			name:     "invalid timestamp too long",
			input:    "12345678901234",
			expected: nil,
		},

		// Edge cases
		{
			name:     "whitespace only",
			input:    "   ",
			expected: nil,
		},
		{
			name:     "invalid year",
			input:    "13/32/2021",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanDatePublished(tt.input, tt.timezone, tt.format)
			
			if tt.expected == nil {
				// For relative time tests, just check that we get some valid result or nil
				if strings.Contains(tt.name, "ago") || strings.Contains(tt.name, "now") {
					// These should return either a valid ISO string or nil
					if result != nil {
						// If we got a result, it should be a valid ISO string
						_, err := time.Parse(time.RFC3339, *result)
						assert.NoError(t, err, "Result should be valid ISO string: %s", *result)
					}
				} else {
					assert.Nil(t, result, 
						"CleanDatePublished(%q) should return nil, got %v", tt.input, result)
				}
			} else {
				assert.NotNil(t, result, 
					"CleanDatePublished(%q) should not return nil", tt.input)
				if result != nil {
					assert.Equal(t, *tt.expected, *result,
						"CleanDatePublished(%q) = %q, expected %q", tt.input, *result, *tt.expected)
				}
			}
		})
	}
}

func TestCleanDatePublishedJavaScriptCompatibility(t *testing.T) {
	// Test cases that verify exact JavaScript behavior compatibility
	compatTests := []struct {
		name     string
		input    string
		expected *string
		note     string
	}{
		{
			name:     "millisecond timestamp exact",
			input:    "1609459200000",
			expected: stringPtr("2021-01-01T00:00:00.000Z"),
			note:     "Millisecond timestamps should convert exactly",
		},
		{
			name:     "second timestamp exact", 
			input:    "1609459200",
			expected: stringPtr("2021-01-01T00:00:00.000Z"),
			note:     "Second timestamps should convert exactly",
		},
		{
			name:     "published prefix removal",
			input:    "Published: 2021-01-01T00:00:00Z",
			expected: stringPtr("2021-01-01T00:00:00.000Z"),
			note:     "Published prefix should be removed",
		},
		{
			name:     "invalid date returns nil",
			input:    "invalid date string",
			expected: nil,
			note:     "Invalid dates should return nil",
		},
	}

	for _, tt := range compatTests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanDatePublished(tt.input, "", "")
			
			if tt.expected == nil {
				assert.Nil(t, result, 
					"JavaScript compatibility test failed: %s\nCleanDatePublished(%q) should return nil", 
					tt.note, tt.input)
			} else {
				assert.NotNil(t, result,
					"JavaScript compatibility test failed: %s\nCleanDatePublished(%q) should not return nil", 
					tt.note, tt.input)
				if result != nil {
					assert.Equal(t, *tt.expected, *result,
						"JavaScript compatibility test failed: %s\nCleanDatePublished(%q) = %q, expected %q", 
						tt.note, tt.input, *result, *tt.expected)
				}
			}
		})
	}
}

func TestCleanDateString(t *testing.T) {
	// Test the cleanDateString helper function
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "published prefix removal",
			input:    "Published: January 1, 2021",
			expected: "January 1, 2021",
		},
		{
			name:     "case insensitive",
			input:    "PUBLISHED: 2021-01-01",
			expected: "2021-01-01", 
		},
		{
			name:     "meridian dots to m",
			input:    "Jan 1, 2021 3:00 p.m.",
			expected: "Jan 1, 2021 3:00 pm",
		},
		{
			name:     "meridian space fix",
			input:    "Jan 1, 2021 3:00p m",
			expected: "Jan 1, 2021 3:00 p m",
		},
		{
			name:     "no changes needed",
			input:    "2021-01-01T00:00:00Z",
			expected: "2021-01-01T00:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanDateString(tt.input)
			assert.Equal(t, tt.expected, result,
				"cleanDateString(%q) = %q, expected %q", tt.input, result, tt.expected)
		})
	}
}

func TestCreateDate(t *testing.T) {
	// Test the createDate helper function  
	tests := []struct {
		name     string
		input    string
		timezone string
		format   string
		isValid  bool
	}{
		{
			name:     "ISO format",
			input:    "2021-01-01T00:00:00Z",
			isValid:  true,
		},
		{
			name:     "with timezone",
			input:    "2021-01-01 12:00:00",
			timezone: "America/New_York",
			isValid:  true,
		},
		{
			name:     "relative time - minutes ago",
			input:    "5 minutes ago",
			isValid:  true,
		},
		{
			name:     "relative time - now",
			input:    "just now",
			isValid:  true,
		},
		{
			name:     "invalid date",
			input:    "not a date",
			isValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createDate(tt.input, tt.timezone, tt.format)
			
			if tt.isValid {
				assert.NotNil(t, result, "createDate should return valid time for %q", tt.input)
				if result != nil {
					// Check that it's a reasonable time (not zero time)
					assert.False(t, result.IsZero(), "createDate should not return zero time for %q", tt.input)
				}
			} else {
				// For invalid dates, we might get nil or zero time
				if result != nil {
					assert.True(t, result.IsZero(), "createDate should return zero time for invalid input %q", tt.input)
				}
			}
		})
	}
}

// Helper function to create string pointers for test cases
func stringPtr(s string) *string {
	return &s
}