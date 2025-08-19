package text

import "testing"

func TestExcerptContent(t *testing.T) {
	// Test with 3 words from JavaScript test case
	content := " One  two three four five six, seven eight, nine, ten."
	result := ExcerptContent(content, 3)
	expected := "One two three"
	if result != expected {
		t.Errorf("ExcerptContent() = %q, want %q", result, expected)
	}

	// Test with 10 words from JavaScript test case
	result10 := ExcerptContent(content, 10)
	expected10 := "One two three four five six, seven eight, nine, ten."
	if result10 != expected10 {
		t.Errorf("ExcerptContent() = %q, want %q", result10, expected10)
	}
}
