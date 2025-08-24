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
				assert.True(t, text.IsValidDate(result), "Parsed date should be valid")
			} else {
				assert.Error(t, err, "Expected parsing to fail for: %s", tt.input)
			}
		})
	}
}

func TestParseDateFromMeta(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"ISO 8601", "2023-04-15T10:30:00.000Z", true},
		{"Simple ISO", "2023-04-15T10:30:00Z", true},
		{"Date Only", "2023-04-15", true},
		{"With Milliseconds", "2023-04-15T10:30:00.123Z", true},
		{"Invalid Format", "April 15, 2023", true}, // Should fallback to general parser
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := text.ParseDateFromMeta(tt.input)
			
			if tt.valid {
				require.NoError(t, err)
				assert.NotNil(t, result)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestIsValidDate(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name  string
		date  *time.Time
		valid bool
	}{
		{"Nil Date", nil, false},
		{"Valid Recent Date", &now, true},
		{"Valid Past Date", func() *time.Time { d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC); return &d }(), true},
		{"Too Old Date", func() *time.Time { d := time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC); return &d }(), false},
		{"Future Date", func() *time.Time { d := now.Add(48 * time.Hour); return &d }(), false},
		{"Near Future Date", func() *time.Time { d := now.Add(12 * time.Hour); return &d }(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := text.IsValidDate(tt.date)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestFormatDateForJSON(t *testing.T) {
	// Test with specific date
	testDate := time.Date(2023, 4, 15, 10, 30, 45, 123456789, time.UTC)
	
	result := text.FormatDateForJSON(&testDate)
	expected := "2023-04-15T10:30:45.123Z"
	
	assert.Equal(t, expected, result)
	
	// Test with nil
	result = text.FormatDateForJSON(nil)
	assert.Equal(t, "", result)
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
			assert.True(t, text.IsValidDate(result))
		})
	}
}