package text_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BumpyClock/hermes/internal/utils/text"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool // true if should parse successfully
	}{
		{"RFC3339", "2023-04-15T10:30:00Z", true},
		{"ISO Date", "2023-04-15", true},
		{"US Format", "04/15/2023", true},
		{"Human Readable", "April 15, 2023", true},
		{"With Time", "2023-04-15 10:30:00", true},
		{"Empty String", "", false},
		{"Invalid Date", "not a date", false},
		{"Partial Date", "April 15", true}, // Should parse with current year
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := text.ParseDate(tt.input)

			if tt.expected {
				require.NoError(t, err, "Expected successful parsing for: %s", tt.input)
				assert.NotNil(t, result)
				assert.True(t, isValidDate(result), "Parsed date should be valid")
			} else {
				assert.Error(t, err, "Expected parsing to fail for: %s", tt.input)
			}
		})
	}
}

func TestDateParsingCompatibility(t *testing.T) {
	// Test formats commonly found in web content
	commonFormats := []string{
		"2023-04-15T10:30:00.000Z",
		"April 15, 2023",
		"15 April 2023",
		"04/15/2023",
		"4/15/2023",
		"2023-04-15",
		"Published: April 15, 2023",
		"Date: 2023-04-15T10:30:00Z",
	}

	for _, dateStr := range commonFormats {
		t.Run(dateStr, func(t *testing.T) {
			result, err := text.ParseDate(dateStr)
			assert.NoError(t, err, "Should parse: %s", dateStr)
			assert.NotNil(t, result)
			assert.True(t, isValidDate(result))
		})
	}

}

func isValidDate(date *time.Time) bool {
	if date == nil {
		return false
	}
	now := time.Now()
	minDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	maxDate := now.Add(24 * time.Hour)
	return date.After(minDate) && date.Before(maxDate)
}
