package text

import (
	"regexp"
	"strings"
)

var whitespaceRegex = regexp.MustCompile(`\s+`)

func ExcerptContent(content string, words ...int) string {
	wordCount := 10
	if len(words) > 0 && words[0] > 0 {
		wordCount = words[0]
	}

	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ""
	}

	wordSlice := whitespaceRegex.Split(trimmed, -1)

	var filteredWords []string
	for _, word := range wordSlice {
		if word != "" {
			filteredWords = append(filteredWords, word)
		}
	}

	if wordCount > len(filteredWords) {
		wordCount = len(filteredWords)
	}

	return strings.Join(filteredWords[:wordCount], " ")
}
