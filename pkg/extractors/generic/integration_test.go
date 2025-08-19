// ABOUTME: Integration tests for ExtractBestNode with real-world scenarios
// ABOUTME: Verifies end-to-end functionality with complex HTML structures

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// TestExtractBestNode_RealWorldArticle tests with realistic article HTML
func TestExtractBestNode_RealWorldArticle(t *testing.T) {
	// Simulates a typical news article layout
	html := `<html><head>
		<title>Test Article</title>
	</head><body>
		<header class="site-header">
			<nav>Navigation links</nav>
		</header>
		<main class="main-content">
			<article class="article-body">
				<h1>Article Headline That Should Be Extracted</h1>
				<div class="article-meta">
					<span class="author">By John Smith</span>
					<time datetime="2023-01-01">January 1, 2023</time>
				</div>
				<div class="article-content">
					<p>This is the first paragraph of the article content. It contains substantial information and should be part of the extracted content.</p>
					<p>This is the second paragraph with even more detailed content. It includes important information that readers need to know about the topic.</p>
					<p>The third paragraph continues the article with additional context and details that are relevant to the main story.</p>
				</div>
			</article>
		</main>
		<aside class="sidebar">
			<div class="widget advertisement">
				<p>This is an advertisement that should not be selected.</p>
			</div>
			<div class="widget related-posts">
				<h3>Related Articles</h3>
				<ul>
					<li><a href="#">Related article 1</a></li>
					<li><a href="#">Related article 2</a></li>
				</ul>
			</div>
		</aside>
		<footer class="site-footer">
			<p>Copyright information</p>
		</footer>
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
	if !strings.Contains(text, "first paragraph of the article") {
		t.Errorf("Expected main article content, got: %s", text)
	}
	
	if !strings.Contains(text, "second paragraph with even more detailed content") {
		t.Errorf("Expected main article content, got: %s", text)
	}
	
	// Should not contain sidebar/footer content
	if strings.Contains(text, "advertisement that should not be selected") {
		t.Errorf("Expected sidebar content to be excluded, but found: %s", text)
	}
	
	if strings.Contains(text, "Copyright information") {
		t.Errorf("Expected footer content to be excluded, but found: %s", text)
	}

	t.Logf("Extracted content length: %d characters", len(text))
	t.Logf("First 200 chars: %s", text[:min(200, len(text))])
}

// TestExtractBestNode_BlogPost tests with blog-style content
func TestExtractBestNode_BlogPost(t *testing.T) {
	html := `<html><body>
		<div class="container">
			<div class="post-content">
				<h2>Blog Post Title</h2>
				<div class="post-body">
					<p>Welcome to this blog post about content extraction. This paragraph introduces the topic and provides context for readers.</p>
					<p>The main content of this blog post discusses various techniques and approaches. It's structured to provide valuable information to readers who are interested in this topic.</p>
					<p>This final paragraph summarizes the key points and provides a conclusion to the blog post. It ties together the main themes discussed throughout the article.</p>
				</div>
			</div>
			<div class="comments">
				<h3>Comments</h3>
				<div class="comment">
					<p>This is a comment that should be filtered out.</p>
				</div>
			</div>
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
	
	// Should contain blog post content
	if !strings.Contains(text, "content extraction") {
		t.Errorf("Expected blog content, got: %s", text)
	}
	
	if !strings.Contains(text, "various techniques and approaches") {
		t.Errorf("Expected blog content, got: %s", text)
	}
	
	// Should have reasonable length
	if len(text) < 100 {
		t.Errorf("Expected substantial content, but got only %d characters: %s", len(text), text)
	}

	t.Logf("Blog content length: %d characters", len(text))
}

// TestExtractBestNode_Scoring verifies the scoring system integration
func TestExtractBestNode_Scoring(t *testing.T) {
	html := `<html><body>
		<div class="low-quality">
			<p>Short text.</p>
		</div>
		<div class="article-body main-content">
			<p>This is a much longer paragraph with substantial content that should score higher. It contains multiple sentences, commas, and provides detailed information about the topic being discussed.</p>
			<p>Another substantial paragraph that continues the discussion with additional details, examples, and context that would be valuable to readers interested in this subject matter.</p>
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
	
	// Should select the longer, higher-quality content
	if !strings.Contains(text, "substantial content that should score higher") {
		t.Errorf("Expected high-scoring content to be selected, got: %s", text)
	}
	
	// Should not select the short content
	if strings.Contains(text, "Short text.") && !strings.Contains(text, "substantial content") {
		t.Errorf("Expected low-scoring content to be rejected, but it was selected: %s", text)
	}

	t.Logf("Selected content length: %d characters", len(text))
}

// Helper function for Go versions that don't have min in stdlib
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestExtractBestNode_CompareOptions tests different option combinations
func TestExtractBestNode_CompareOptions(t *testing.T) {
	html := `<html><body>
		<div class="content-wrapper">
			<div class="article-content main-story">
				<p>This is the main article content with substantial text and good class names.</p>
			</div>
			<div class="comment-section">
				<p>This is a comment section that might be stripped.</p>
			</div>
		</div>
	</body></html>`

	doc1, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	doc2, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	// Test with stripping enabled
	optsWithStripping := ExtractBestNodeOptions{
		StripUnlikelyCandidates: true,
		WeightNodes:             true,
	}

	candidateWithStripping := ExtractBestNode(doc1, optsWithStripping)

	// Test with stripping disabled  
	optsNoStripping := ExtractBestNodeOptions{
		StripUnlikelyCandidates: false,
		WeightNodes:             true,
	}

	candidateNoStripping := ExtractBestNode(doc2, optsNoStripping)

	// Both should find candidates
	if candidateWithStripping == nil {
		t.Error("Expected candidate with stripping enabled")
	}
	
	if candidateNoStripping == nil {
		t.Error("Expected candidate with stripping disabled")
	}

	if candidateWithStripping != nil {
		text1 := strings.TrimSpace(candidateWithStripping.Text())
		t.Logf("With stripping: %s", text1)
	}

	if candidateNoStripping != nil {
		text2 := strings.TrimSpace(candidateNoStripping.Text())  
		t.Logf("Without stripping: %s", text2)
	}
}