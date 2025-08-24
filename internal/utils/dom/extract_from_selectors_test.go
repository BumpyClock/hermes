// ABOUTME: Tests for the ExtractFromSelectors function
// ABOUTME: Validates CSS selector content extraction with JavaScript compatibility

package dom

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractFromSelectors(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		selectors   []string
		maxChildren int
		textOnly    bool
		expected    string
		expectNil   bool
	}{
		{
			name: "extracts an arbitrary node by selector",
			html: `
				<html>
					<div class="author">Adam</div>
				</html>
			`,
			selectors:   []string{".author"},
			maxChildren: 1,
			textOnly:    true,
			expected:    "Adam",
			expectNil:   false,
		},
		{
			name: "ignores comments",
			html: `
				<html>
					<div class="comments-section">
						<div class="author">Adam</div>
					</div>
				</html>
			`,
			selectors:   []string{".author"},
			maxChildren: 1,
			textOnly:    true,
			expected:    "",
			expectNil:   true,
		},
		{
			name: "skips a selector if it matches multiple nodes",
			html: `
				<html>
					<div>
						<div class="author">Adam</div>
						<div class="author">Adam</div>
					</div>
				</html>
			`,
			selectors:   []string{".author"},
			maxChildren: 1,
			textOnly:    true,
			expected:    "",
			expectNil:   true,
		},
		{
			name: "skips a node with too many children",
			html: `
				<html>
					<div>
						<div class="author">
							<span>Adam</span>
							<span>Pash</span>
						</div>
					</div>
				</html>
			`,
			selectors:   []string{".author"},
			maxChildren: 1,
			textOnly:    true,
			expected:    "",
			expectNil:   true,
		},
		{
			name: "returns HTML when textOnly is false",
			html: `
				<html>
					<div class="content"><strong>Bold</strong> text</div>
				</html>
			`,
			selectors:   []string{".content"},
			maxChildren: 1,
			textOnly:    false,
			expected:    "<strong>Bold</strong> text",
			expectNil:   false,
		},
		{
			name: "handles multiple selectors - first match wins",
			html: `
				<html>
					<div class="author">First Author</div>
					<div class="writer">Second Author</div>
				</html>
			`,
			selectors:   []string{".missing", ".author", ".writer"},
			maxChildren: 1,
			textOnly:    true,
			expected:    "First Author",
			expectNil:   false,
		},
		{
			name: "handles meta tags (which have empty text)",
			html: `
				<html>
					<meta name="author" content="John Doe" />
				</html>
			`,
			selectors:   []string{"meta[name='author']"},
			maxChildren: 1,
			textOnly:    true,
			expected:    "",
			expectNil:   true,
		},
		{
			name: "respects maxChildren parameter",
			html: `
				<html>
					<div class="content">
						<p>Para 1</p>
						<p>Para 2</p>
						<p>Para 3</p>
					</div>
				</html>
			`,
			selectors:   []string{".content"},
			maxChildren: 5,
			textOnly:    true,
			expected:    "Para 1 Para 2 Para 3",
			expectNil:   false,
		},
		{
			name: "returns nil for empty content",
			html: `
				<html>
					<div class="empty"></div>
				</html>
			`,
			selectors:   []string{".empty"},
			maxChildren: 1,
			textOnly:    true,
			expected:    "",
			expectNil:   true,
		},
		{
			name: "handles whitespace-only content as empty after normalization",
			html: `
				<html>
					<div class="whitespace">   
	   </div>
				</html>
			`,
			selectors:   []string{".whitespace"},
			maxChildren: 1,
			textOnly:    true,
			expected:    "",
			expectNil:   true,
		},
		{
			name: "extracts from complex selectors",
			html: `
				<html>
					<article>
						<header>
							<h1 class="title">Article Title</h1>
						</header>
					</article>
				</html>
			`,
			selectors:   []string{"article header h1.title"},
			maxChildren: 1,
			textOnly:    true,
			expected:    "Article Title",
			expectNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := ExtractFromSelectors(doc.Selection, tt.selectors, tt.maxChildren, tt.textOnly)

			if tt.expectNil {
				if result != nil {
					t.Errorf("Expected nil, got %q", *result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected %q, got nil", tt.expected)
				} else if *result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, *result)
				}
			}
		})
	}
}

func TestExtractFromSelectorsDefaultParameters(t *testing.T) {
	html := `
		<html>
			<div class="author">Test Author</div>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// Test with default parameters (maxChildren=1, textOnly=true)
	result := ExtractFromSelectors(doc.Selection, []string{".author"}, 1, true)
	if result == nil {
		t.Errorf("Expected 'Test Author', got nil")
	} else if *result != "Test Author" {
		t.Errorf("Expected 'Test Author', got %q", *result)
	}
}

func TestIsGoodNode(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		selector    string
		maxChildren int
		expected    bool
	}{
		{
			name: "good node with no children",
			html: `
				<html>
					<div class="test">Content</div>
				</html>
			`,
			selector:    ".test",
			maxChildren: 1,
			expected:    true,
		},
		{
			name: "good node with allowed children",
			html: `
				<html>
					<div class="test">
						<span>Child</span>
					</div>
				</html>
			`,
			selector:    ".test",
			maxChildren: 1,
			expected:    true,
		},
		{
			name: "bad node with too many children",
			html: `
				<html>
					<div class="test">
						<span>Child 1</span>
						<span>Child 2</span>
					</div>
				</html>
			`,
			selector:    ".test",
			maxChildren: 1,
			expected:    false,
		},
		{
			name: "bad node within comment section",
			html: `
				<html>
					<div class="comments">
						<div class="test">Content</div>
					</div>
				</html>
			`,
			selector:    ".test",
			maxChildren: 1,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			node := doc.Find(tt.selector)
			if node.Length() == 0 {
				t.Fatalf("Node not found with selector %s", tt.selector)
			}

			result := isGoodNode(node, tt.maxChildren)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}