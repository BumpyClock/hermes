// ABOUTME: Tests for text processing constants and regex patterns
// ABOUTME: Validates regex patterns match expected patterns from JavaScript version
package text

import "testing"

func TestPageInHrefRegex(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
		desc     string
	}{
		{"page=1", true, "page=1 should match"},
		{"pg=1", true, "pg=1 should match"},
		{"p=1", true, "p=1 should match"},
		{"paging=12", true, "paging=12 should match"},
		{"pag=7", true, "pag=7 should match"},
		{"pagination/1", true, "pagination/1 should match"},
		{"paging/88", true, "paging/88 should match"},
		{"pa/83", true, "pa/83 should match"},
		{"p/11", true, "p/11 should match"},
		{"pg=102", true, "pg=102 should match (regex allows 1-3 digits, app logic filters >= 100)"},
		{"page:2", false, "page:2 should not match (wrong separator)"},
		{"randomtext", false, "randomtext should not match"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := PAGE_IN_HREF_RE.MatchString(tc.input)
			if result != tc.expected {
				t.Errorf("PAGE_IN_HREF_RE.MatchString(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}
