package generic

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordCountExtractor(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "simple text",
			content:  "<div>Hello world test content</div>",
			expected: 4,
		},
		{
			name:     "nested inline tags",
			content:  "<div>The <strong>quick</strong> brown <em>fox</em> jumps</div>",
			expected: 5,
		},
		{
			name:     "html without div falls back",
			content:  "Text <span>with</span> <strong>tags</strong> removed",
			expected: 4,
		},
		{
			name:     "empty content",
			content:  "<div></div>",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenericWordCountExtractor.Extract(map[string]interface{}{"content": tt.content})
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWordCountExtractorInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
	}{
		{name: "nil input", input: nil},
		{name: "missing content", input: map[string]interface{}{"other": "value"}},
		{name: "non-string content", input: map[string]interface{}{"content": 123}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, 0, GenericWordCountExtractor.Extract(tt.input))
		})
	}
}

func TestWordCountExtractorLargeContent(t *testing.T) {
	var builder strings.Builder
	builder.WriteString("<div>")
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(&builder, "<p>This is paragraph number %d with content to count.</p>", i)
	}
	builder.WriteString("</div>")

	result := GenericWordCountExtractor.Extract(map[string]interface{}{"content": builder.String()})
	assert.Greater(t, result, 5000)
	assert.Less(t, result, 10000)
}

func BenchmarkWordCountExtractor(b *testing.B) {
	content := "<div>The quick brown fox jumps over the lazy dog multiple times in this test content.</div>"

	for i := 0; i < b.N; i++ {
		GenericWordCountExtractor.Extract(map[string]interface{}{"content": content})
	}
}
