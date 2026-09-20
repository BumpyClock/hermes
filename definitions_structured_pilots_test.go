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

func TestCanonicalDefinitionsStructuredPilotsOffline(t *testing.T) {
	root := os.Getenv("HERMES_DEFINITIONS_PILOT")
	if root == "" {
		t.Skip("set HERMES_DEFINITIONS_PILOT to run external canonical synthetic cases")
	}
	for _, pilot := range []struct {
		name, host, yamlHost string
	}{
		{name: "apartmenttherapy", host: "www.apartmenttherapy.com", yamlHost: "apartmenttherapy.com"},
		{name: "natgeo", host: "www.nationalgeographic.com", yamlHost: "nationalgeographic.com"},
		{name: "newsnatgeo", host: "news.nationalgeographic.com", yamlHost: "news.nationalgeographic.com"},
		{name: "abendblatt", host: "www.abendblatt.de", yamlHost: "abendblatt.de"},
	} {
		t.Run(pilot.name, func(t *testing.T) {
			//nolint:gosec // The operator selects this external synthetic-fixture repository.
			yaml, err := os.ReadFile(filepath.Join(root, "definitions", pilot.name+".yaml"))
			if err != nil {
				t.Fatal(err)
			}
			s, _ := localDefinitions(t, strings.Replace(string(yaml), "hosts: ["+pilot.yamlHost+"]", "hosts: [93.184.216.34]", 1))
			//nolint:gosec // The operator selects this external synthetic-fixture repository.
			data, err := os.ReadFile(filepath.Join(root, "fixtures", pilot.name, "cases.json"))
			if err != nil {
				t.Fatal(err)
			}
			var manifest struct {
				Cases []struct {
					File, URL, Title, Author string
					DatePublished            string `json:"date_published"`
					LeadImageURL             string `json:"lead_image_url"`
					ExactlyOnce              []string
					Contains                 []string
					Exclude                  []string
				}
			}
			if err := json.Unmarshal(data, &manifest); err != nil {
				t.Fatal(err)
			}
			remap := func(value string) string { return strings.ReplaceAll(value, pilot.host, "93.184.216.34") }
			for _, test := range manifest.Cases {
				t.Run(test.File, func(t *testing.T) {
					if filepath.Base(test.File) != test.File {
						t.Fatal("fixture path must be a basename")
					}
					//nolint:gosec // Fixture filename is constrained to a basename above.
					input, err := os.ReadFile(filepath.Join(root, "fixtures", pilot.name, test.File))
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
							if result.Title != test.Title || result.Author != test.Author || result.DatePublished == nil || result.DatePublished.Format("2006-01-02") != test.DatePublished || result.LeadImageURL != test.LeadImageURL {
								t.Fatalf("%s/%s metadata: %+v", op, format, result)
							}
							for _, value := range test.ExactlyOnce {
								if strings.Count(result.Content, value) != 1 {
									t.Errorf("%s/%s exactly_once %q: %s", op, format, value, result.Content)
								}
							}
							for _, value := range test.Exclude {
								if strings.Contains(result.Content, value) {
									t.Errorf("%s/%s excluded %q: %s", op, format, value, result.Content)
								}
							}
							if format != "text" {
								for _, value := range test.Contains {
									if !strings.Contains(result.Content, remap(value)) {
										t.Errorf("%s/%s missing %q: %s", op, format, value, result.Content)
									}
								}
							}
						}
					}
				})
			}
		})
	}
}
