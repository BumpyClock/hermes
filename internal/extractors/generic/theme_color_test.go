package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestGenericThemeColorExtractor(t *testing.T) {
	extractor := &GenericThemeColorExtractor{}

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Valid hex color",
			html:     `<meta name="theme-color" content="#007acc">`,
			expected: "#007acc",
		},
		{
			name:     "Valid short hex color",
			html:     `<meta name="theme-color" content="#0ac">`,
			expected: "#0ac",
		},
		{
			name:     "Valid RGB color",
			html:     `<meta name="theme-color" content="rgb(255, 0, 0)">`,
			expected: "rgb(255, 0, 0)",
		},
		{
			name:     "Valid RGBA color",
			html:     `<meta name="theme-color" content="rgba(255, 0, 0, 0.5)">`,
			expected: "rgba(255, 0, 0, 0.5)",
		},
		{
			name:     "Valid HSL color",
			html:     `<meta name="theme-color" content="hsl(120, 100%, 50%)">`,
			expected: "hsl(120, 100%, 50%)",
		},
		{
			name:     "Valid HSLA color",
			html:     `<meta name="theme-color" content="hsla(120, 100%, 50%, 0.3)">`,
			expected: "hsla(120, 100%, 50%, 0.3)",
		},
		{
			name:     "Valid named color",
			html:     `<meta name="theme-color" content="blue">`,
			expected: "blue",
		},
		{
			name:     "Case insensitive",
			html:     `<meta name="theme-color" content="#FF0000">`,
			expected: "#ff0000",
		},
		{
			name:     "Whitespace trimmed",
			html:     `<meta name="theme-color" content="  #007acc  ">`,
			expected: "#007acc",
		},
		{
			name:     "Microsoft tile color fallback",
			html:     `<meta name="msapplication-TileColor" content="#da532c">`,
			expected: "#da532c",
		},
		{
			name:     "Theme color takes priority over tile color",
			html:     `<meta name="theme-color" content="#007acc"><meta name="msapplication-TileColor" content="#da532c">`,
			expected: "#007acc",
		},
		{
			name:     "Invalid color format",
			html:     `<meta name="theme-color" content="invalid-color">`,
			expected: "",
		},
		{
			name:     "Empty content",
			html:     `<meta name="theme-color" content="">`,
			expected: "",
		},
		{
			name:     "No theme color meta",
			html:     `<meta name="description" content="A test page">`,
			expected: "",
		},
		{
			name:     "Invalid hex color",
			html:     `<meta name="theme-color" content="#gggggg">`,
			expected: "",
		},
		{
			name:     "Malformed RGB color",
			html:     `<meta name="theme-color" content="rgb(300, -10, abc)">`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := extractor.Extract(doc.Selection, "https://example.com", []string{})

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestValidateAndNormalizeColor(t *testing.T) {
	extractor := &GenericThemeColorExtractor{}

	tests := []struct {
		input    string
		expected string
	}{
		// Valid hex colors
		{"#fff", "#fff"},
		{"#ffffff", "#ffffff"},
		{"#FF0000", "#ff0000"},
		{"#abc123", "#abc123"},

		// Valid RGB colors
		{"rgb(255, 0, 0)", "rgb(255, 0, 0)"},
		{"rgba(255, 0, 0, 0.5)", "rgba(255, 0, 0, 0.5)"},
		{"rgb( 255 , 0 , 0 )", "rgb( 255 , 0 , 0 )"},

		// Valid HSL colors
		{"hsl(120, 100%, 50%)", "hsl(120, 100%, 50%)"},
		{"hsla(120, 100%, 50%, 0.3)", "hsla(120, 100%, 50%, 0.3)"},

		// Valid named colors
		{"red", "red"},
		{"blue", "blue"},
		{"transparent", "transparent"},

		// Invalid colors
		{"#gg", ""},
		{"#1234567", ""},
		{"rgb(300, -10, abc)", ""},
		{"invalid", ""},
		{"", ""},
		{"  ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractor.validateAndNormalizeColor(tt.input)
			if result != tt.expected {
				t.Errorf("validateAndNormalizeColor(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
