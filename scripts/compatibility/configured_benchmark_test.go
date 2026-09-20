//go:build compatibility_configured

// The runner pairs this workload with a version-specific adapter in each source archive.
package compatibility_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/BumpyClock/hermes/internal/parser"
)

var configuredFixtures = []struct{ name, host string }{
	{"nytimes.html", "www.nytimes.com"},
	{"arstechnica.html", "arstechnica.com"},
}

func configuredHTML(tb testing.TB, name string) string {
	tb.Helper()
	//nolint:gosec // The offline runner selects the local immutable fixture directory.
	body, err := os.ReadFile(filepath.Join(os.Getenv("HERMES_COMPAT_FIXTURES"), name))
	if err != nil {
		tb.Fatal(err)
	}
	return string(body)
}

func TestConfiguredEvidence(t *testing.T) {
	options := configuredOptions(t)
	engine := parser.New(&options)
	for _, fixture := range configuredFixtures {
		input := configuredHTML(t, fixture.name)
		for _, format := range []string{"html", "markdown", "text"} {
			options.ContentType = format
			result, err := engine.ParseHTMLWithContext(context.Background(), input, "https://"+fixture.host+"/article", &options)
			if err != nil {
				t.Fatal(err)
			}
			if result.Title == "" || result.Content == "" || result.ExtractorUsed != configuredExtractor(fixture.host) {
				t.Fatalf("nonempty configured extraction required: %+v", result)
			}
			row := struct {
				ID     string         `json:"id"`
				Result *parser.Result `json:"result"`
			}{fixture.name + "/" + format, result}
			data, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("OBSERVATION %s", data)
		}
	}
}

func BenchmarkConfigured(b *testing.B) {
	options := configuredOptions(b)
	for _, fixture := range configuredFixtures {
		input := configuredHTML(b, fixture.name)
		for _, format := range []string{"html", "markdown", "text"} {
			b.Run(fixture.name+"/"+format, func(b *testing.B) {
				opts := options
				opts.ContentType = format
				engine := parser.New(&opts)
				url := "https://" + fixture.host + "/article"
				b.ReportAllocs()
				b.SetBytes(int64(len(input)))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, err := engine.ParseHTMLWithContext(context.Background(), input, url, &opts)
					if err != nil || result.Content == "" {
						b.Fatalf("configured extraction failed: result=%+v error=%v", result, err)
					}
				}
			})
		}
	}
}
