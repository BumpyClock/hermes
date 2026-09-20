package hermes

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCanonicalDefinitionsNYTimesOffline(t *testing.T) {
	root := os.Getenv("HERMES_DEFINITIONS_PILOT")
	if root == "" {
		t.Skip("set HERMES_DEFINITIONS_PILOT to run external canonical synthetic cases")
	}
	snapshot, err := LoadDefinitions(filepath.Join(root, "definitions"))
	if err != nil {
		t.Fatal(err)
	}
	if rule := snapshot.snapshot.Match("www.nytimes.com"); rule == nil || rule.Domain != "nytimes" {
		t.Fatal("canonical hostname did not select NYTimes")
	}
	// The public client resolves hostnames even for ParseHTML. Remap only the
	// declared host and expected URLs to an IP; selectors and operations stay canonical.
	//nolint:gosec // The operator selects this external synthetic-fixture repository.
	yaml, err := os.ReadFile(filepath.Join(root, "definitions", "nytimes.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(yaml), "hosts: [nytimes.com]") != 1 {
		t.Fatal("update the explicit offline hostname mapping for the changed pilot")
	}
	s, _ := localDefinitions(t, strings.Replace(string(yaml), "hosts: [nytimes.com]", "hosts: [93.184.216.34]", 1))
	//nolint:gosec // The operator selects this external synthetic-fixture repository.
	data, err := os.ReadFile(filepath.Join(root, "fixtures", "nytimes", "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Cases []struct {
			File, URL, Title, Author string
			DatePublished            string   `json:"date_published"`
			LeadImageURL             string   `json:"lead_image_url"`
			ExactlyOnce              []string `json:"exactly_once"`
			Contains, Exclude        []string
		}
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Cases) != 2 {
		t.Fatalf("expected two independent NYTimes cases, got %d", len(manifest.Cases))
	}
	remap := func(s string) string { return strings.ReplaceAll(s, "www.nytimes.com", "93.184.216.34") }
	for _, test := range manifest.Cases {
		t.Run(test.File, func(t *testing.T) {
			if filepath.Base(test.File) != test.File {
				t.Fatal("fixture path must be a basename")
			}
			//nolint:gosec // Explicit external fixture repository; manifest filename is checked above.
			input, err := os.ReadFile(filepath.Join(root, "fixtures", "nytimes", test.File))
			if err != nil {
				t.Fatal(err)
			}
			for _, format := range []string{"html", "markdown", "text"} {
				client := New(WithDefinitions(s), WithContentType(format), WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(string(input))), Request: r}, nil
				})))
				for _, op := range []string{"Parse", "ParseHTML"} {
					var result *Result
					if op == "Parse" {
						result, err = client.Parse(context.Background(), remap(test.URL))
					} else {
						result, err = client.ParseHTML(context.Background(), string(input), remap(test.URL))
					}
					if err != nil {
						t.Fatal(err)
					}
					if result.Title != test.Title || result.Author != test.Author || result.DatePublished == nil || result.DatePublished.Format("2006-01-02") != test.DatePublished || result.LeadImageURL != remap(test.LeadImageURL) {
						t.Fatalf("%s/%s metadata: %+v", op, format, result)
					}
					for _, text := range test.ExactlyOnce {
						if strings.Count(result.Content, text) != 1 {
							t.Errorf("%s/%s missing/duplicated %q: %s", op, format, text, result.Content)
						}
					}
					for _, text := range test.Exclude {
						if strings.Contains(result.Content, text) {
							t.Errorf("%s/%s unwanted %q: %s", op, format, text, result.Content)
						}
					}
					if format != "text" {
						for _, value := range test.Contains {
							if !strings.Contains(result.Content, remap(value)) {
								t.Errorf("%s/%s missing repaired URL %q: %s", op, format, value, result.Content)
							}
						}
					}
				}
			}
		})
	}
}
