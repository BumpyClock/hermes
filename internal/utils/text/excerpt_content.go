package text

import (
	"strings"
)

func ExcerptContent(content string, words ...int) string {
	wordCount := 10
	if len(words) > 0 && words[0] > 0 {
		wordCount = words[0]
	}

	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ""
	}

	var excerpt strings.Builder
	for start := 0; start < len(trimmed); {
		end := start
		for end < len(trimmed) && !excerptWhitespace(trimmed[end]) {
			end++
		}
		if start == 0 && end == len(trimmed) {
			return trimmed
		}
		if excerpt.Len() != 0 {
			excerpt.WriteByte(' ')
		}
		excerpt.WriteString(trimmed[start:end])
		wordCount--
		if wordCount == 0 {
			break
		}
		start = end
		for start < len(trimmed) && excerptWhitespace(trimmed[start]) {
			start++
		}
	}
	return excerpt.String()
}

func excerptWhitespace(char byte) bool {
	// Match RE2 \s, not Unicode whitespace or vertical tabs.
	return char == ' ' || char == '\t' || char == '\n' || char == '\f' || char == '\r'
}
