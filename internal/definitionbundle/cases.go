package definitionbundle

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
)

// Suite contains definitions-owned assertions. It has no executable expressions.
type Suite struct {
	Protocol int    `json:"protocol"`
	Cases    []Case `json:"cases"`
}

// Case identifies every input needed to diagnose one extraction.
type Case struct {
	ID          string              `json:"id"`
	Definition  string              `json:"definition"`
	Fixture     string              `json:"fixture"`
	URL         string              `json:"url"`
	Synthetic   *bool               `json:"synthetic"`
	Expect      map[string]string   `json:"expect"`
	ExactlyOnce []string            `json:"exactly_once"`
	Exclude     []string            `json:"exclude"`
	Formats     []string            `json:"formats,omitempty"`
	Contains    map[string][]string `json:"contains,omitempty"`
}

// ParseSuite verifies all fixture references and demands coverage of every rule.
func ParseSuite(files map[string][]byte) (*Suite, error) {
	var suite Suite
	if err := DecodeJSON(files["conformance.json"], &suite); err != nil {
		return nil, err
	}
	if suite.Protocol != Protocol || len(suite.Cases) == 0 || len(suite.Cases) > 1024 {
		return nil, fmt.Errorf("invalid conformance protocol or case count")
	}
	seen, used := map[string]bool{}, map[string]bool{"conformance.json": true, "coverage.json": true}
	for i := range suite.Cases {
		c := &suite.Cases[i]
		if c.Formats == nil {
			c.Formats = []string{"html"}
		}
		if len(c.Formats) == 0 || len(c.Formats) > 3 || !slices.IsSorted(c.Formats) {
			return nil, fmt.Errorf("case %s: formats must be sorted, unique html/markdown/text", c.ID)
		}
		for index, format := range c.Formats {
			if !slices.Contains([]string{"html", "markdown", "text"}, format) ||
				(index > 0 && format == c.Formats[index-1]) {
				return nil, fmt.Errorf("case %s: unsupported or duplicate format %q", c.ID, format)
			}
		}
		if !identifier.MatchString(c.ID) || seen[c.ID] || c.Synthetic == nil {
			return nil, fmt.Errorf("invalid, duplicate or incomplete case %q", c.ID)
		}
		seen[c.ID] = true
		if !strings.HasPrefix(c.Definition, "definitions/") || files[c.Definition] == nil ||
			!strings.HasPrefix(c.Fixture, "fixtures/") || files[c.Fixture] == nil {
			return nil, fmt.Errorf("case %s: missing definition %s or fixture %s", c.ID, c.Definition, c.Fixture)
		}
		u, err := url.Parse(c.URL)
		if err != nil || u.Hostname() == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") ||
			len(c.URL) > 2048 || u.Fragment != "" {
			return nil, fmt.Errorf("case %s: invalid article URL", c.ID)
		}
		if c.Expect == nil || c.ExactlyOnce == nil || c.Exclude == nil ||
			len(c.Expect)+len(c.ExactlyOnce)+len(c.Contains) == 0 || len(c.ExactlyOnce)+len(c.Exclude) > 128 {
			return nil, fmt.Errorf("case %s: missing or excessive assertions", c.ID)
		}
		for field := range c.Expect {
			if !slices.Contains([]string{"title", "author", "date_published", "lead_image_url", "content", "url", "domain"}, field) {
				return nil, fmt.Errorf("case %s: unsupported expected field %q", c.ID, field)
			}
		}
		lists := [][]string{c.ExactlyOnce, c.Exclude}
		assertionCount := len(c.ExactlyOnce) + len(c.Exclude)
		if c.Contains != nil && len(c.Contains) == 0 {
			return nil, fmt.Errorf("case %s: contains must specify at least one format", c.ID)
		}
		for format, assertions := range c.Contains {
			if !slices.Contains(c.Formats, format) || len(assertions) == 0 {
				return nil, fmt.Errorf("case %s: contains requires a selected format and nonempty assertions", c.ID)
			}
			assertionCount += len(assertions)
			lists = append(lists, assertions)
		}
		if assertionCount > 128 {
			return nil, fmt.Errorf("case %s: assertion count exceeds 128", c.ID)
		}
		for _, format := range c.Formats {
			if len(c.Expect)+len(c.ExactlyOnce)+len(c.Contains[format]) == 0 {
				return nil, fmt.Errorf("case %s: %s requires a positive assertion", c.ID, format)
			}
		}
		for _, list := range lists {
			for _, assertion := range list {
				if assertion == "" || len(assertion) > 8192 {
					return nil, fmt.Errorf("case %s: invalid content assertion", c.ID)
				}
			}
		}
		used[c.Definition], used[c.Fixture] = true, true
	}
	for file := range files {
		if !used[file] {
			return nil, fmt.Errorf("payload %s has no conformance case", file)
		}
	}
	return &suite, nil
}

// Compare returns every mismatch rather than stopping at the first assertion.
func (c *Case) Compare(fields map[string]string, format string) []string {
	var failures []string
	keys := make([]string, 0, len(c.Expect))
	for field := range c.Expect {
		keys = append(keys, field)
	}
	slices.Sort(keys)
	for _, field := range keys {
		if fields[field] != c.Expect[field] {
			failures = append(failures, fmt.Sprintf("%s: got %q, want %q", field, fields[field], c.Expect[field]))
		}
	}
	for _, text := range c.ExactlyOnce {
		if count := strings.Count(fields["content"], text); count != 1 {
			failures = append(failures, fmt.Sprintf("content: %q occurs %d times, want 1", text, count))
		}
	}
	for _, text := range c.Exclude {
		if strings.Contains(fields["content"], text) {
			failures = append(failures, fmt.Sprintf("content: excluded text %q present", text))
		}
	}
	for _, text := range c.Contains[format] {
		if !strings.Contains(fields["content"], text) {
			failures = append(failures, fmt.Sprintf("content: required substring %q missing in %s", text, format))
		}
	}
	return failures
}
