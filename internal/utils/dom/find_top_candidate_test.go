package dom

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestFindTopCandidate(t *testing.T) {
	tests := []struct {
		name           string
		html          string
		expectedTag    string
		expectedClass  string
		expectedText   string
		description    string
	}{
		{
			name: "single candidate with score",
			html: `<html><body>
				<div class="content" score="50">This is the main content with good score</div>
				<div class="sidebar" score="10">This is sidebar content</div>
			</body></html>`,
			expectedTag:   "div",
			expectedClass: "content",
			expectedText:  "This is the main content with good score",
			description:   "Should select the element with the highest score",
		},
		{
			name: "multiple candidates different scores",
			html: `<html><body>
				<div score="25">Content with medium score</div>
				<div score="100">Content with highest score</div>
				<div score="5">Content with low score</div>
			</body></html>`,
			expectedTag:  "div",
			expectedText: "Content with highest score",
			description:  "Should select element with highest score among multiple candidates",
		},
		{
			name: "filter out non-candidate tags",
			html: `<html><body>
				<br score="100">
				<hr score="90">
				<img score="80" alt="image">
				<div score="30">This should be selected</div>
			</body></html>`,
			expectedTag:  "div",
			expectedText: "This should be selected",
			description:  "Should ignore non-candidate tags like br, hr, img even with high scores",
		},
		{
			name: "no scored elements - fallback to body",
			html: `<html><body>
				<div>No score here</div>
				<p>No score here either</p>
			</body></html>`,
			expectedTag: "body",
			description: "Should fallback to body element when no scored elements found",
		},
		{
			name: "no body - fallback to first element",
			html: `<html>
				<div>First element</div>
				<p>Second element</p>
			</html>`,
			expectedTag: "body", // goquery automatically adds body even when not present
			description: "Should fallback to body element (auto-added by goquery) when no scored elements",
		},
		{
			name: "tie in scores - first wins",
			html: `<html><body>
				<div score="50">First with score 50</div>
				<div score="50">Second with score 50</div>
			</body></html>`,
			expectedTag:  "div",
			expectedText: "First with score 50",
			description:  "When scores are tied, should select the first element encountered",
		},
		{
			name: "zero scores ignored",
			html: `<html><body>
				<div score="0">Zero score content</div>
				<div score="25">Positive score content</div>
			</body></html>`,
			expectedTag:  "div",
			expectedText: "Positive score content",
			description:  "Should ignore elements with zero scores",
		},
		{
			name: "negative scores handled",
			html: `<html><body>
				<div score="-10">Negative score content</div>
				<div score="5">Low positive score content</div>
			</body></html>`,
			expectedTag:  "div",
			expectedText: "Low positive score content",
			description:  "Should prefer positive scores over negative ones",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			candidate := FindTopCandidate(doc)

			// Verify we got a result
			if candidate.Length() == 0 {
				t.Fatal("FindTopCandidate returned empty selection")
			}

			// Check tag name
			actualTag := strings.ToLower(goquery.NodeName(candidate))
			if actualTag != tt.expectedTag {
				t.Errorf("Expected tag %s, got %s", tt.expectedTag, actualTag)
			}

			// Check class if specified
			if tt.expectedClass != "" {
				actualClass, _ := candidate.Attr("class")
				if actualClass != tt.expectedClass {
					t.Errorf("Expected class %s, got %s", tt.expectedClass, actualClass)
				}
			}

			// Check text content if specified
			if tt.expectedText != "" {
				actualText := strings.TrimSpace(candidate.Text())
				if actualText != tt.expectedText {
					t.Errorf("Expected text '%s', got '%s'", tt.expectedText, actualText)
				}
			}
		})
	}
}

func TestFindTopCandidateWithMergeSiblings(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		description string
		verify      func(t *testing.T, candidate *goquery.Selection)
	}{
		{
			name: "candidate with high score gets merged with siblings",
			html: `<html><body><div class="container">
				<div class="content" score="100">Main content with high score</div>
				<div class="related" score="30">Related content with decent score</div>
				<div class="footer" score="5">Footer with low score</div>
			</div></body></html>`,
			description: "Should merge high-scoring siblings with the main candidate",
			verify: func(t *testing.T, candidate *goquery.Selection) {
				// The candidate should be processed through MergeSiblings
				if candidate.Length() == 0 {
					t.Fatal("No candidate found")
				}
				// Basic verification that we got a reasonable result
				text := strings.TrimSpace(candidate.Text())
				if !strings.Contains(text, "Main content") {
					t.Errorf("Expected candidate to contain 'Main content', got: %s", text)
				}
			},
		},
		{
			name: "candidate without parent returns unchanged",
			html: `<html><body>
				<div score="50">Standalone content</div>
			</body></html>`,
			description: "Should return candidate unchanged when it has no parent for sibling merging",
			verify: func(t *testing.T, candidate *goquery.Selection) {
				text := strings.TrimSpace(candidate.Text())
				if text != "Standalone content" {
					t.Errorf("Expected 'Standalone content', got: %s", text)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			candidate := FindTopCandidate(doc)
			tt.verify(t, candidate)
		})
	}
}

func TestFindTopCandidateEdgeCases(t *testing.T) {
	t.Run("empty document", func(t *testing.T) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(""))
		if err != nil {
			t.Fatalf("Failed to parse HTML: %v", err)
		}

		candidate := FindTopCandidate(doc)
		// Should fallback to body even for empty document (goquery creates structure)
		if candidate.Length() == 0 {
			t.Error("Expected fallback to body element, got empty selection")
		}
		actualTag := strings.ToLower(goquery.NodeName(candidate))
		if actualTag != "body" {
			t.Errorf("Expected fallback to body tag, got %s", actualTag)
		}
	})

	t.Run("malformed HTML", func(t *testing.T) {
		malformedHTML := `<html><body><div score="50">Unclosed div<p>Nested content</body></html>`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(malformedHTML))
		if err != nil {
			t.Fatalf("Failed to parse HTML: %v", err)
		}

		candidate := FindTopCandidate(doc)
		// Should still work with malformed HTML (goquery handles this)
		if candidate.Length() == 0 {
			t.Fatal("Expected candidate even with malformed HTML")
		}
	})

	t.Run("very large scores", func(t *testing.T) {
		html := `<html><body>
			<div score="999999">Very high score</div>
			<div score="1000000">Even higher score</div>
		</body></html>`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			t.Fatalf("Failed to parse HTML: %v", err)
		}

		candidate := FindTopCandidate(doc)
		text := strings.TrimSpace(candidate.Text())
		if text != "Even higher score" {
			t.Errorf("Expected 'Even higher score', got: %s", text)
		}
	})
}

func TestFindTopCandidateIntegration(t *testing.T) {
	// Test integration with actual scoring functions
	t.Run("integration with getScore function", func(t *testing.T) {
		html := `<html><body>
			<div data-content-score="75">Content with data-content-score</div>
			<div score="50">Content with score attribute</div>
		</body></html>`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			t.Fatalf("Failed to parse HTML: %v", err)
		}

		candidate := FindTopCandidate(doc)
		text := strings.TrimSpace(candidate.Text())
		if text != "Content with data-content-score" {
			t.Errorf("Expected to select element with higher data-content-score, got: %s", text)
		}
	})

	t.Run("non-candidate tags filtering", func(t *testing.T) {
		// Test all the non-candidate tags defined in NON_TOP_CANDIDATE_TAGS_RE
		nonCandidateTags := []string{"br", "b", "i", "label", "hr", "area", "base", "basefont", "input", "img", "link", "meta"}
		
		for _, tag := range nonCandidateTags {
			html := fmt.Sprintf(`<html><body>
				<%s score="100">Should be ignored</%s>
				<div score="10">Should be selected</div>
			</body></html>`, tag, tag)
			
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				t.Fatalf("Failed to parse HTML for tag %s: %v", tag, err)
			}

			candidate := FindTopCandidate(doc)
			actualTag := strings.ToLower(goquery.NodeName(candidate))
			if actualTag == tag {
				t.Errorf("Tag %s should be filtered out but was selected", tag)
			}
			if actualTag != "div" {
				t.Errorf("Expected div to be selected when %s is filtered, got %s", tag, actualTag)
			}
		}
	})
}

