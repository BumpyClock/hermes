// ABOUTME: Test suite for ExtractBestNode function - main content extraction orchestrator
// ABOUTME: Tests integration of stripping, paragraph conversion, scoring, and candidate selection

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// Test imports

// Test basic functionality with simple HTML
func TestExtractBestNode_BasicFunctionality(t *testing.T) {
	html := `<html><body>
		<div class="content">
			<p>This is the main article content with multiple sentences.</p>
			<p>This paragraph has more content and is likely the best candidate.</p>
		</div>
		<div class="sidebar">
			<p>This is sidebar content.</p>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	opts := ExtractBestNodeOptions{
		StripUnlikelyCandidates: false,
		WeightNodes:             true,
	}

	candidate := ExtractBestNode(doc, opts)
	
	if candidate == nil {
		t.Fatal("Expected a candidate, got nil")
	}

	// Should contain the main content
	text := strings.TrimSpace(candidate.Text())
	if !strings.Contains(text, "main article content") {
		t.Errorf("Expected candidate to contain main article content, got: %s", text)
	}
}

// Test with stripUnlikelyCandidates option enabled
func TestExtractBestNode_StripUnlikelyCandidates(t *testing.T) {
	html := `<html><body>
		<div class="article-content">
			<p>This is the main article content that should be selected.</p>
			<p>More important article content goes here with detailed information.</p>
		</div>
		<div class="comment">
			<p>This is a comment that should be stripped.</p>
		</div>
		<div class="advertisement">
			<p>This is an ad that should be removed.</p>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	opts := ExtractBestNodeOptions{
		StripUnlikelyCandidates: true,
		WeightNodes:             true,
	}

	candidate := ExtractBestNode(doc, opts)
	
	if candidate == nil {
		t.Fatal("Expected a candidate, got nil")
	}

	text := strings.TrimSpace(candidate.Text())
	
	// Should contain main content
	if !strings.Contains(text, "main article content") {
		t.Errorf("Expected candidate to contain main article content, got: %s", text)
	}
	
	// Should not contain stripped content
	if strings.Contains(text, "comment that should be stripped") {
		t.Errorf("Expected stripped content to be removed, but found in: %s", text)
	}
	
	if strings.Contains(text, "ad that should be removed") {
		t.Errorf("Expected ad content to be removed, but found in: %s", text)
	}
}

// Test with stripUnlikelyCandidates disabled
func TestExtractBestNode_NoStripUnlikelyCandidates(t *testing.T) {
	html := `<html><body>
		<div class="content">
			<p>Main content paragraph here.</p>
		</div>
		<div class="comment">
			<p>Comment content here.</p>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	opts := ExtractBestNodeOptions{
		StripUnlikelyCandidates: false,
		WeightNodes:             true,
	}

	candidate := ExtractBestNode(doc, opts)
	
	if candidate == nil {
		t.Fatal("Expected a candidate, got nil")
	}

	// With stripping disabled, the scoring system should still work
	text := strings.TrimSpace(candidate.Text())
	if len(text) == 0 {
		t.Error("Expected candidate to have some content")
	}
}

// Test paragraph conversion functionality
func TestExtractBestNode_ParagraphConversion(t *testing.T) {
	html := `<html><body>
		<div class="article">
			<div>This div should be converted to paragraph.<br><br>After double BR.</div>
			<span>This span should be converted to paragraph.</span>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	opts := ExtractBestNodeOptions{
		StripUnlikelyCandidates: false,
		WeightNodes:             false,
	}

	candidate := ExtractBestNode(doc, opts)
	
	if candidate == nil {
		t.Fatal("Expected a candidate, got nil")
	}

	// Should have processed paragraph conversion
	text := strings.TrimSpace(candidate.Text())
	if !strings.Contains(text, "converted to paragraph") {
		t.Errorf("Expected converted content, got: %s", text)
	}
}

// Test scoring integration
func TestExtractBestNode_ScoringIntegration(t *testing.T) {
	html := `<html><body>
		<div class="article-body" id="main-content">
			<p>This is a long article paragraph with substantial content. It contains multiple sentences and should score highly.</p>
			<p>Another substantial paragraph with more content. This should also contribute to the overall score.</p>
		</div>
		<div class="sidebar">
			<p>Short sidebar text.</p>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	opts := ExtractBestNodeOptions{
		StripUnlikelyCandidates: false,
		WeightNodes:             true,
	}

	candidate := ExtractBestNode(doc, opts)
	
	if candidate == nil {
		t.Fatal("Expected a candidate, got nil")
	}

	text := strings.TrimSpace(candidate.Text())
	
	// Should select the content with higher score (longer, better class names)
	if !strings.Contains(text, "substantial content") {
		t.Errorf("Expected high-scoring content to be selected, got: %s", text)
	}
}

// Test with no content (edge case)
func TestExtractBestNode_NoContent(t *testing.T) {
	html := `<html><body></body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	opts := ExtractBestNodeOptions{
		StripUnlikelyCandidates: false,
		WeightNodes:             true,
	}

	candidate := ExtractBestNode(doc, opts)
	
	// Should handle empty content gracefully - may return body or nil
	if candidate != nil {
		text := strings.TrimSpace(candidate.Text())
		if len(text) > 0 {
			t.Errorf("Expected empty content, got: %s", text)
		}
	}
}

// Test with malformed HTML
func TestExtractBestNode_MalformedHTML(t *testing.T) {
	html := `<html><body>
		<div class="content">
			<p>Content with <unclosed tag
			<p>Another paragraph</p>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	opts := ExtractBestNodeOptions{
		StripUnlikelyCandidates: false,
		WeightNodes:             true,
	}

	// Should not panic with malformed HTML
	candidate := ExtractBestNode(doc, opts)
	
	if candidate == nil {
		t.Fatal("Expected a candidate even with malformed HTML")
	}
}

// Test integration with all options enabled
func TestExtractBestNode_AllOptionsEnabled(t *testing.T) {
	html := `<html><body>
		<div class="article-content main-content">
			<p>This is the primary article content with multiple sentences and substantial length.</p>
			<p>Additional article content that should be part of the main selection.</p>
		</div>
		<div class="comment-section">
			<p>This is a comment that should be filtered out.</p>
		</div>
		<div class="advertisement sidebar">
			<p>Ad content to be removed.</p>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	opts := ExtractBestNodeOptions{
		StripUnlikelyCandidates: true,
		WeightNodes:             true,
	}

	candidate := ExtractBestNode(doc, opts)
	
	if candidate == nil {
		t.Fatal("Expected a candidate, got nil")
	}

	text := strings.TrimSpace(candidate.Text())
	
	// Should contain main article content
	if !strings.Contains(text, "primary article content") {
		t.Errorf("Expected main content to be selected, got: %s", text)
	}
	
	// Should not contain filtered content
	if strings.Contains(text, "comment that should be filtered") {
		t.Errorf("Expected comments to be filtered out, but found in: %s", text)
	}
	
	if strings.Contains(text, "Ad content to be removed") {
		t.Errorf("Expected ads to be filtered out, but found in: %s", text)
	}
}

// Test weight nodes option
func TestExtractBestNode_WeightNodesOption(t *testing.T) {
	html := `<html><body>
		<div class="article-body">
			<p>Article content with good class name.</p>
		</div>
		<div class="random-div">
			<p>Content with neutral class name.</p>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// Test with weight nodes enabled
	optsWithWeights := ExtractBestNodeOptions{
		StripUnlikelyCandidates: false,
		WeightNodes:             true,
	}

	candidateWithWeights := ExtractBestNode(doc, optsWithWeights)
	
	if candidateWithWeights == nil {
		t.Fatal("Expected a candidate with weights enabled")
	}

	// Test with weight nodes disabled
	optsNoWeights := ExtractBestNodeOptions{
		StripUnlikelyCandidates: false,
		WeightNodes:             false,
	}

	candidateNoWeights := ExtractBestNode(doc, optsNoWeights)
	
	if candidateNoWeights == nil {
		t.Fatal("Expected a candidate with weights disabled")
	}

	// Both should return valid candidates, but scoring may differ
	// The key is that the function doesn't panic with either setting
}