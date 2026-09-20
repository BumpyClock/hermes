package compatibility_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	hermes "github.com/BumpyClock/hermes"
)

func BenchmarkGeneric(b *testing.B) {
	for _, name := range []string{"article.html", "nytimes.html", "arstechnica.html"} {
		body, err := os.ReadFile(filepath.Join(os.Getenv("HERMES_COMPAT_FIXTURES"), name)) //nolint:gosec // The compatibility runner selects a local fixture directory.
		if err != nil {
			b.Fatal(err)
		}
		input := string(body)
		for _, format := range []string{"html", "markdown", "text"} {
			b.Run(name+"/"+format, func(b *testing.B) {
				client := hermes.New(hermes.WithContentType(format))
				b.ReportAllocs()
				b.SetBytes(int64(len(body)))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, err := client.ParseHTML(context.Background(), input, "https://93.184.216.34/article")
					if err != nil || result.IsEmpty() {
						b.Fatalf("ParseHTML: result=%+v error=%v", result, err)
					}
				}
			})
		}
	}
}
