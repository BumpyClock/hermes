package parser

import (
	"context"
	"fmt"
	"testing"
)

func TestGenericLeadImageScoresExtractedArticleImages(t *testing.T) {
	// The topper image outscores the article image if the whole page is scored.
	const page = `<html><head><title>Lead image scope</title></head><body>
		<div class="topper"><img src="https://example.com/uploads/topper.jpg" width="1200" height="800"></div>
		<article>
			<p>The first paragraph has enough words to make this article the best content candidate, with commas, clauses, and detail.</p>
			<img src="https://example.com/article.png" width="400" height="300">
			<p>The second paragraph adds more words so the article remains the best candidate, with commas, clauses, and detail.</p>
			<p>The third paragraph keeps the article long enough for the generic content extractor to accept it as sufficient.</p>
		</article>
	</body></html>`

	for _, contentType := range []string{"html", "markdown", "text"} {
		t.Run(contentType, func(t *testing.T) {
			result, err := New().ParseHTMLWithContext(context.Background(), page, "https://example.com/story",
				&ParserOptions{Fallback: true, ContentType: contentType})
			if err != nil {
				t.Fatal(err)
			}
			if want := "https://example.com/article.png"; result.LeadImageURL != want {
				t.Errorf("lead image = %q, want %q", result.LeadImageURL, want)
			}
		})
	}
}

func TestGenericLeadImagePrecedenceMatchesMercury(t *testing.T) {
	// Mercury checks meta tags, then scores article images, then falls back to
	// link[rel=image_src].
	const (
		ogImage  = `<meta property="og:image" content="https://example.com/og.png">`
		imageSrc = `<link rel="image_src" href="https://example.com/link.png">`
		article  = `<article>
			<p>The first paragraph has enough words to make this article the best content candidate, with commas, clauses, and detail.</p>
			%s
			<p>The second paragraph adds more words so the article remains the best candidate, with commas, clauses, and detail.</p>
			<p>The third paragraph keeps the article long enough for the generic content extractor to accept it as sufficient.</p>
		</article>`
		articleImage = `<img src="https://example.com/article.png" width="400" height="300">`
	)
	tests := []struct {
		name string
		head string
		body string
		want string
	}{
		{
			name: "meta tag beats article image",
			head: ogImage + imageSrc,
			body: fmt.Sprintf(article, articleImage),
			want: "https://example.com/og.png",
		},
		{
			name: "article image beats image_src link",
			head: imageSrc,
			body: fmt.Sprintf(article, articleImage),
			want: "https://example.com/article.png",
		},
		{
			name: "image_src link is the last resort",
			head: imageSrc,
			body: fmt.Sprintf(article, ""),
			want: "https://example.com/link.png",
		},
		{
			name: "meta tag without extracted content",
			head: ogImage,
			want: "https://example.com/og.png",
		},
		{
			name: "image_src link without extracted content",
			head: imageSrc,
			want: "https://example.com/link.png",
		},
	}

	for _, tt := range tests {
		page := "<html><head><title>Lead image precedence</title>" + tt.head + "</head><body>" + tt.body + "</body></html>"
		for _, contentType := range []string{"html", "markdown", "text"} {
			t.Run(tt.name+"/"+contentType, func(t *testing.T) {
				result, err := New().ParseHTMLWithContext(context.Background(), page, "https://example.com/story",
					&ParserOptions{Fallback: true, ContentType: contentType})
				if err != nil {
					t.Fatal(err)
				}
				if result.LeadImageURL != tt.want {
					t.Errorf("lead image = %q, want %q", result.LeadImageURL, tt.want)
				}
			})
		}
	}
}

func TestGenericLeadImagePenalizesNarrowImageAfterHeightCleanup(t *testing.T) {
	// CleanImages removes the narrow image's height. Mercury still applies the
	// width penalty, so the later photo wins: -19 against 20. Without the
	// penalty, the narrow image scores 31.
	const page = `<html><head><title>Narrow lead image</title></head><body>
		<article>
			<p>The first paragraph has enough words to make this article the best content candidate, with commas, clauses, and detail.</p>
			<img src="https://example.com/wp-content/narrow.jpg" width="40" height="400">
			<p>The second paragraph adds more words so the article remains the best candidate, with commas, clauses, and detail.</p>
			<img src="https://example.com/photo.png">
			<p>The third paragraph keeps the article long enough for the generic content extractor to accept it as sufficient.</p>
		</article>
	</body></html>`

	for _, contentType := range []string{"html", "markdown", "text"} {
		t.Run(contentType, func(t *testing.T) {
			result, err := New().ParseHTMLWithContext(context.Background(), page, "https://example.com/story",
				&ParserOptions{Fallback: true, ContentType: contentType})
			if err != nil {
				t.Fatal(err)
			}
			if want := "https://example.com/photo.png"; result.LeadImageURL != want {
				t.Errorf("lead image = %q, want %q", result.LeadImageURL, want)
			}
		})
	}
}
