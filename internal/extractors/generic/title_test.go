package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
)

func TestCleanTitle_PreservesExtractorOrder(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body></body></html>"))
	require.NoError(t, err)

	for _, tc := range []struct {
		name  string
		title string
		url   string
		want  string
	}{
		{
			name:  "split before tag removal",
			title: `<span class="large">Example News</span> - Actual Article`,
			url:   "https://examplenews.com",
			want:  "Example News - Actual Article",
		},
		{
			name:  "remove all domain spaces",
			title: "e x a m p l e n e w s - Actual Article",
			url:   "https://example.com",
			want:  "Actual Article",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, cleanTitle(tc.title, tc.url, doc.Selection))
		})
	}
}
