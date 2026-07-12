package parser

import (
	"strings"
	"testing"
)

func TestPopularSiteExtractors(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		html        string
		wantTitle   string
		wantContent []string
		reject      []string
	}{
		{
			name: "ars technica current article boundary",
			url:  "https://arstechnica.com/science/2026/07/example/",
			html: `<article><header><h1>Current Ars headline</h1><p>header noise</p></header>
				<div class="post-content"><p>Opening Ars paragraph with enough meaningful article text.</p></div>
				<div class="post-content"><h2>Analysis</h2><p>Terminal Ars paragraph confirms every content group survives.</p></div></article>`,
			wantTitle:   "Current Ars headline",
			wantContent: []string{"Opening Ars paragraph", "Terminal Ars paragraph", "Analysis"},
			reject:      []string{"header noise"},
		},
		{
			name: "bbc semantic article boundary",
			url:  "https://www.bbc.com/news/articles/example",
			html: `<main><article><h1>BBC current headline</h1><div data-testid="byline"><span data-testid="byline-contributors">Ada Reporter</span><button>Share</button></div>
				<p>Opening BBC paragraph describes the actual reported event in detail.</p>
				<p>Final BBC paragraph closes the report with verified context.</p></article>
				<section data-component="recommendations"><p>Wrong related story card selected by generic scoring.</p></section></main>`,
			wantTitle:   "BBC current headline",
			wantContent: []string{"Opening BBC paragraph", "Final BBC paragraph"},
			reject:      []string{"Wrong related story", "Share"},
		},
		{
			name: "apple newsroom visible body",
			url:  "https://www.apple.com/newsroom/2026/07/example/",
			html: `<article class="article"><h1 class="hero-headline">Apple current announcement</h1>
				<div class="pagebody"><p>CUPERTINO, CALIFORNIA Apple announced its current manufacturing commitment.</p><p>Final substantive Apple paragraph appears exactly once.</p></div>
				<div class="docsanddownloads"><p>Text of this article duplicated for downloads.</p></div><div class="presscontacts">media@example.com</div></article>`,
			wantTitle:   "Apple current announcement",
			wantContent: []string{"CUPERTINO, CALIFORNIA", "Final substantive Apple paragraph"},
			reject:      []string{"Text of this article", "media@example.com"},
		},
		{
			name: "le monde current title and content",
			url:  "https://www.lemonde.fr/afrique/article/2026/07/10/example.html",
			html: `<article><h1>Le titre actuel du Monde</h1><a class="article__author-link">Marie Dupont</a>
				<div class="article__content"><p>Premier paragraphe visible de l’article du Monde.</p><p>Dernier paragraphe disponible avant le paywall.</p><figcaption>caption removed</figcaption></div></article>`,
			wantTitle:   "Le titre actuel du Monde",
			wantContent: []string{"Premier paragraphe", "Dernier paragraphe"},
			reject:      []string{"caption removed"},
		},
		{
			name: "times of india scoped body",
			url:  "https://timesofindia.indiatimes.com/india/example/articleshow/1.cms",
			html: `<div class="contentwrapper"><p>Trending navigation and puzzle noise.</p><section>lead media</section>
				<div data-articlebody="1"><section>lead media inside body</section><p>Opening Times of India report paragraph.</p><p>Final Times of India report paragraph.</p></div><p>Recommended story noise.</p></div><h1>Times current headline</h1>`,
			wantTitle:   "Times current headline",
			wantContent: []string{"Opening Times of India", "Final Times of India"},
			reject:      []string{"Trending navigation", "Recommended story", "lead media"},
		},
		{
			name: "washington post modern topper",
			url:  "https://www.washingtonpost.com/nation/2026/07/11/example/",
			html: `<meta name="author" content="Ada Post"><div id="topper-text-elems"><h1>Washington Post current headline</h1></div>
				<article><h1>Two ways to read this article</h1><div class="article-body"><p>Opening Washington Post paragraph.</p></div><div class="article-body"><p>Final Washington Post paragraph.</p></div></article>`,
			wantTitle:   "Washington Post current headline",
			wantContent: []string{"Opening Washington Post", "Final Washington Post"},
			reject:      []string{"Two ways to read"},
		},
		{
			name: "reuters current data testids",
			url:  "https://www.reuters.com/world/example/",
			html: `<h1 data-testid="Heading">Reuters current headline</h1><div data-testid="AuthorName">Ada Reuters</div>
				<div data-testid="ArticleBody"><p>Opening Reuters paragraph.</p><div data-testid="ContextWidget">duplicate summary</div><h2>What happens next</h2><p>Final Reuters paragraph.</p><div data-testid="promo-box">newsletter promo</div></div>`,
			wantTitle:   "Reuters current headline",
			wantContent: []string{"Opening Reuters", "What happens next", "Final Reuters"},
			reject:      []string{"duplicate summary", "newsletter promo"},
		},
		{
			name: "techcrunch wordpress article body",
			url:  "https://techcrunch.com/2026/07/11/example/",
			html: `<meta name="author" content="Ada Tech"><article><h1>TechCrunch current headline</h1>
				<div class="entry-content wp-block-post-content"><p>Opening TechCrunch article paragraph.</p><div class="wp-block-techcrunch-newsletter">newsletter promo</div><p>Final TechCrunch article paragraph.</p></div></article><p>Related TechCrunch story title.</p>`,
			wantTitle:   "TechCrunch current headline",
			wantContent: []string{"Opening TechCrunch", "Final TechCrunch"},
			reject:      []string{"newsletter promo", "Related TechCrunch"},
		},
		{
			name: "medium preserves linked prose",
			url:  "https://medium.com/@example/current-story",
			html: `<article><h1>Medium current headline</h1><div class="follow-card">4 followers</div><button>Follow</button>
				<p>Opening Medium paragraph includes <span><a href="/linked">meaningful linked words</a></span> in its prose.</p><p>Final Medium paragraph remains readable.</p></article>`,
			wantTitle:   "Medium current headline",
			wantContent: []string{"Opening Medium", "meaningful linked words", "Final Medium"},
			reject:      []string{"4 followers", "Follow"},
		},
		{
			name:        "guardian removes inline newsletter",
			url:         "https://www.theguardian.com/world/2026/jul/12/example",
			html:        `<h1>Guardian current headline</h1><div id="maincontent"><p>Opening Guardian report paragraph.</p><aside>Get a different world view newsletter</aside><p>Final Guardian report paragraph.</p></div>`,
			wantTitle:   "Guardian current headline",
			wantContent: []string{"Opening Guardian", "Final Guardian"},
			reject:      []string{"different world view"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := New().ParseHTML(tt.html, tt.url, &ParserOptions{Fallback: true, ContentType: "html"})
			if err != nil {
				t.Fatal(err)
			}
			if got := strings.TrimSpace(result.Title); got != tt.wantTitle {
				t.Fatalf("title %q does not equal %q", got, tt.wantTitle)
			}
			for _, want := range tt.wantContent {
				if !strings.Contains(result.Content, want) {
					t.Errorf("content missing %q: %q", want, result.Content)
				}
			}
			for _, reject := range tt.reject {
				if strings.Contains(result.Content, reject) {
					t.Errorf("content leaked %q: %q", reject, result.Content)
				}
			}
		})
	}
}

func TestBBCExtractorExcludesNestedRecommendationArticle(t *testing.T) {
	const mainParagraph = "The main BBC report contains the verified details readers need."
	const recommendationParagraph = "Nested recommendation article that must not appear in extracted content."

	result, err := New().ParseHTML(`<main><article><h1>BBC nested recommendation headline</h1>
		<p>`+mainParagraph+`</p>
		<section data-component="recommendations"><article><p>`+recommendationParagraph+`</p></article></section>
		<p>The final BBC report paragraph provides the concluding context.</p></article></main>`, "https://www.bbc.com/news/articles/example", &ParserOptions{Fallback: true, ContentType: "html"})
	if err != nil {
		t.Fatal(err)
	}

	if result.ExtractorUsed != "custom:www.bbc.com" {
		t.Fatalf("extractor used %q, want custom:www.bbc.com", result.ExtractorUsed)
	}
	if strings.Count(result.Content, mainParagraph) != 1 {
		t.Fatalf("main article content duplicated: %q", result.Content)
	}
	if strings.Contains(result.Content, recommendationParagraph) {
		t.Fatalf("nested recommendation leaked: %q", result.Content)
	}
}
