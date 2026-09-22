package text

import (
	"regexp"
	"strings"
	"testing"
)

func TestExcerptContent(t *testing.T) {
	// Test with 3 words from JavaScript test case
	content := " One  two three four five six, seven eight, nine, ten."
	result := ExcerptContent(content, 3)
	expected := "One two three"
	if result != expected {
		t.Errorf("ExcerptContent() = %q, want %q", result, expected)
	}

	result10 := ExcerptContent(content, 10)
	expected10 := "One two three four five six, seven eight, nine, ten."
	if result10 != expected10 {
		t.Errorf("ExcerptContent() = %q, want %q", result10, expected10)
	}
}

func TestExcerptWhitespaceCompatibility(t *testing.T) {
	inputs := []string{"", " \t\n", "one", " one two three ", "a\v b", "\u00a0one\u00a0two\u2003three\u00a0", "\xff a\xfe b"}
	for separator := 0; separator < 256; separator++ {
		inputs = append(inputs, " \u2003first"+string([]byte{byte(separator)})+"second\tthird \u00a0")
	}
	for _, input := range inputs {
		for _, limit := range []int{-1, 0, 1, 2, 10, 160, 1000} {
			if got, want := ExcerptContent(input, limit), originalExcerpt(input, limit); got != want {
				t.Fatalf("input=%q limit=%d: got %q, want %q", input, limit, got, want)
			}
		}
	}
}

var originalExcerptWhitespace = regexp.MustCompile(`\s+`)

func originalExcerpt(content string, limit int) string {
	if limit <= 0 {
		limit = 10
	}
	parts := originalExcerptWhitespace.Split(strings.TrimSpace(content), -1)
	var words []string
	for _, part := range parts {
		if part != "" {
			words = append(words, part)
		}
	}
	return strings.Join(words[:min(limit, len(words))], " ")
}

func FuzzExcerptCompatibility(f *testing.F) {
	for _, source := range []string{"One two three", "one\vword\u00a0still\tsecond", "\xff first\r\nsecond", "<p>Keep literal tags.</p>"} {
		f.Add(source, 2)
	}
	f.Fuzz(func(t *testing.T, source string, limit int) {
		if len(source) > 8192 {
			t.Skip()
		}
		if got, want := ExcerptContent(source, limit), originalExcerpt(source, limit); got != want {
			t.Fatalf("limit=%d: got %q, want %q", limit, got, want)
		}
	})
}

func BenchmarkExcerptBounded(b *testing.B) {
	source := strings.Repeat("A long article with ordinary words and whitespace. ", 2000)
	b.Run("original", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = originalExcerpt(source, 160)
		}
	})
	b.Run("bounded", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ExcerptContent(source, 160)
		}
	})
}
