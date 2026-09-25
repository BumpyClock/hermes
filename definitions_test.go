package hermes

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestDefinitionsRelativeAttributeRules(t *testing.T) {
	for _, cleaner := range []string{"true", "false"} {
		t.Run(cleaner, func(t *testing.T) {
			s, _ := localDefinitions(t, `schema: 1
site: relative
hosts: [example.com]
metadata:
  title:
    - text: "a[href='/headline']"
  author:
    - text: "a[href='/author']"
  lead_image_url:
    - attribute: {selector: "img[src='photo.jpg']", name: src}
content:
  groups: [["article:has(a[href='/headline'])"]]
  remove: ['a[href^="/ads/"]']
  default_cleaner: `+cleaner+"\n")
			source := `<html><head><base target="_blank"><base href="../assets/"></head><body><article>
<a href="/headline">Selected relative headline</a>
<p>A substantive article paragraph supplies verified reporting and sufficient context for meaningful extraction.</p>
<a href="/ads/promotion">Excluded advertisement</a>
<img src="photo.jpg" width="640" height="480" alt="Article photograph">
<p>The concluding paragraph includes <a href="next">surviving relative link</a> and <a href="javascript:alert(1)">unsafe link</a>.</p>
<a href="/author">Ada Reporter</a><script>unsafe-script-marker</script></article></body></html>`
			for _, format := range []string{"html", "markdown", "text"} {
				r, err := New(WithDefinitions(s), WithContentType(format)).ParseHTML(context.Background(), source, "https://example.com/news/story")
				if err != nil {
					t.Fatal(err)
				}
				if r.Title != "Selected relative headline" || r.Author != "Ada Reporter" {
					t.Errorf("%s metadata lost: title=%q author=%q", format, r.Title, r.Author)
				}
				if r.LeadImageURL != "https://example.com/assets/photo.jpg" {
					t.Errorf("%s incorrect relative image metadata: %q", format, r.LeadImageURL)
				}
				if format != "text" && !strings.Contains(r.Content, "https://example.com/assets/photo.jpg") {
					t.Errorf("%s incorrect content image URL: %s", format, r.Content)
				}
				for _, unwanted := range []string{"Excluded advertisement", "javascript:", "unsafe-script-marker"} {
					if strings.Contains(r.Content, unwanted) {
						t.Errorf("%s leaked %q", format, unwanted)
					}
				}
				if !strings.Contains(r.Content, "surviving relative link") {
					t.Errorf("%s lost surviving content", format)
				}
				if format != "text" && !strings.Contains(r.Content, "https://example.com/assets/next") {
					t.Errorf("%s unresolved base link: %s", format, r.Content)
				}
			}
		})
	}
}

func TestDefinitionsGenericFallbackSanitization(t *testing.T) {
	s, _ := localDefinitions(t, `schema: 1
site: missing
hosts: [example.com]
content:
  groups: [[.missing]]
  default_cleaner: false
`)
	source := `<article><p>A substantive report preserves factual context and a <a href="javascript:alert(1)">readable unsafe link label</a>.</p><p>The conclusion retains enough article text to exercise generic fallback meaningfully.</p><script>unsafe-script-marker</script></article>`
	for _, configure := range [][]Option{nil, {WithDefinitions(nil)}, {WithDefinitions(s)}} {
		for _, format := range []string{"html", "markdown", "text"} {
			r, err := New(append(configure, WithContentType(format))...).ParseHTML(context.Background(), source, "https://example.com/story")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(r.Content, "substantive report") || strings.Contains(r.Content, "javascript:") || strings.Contains(r.Content, "unsafe-script-marker") {
				t.Errorf("%s unsafe fallback: %s", format, r.Content)
			}
		}
	}
}

func TestDefinitionsContentSourceOrder(t *testing.T) {
	s, _ := localDefinitions(t, `schema: 1
site: order
hosts: [example.com]
content:
  groups: [[.missing], [.second, .first, .first]]
  default_cleaner: false
`)
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<main><div class="first"><p>First selected article paragraph provides a substantive opening.</p></div><div class="second"><p>Second selected article paragraph provides a substantive ending.</p></div></main>`, "https://example.com/article")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(r.Content, "First selected") != 1 || strings.Count(r.Content, "Second selected") != 1 || strings.Index(r.Content, "First selected") > strings.Index(r.Content, "Second selected") {
		t.Fatalf("not a source-ordered deduplicated union: %s", r.Content)
	}
}

const bbcDefinition = `schema: 1
site: bbc
hosts: [bbc.com, bbc.co.uk]
metadata:
  title:
    - text: .missing
    - text: h1
  author:
    - text: '[data-testid="byline-contributors"]'
  date_published:
    - text: .invalid-date
    - attribute: {selector: time, name: datetime}
  lead_image_url:
    - attribute: {selector: 'meta[name="og:image"]', name: value}
content:
  groups:
    - [.missing]
    - [article, 'article p']
    - [main]
  remove:
    - '[data-component="recommendations"]'
    - '[data-testid="byline"]'
`

const bbcHTML = `<html><head><base href="/assets/"><meta property="og:image" content="https://www.bbc.com/photo.jpg"></head><body><main>
<article><h1>BBC current headline</h1><div data-testid="byline"><span data-testid="byline-contributors">Ada Reporter</span><button>Share</button></div>
<span class="invalid-date">not a date</span><time datetime="2026-07-01T12:00:00Z"></time>
<p>Opening BBC paragraph describes the actual reported event in detail.</p>
<section data-component="recommendations"><article><p>Nested recommendation article that must not appear in extracted content.</p></article></section>
<p>Final BBC paragraph closes the report with verified context. <a href="next">Further reporting</a></p></article>
<section data-component="recommendations"><p>Wrong related story card selected by generic scoring.</p></section></main></body></html>`

func localDefinitions(t *testing.T, source string) (*Definitions, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "site.yaml")
	//nolint:gosec // Only t.TempDir and a fixed filename form this test output path.
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := LoadDefinitions(dir)
	if err != nil {
		t.Fatal(err)
	}
	return s, path
}

func TestDefinitionsBBCSnapshotSharing(t *testing.T) {
	s, path := localDefinitions(t, bbcDefinition)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	clients := []*Client{New(WithDefinitions(s)), New(WithDefinitions(s))}
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r, err := clients[i%2].ParseHTML(context.Background(), bbcHTML, "https://www.bbc.com/news/articles/example")
			if err != nil {
				t.Error(err)
				return
			}
			if r.Title != "BBC current headline" || r.Author != "Ada Reporter" || r.DatePublished == nil || r.DatePublished.Format("2006-01-02") != "2026-07-01" || r.LeadImageURL != "https://www.bbc.com/photo.jpg" {
				t.Errorf("metadata: %+v", r)
			}
			for _, text := range []string{"Opening BBC paragraph", "Final BBC paragraph"} {
				if strings.Count(r.Content, text) != 1 {
					t.Errorf("boundary/dedup: %s", r.Content)
				}
			}
			for _, text := range []string{"Nested recommendation", "Wrong related story", "Share"} {
				if strings.Contains(r.Content, text) {
					t.Errorf("noise %q: %s", text, r.Content)
				}
			}
			if !strings.Contains(r.Content, `https://www.bbc.com/assets/next`) {
				t.Errorf("base URL: %s", r.Content)
			}
		}(i)
	}
	wg.Wait()
}

func TestDefinitionsCleanerAndSanitization(t *testing.T) {
	source := strings.ReplaceAll(bbcDefinition, "bbc.com, bbc.co.uk", "example.com")
	html := `<article><h1>Meaningful report</h1><p>This is the main paragraph with enough factual reporting to retain the article and test actual default cleaning.</p>
<h1>Optional sidebar explanation survives only without heuristic cleaning.</h1>
<div data-component="recommendations">Explicit removal always wins.</div>
<p>Another substantive paragraph closes the report and includes <a href="javascript:alert(1)">unsafe link</a> and <a href="next">safe link</a>.</p>
<script>alert("unsafe-script")</script><img src="/photo.jpg" onerror="alert(1)"></article>`
	for _, mode := range []string{"absent", "true", "false"} {
		t.Run(mode, func(t *testing.T) {
			input := source
			if mode != "absent" {
				input += "  default_cleaner: " + mode + "\n"
			}
			s, _ := localDefinitions(t, input)
			for _, format := range []string{"html", "markdown", "text"} {
				r, err := New(WithDefinitions(s), WithContentType(format)).ParseHTML(context.Background(), html, "https://example.com/report")
				if err != nil {
					t.Fatal(err)
				}
				kept := strings.Contains(r.Content, "Optional sidebar explanation")
				if kept != (mode == "false") {
					t.Errorf("%s cleaner %s: %s", format, mode, r.Content)
				}
				for _, bad := range []string{"Explicit removal", "unsafe-script", "javascript:", "onerror"} {
					if strings.Contains(r.Content, bad) {
						t.Errorf("%s leaked %s: %s", format, bad, r.Content)
					}
				}
				if !strings.Contains(r.Content, "main paragraph") {
					t.Errorf("lost article: %s", r.Content)
				}
				if format == "html" && !strings.Contains(r.Content, "https://example.com/next") {
					t.Errorf("lost URL: %s", r.Content)
				}
			}
		})
	}
}

type definitionTransport func(*http.Request) (*http.Response, error)

func (f definitionTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDefinitionsParseSafetyAndContext(t *testing.T) {
	s, _ := localDefinitions(t, bbcDefinition)
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "propagated")
	calls := 0
	client := New(WithDefinitions(s), WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Context().Value(key{}) != "propagated" {
			t.Error("request context missing")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"text/html"}}, Body: io.NopCloser(strings.NewReader(bbcHTML)), Request: r}, nil
	})))
	result, err := client.Parse(ctx, "https://www.bbc.com/news/articles/example")
	if err != nil || result.Title != "BBC current headline" || calls != 1 {
		t.Fatalf("parse: %+v %v calls=%d", result, err, calls)
	}
	for i, parse := range []func(context.Context, string) (*Result, error){client.Parse, func(ctx context.Context, u string) (*Result, error) { return client.ParseHTML(ctx, bbcHTML, u) }} {
		_, err := parse(ctx, "http://127.0.0.1/article")
		var pe *ParseError
		wantCode := ErrSSRF
		if i == 1 {
			wantCode = ErrInvalidURL
		}
		if !errors.As(err, &pe) || pe.Code != wantCode {
			t.Fatalf("expected URL denial: %v", err)
		}
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		_, err = parse(cancelled, "https://www.bbc.com/article")
		if err == nil {
			t.Fatal("cancelled parse succeeded")
		}
	}
	if calls != 1 {
		t.Fatalf("denied requests reached transport: %d", calls)
	}
}

func TestDefinitionsGenericIsolationAndOrderedGroups(t *testing.T) {
	// The legacy BBC selector chooses article; this snapshot deliberately chooses a different boundary.
	s, _ := localDefinitions(t, strings.Replace(bbcDefinition, "[article, 'article p']", "[.yaml-only]", 1))
	html := `<main><article><h1>Visible headline</h1><p>The compiled BBC article must not be selected when the local rule owns this hostname.</p></article>
<div class="yaml-only"><p>Local definition selects this meaningful paragraph exclusively.</p></div></main>`
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), html, "https://www.bbc.com/news/a")
	if err != nil || !strings.Contains(r.Content, "Local definition") || strings.Contains(r.Content, "compiled BBC") {
		t.Fatalf("%+v %v", r, err)
	}

	for _, snapshot := range []*Definitions{nil, {}, s} {
		client := New(WithDefinitions(snapshot))
		r, err := client.ParseHTML(context.Background(), `<article><h1>Generic title</h1><p>Generic fallback provides meaningful article content without any local matching definition.</p></article>`, "https://example.com/a")
		if err != nil || !strings.Contains(r.Content, "Generic fallback") {
			t.Fatalf("%+v %v", r, err)
		}
	}
}

func TestCanonicalDefinitionsBBC(t *testing.T) {
	root := os.Getenv("HERMES_DEFINITIONS_PILOT")
	if root == "" {
		t.Skip("set HERMES_DEFINITIONS_PILOT to run external canonical synthetic cases")
	}
	s, err := LoadDefinitions(filepath.Join(root, "definitions"))
	if err != nil {
		t.Fatal(err)
	}
	//nolint:gosec // The opt-in fixture repository is explicitly selected by the test operator.
	data, err := os.ReadFile(filepath.Join(root, "fixtures", "bbc", "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Cases []struct {
			File, URL, Title, Author string
			DatePublished            string   `json:"date_published"`
			LeadImageURL             string   `json:"lead_image_url"`
			ExactlyOnce              []string `json:"exactly_once"`
			Exclude                  []string
		}
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Cases) != 2 {
		t.Fatalf("expected two synthetic seed cases, got %d", len(manifest.Cases))
	}
	for _, c := range manifest.Cases {
		t.Run(c.File, func(t *testing.T) {
			//nolint:gosec // Case paths are local synthetic fixture inputs, not article data.
			html, err := os.ReadFile(filepath.Join(root, "fixtures", "bbc", c.File))
			if err != nil {
				t.Fatal(err)
			}
			r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), string(html), c.URL)
			if err != nil {
				t.Fatal(err)
			}
			if r.Title != c.Title || r.Author != c.Author || r.DatePublished == nil || r.DatePublished.Format("2006-01-02") != c.DatePublished || r.LeadImageURL != c.LeadImageURL {
				t.Fatalf("metadata: %+v", r)
			}
			for _, text := range c.ExactlyOnce {
				if strings.Count(r.Content, text) != 1 {
					t.Errorf("missing/duplicated %q: %s", text, r.Content)
				}
			}
			for _, text := range c.Exclude {
				if strings.Contains(r.Content, text) {
					t.Errorf("unexpected %q: %s", text, r.Content)
				}
			}
		})
	}
}

func TestDefinitionCapabilitiesIndependent(t *testing.T) {
	support := DefinitionCapabilities()
	if support.Schema != 1 || support.MaxFileBytes != 262144 {
		t.Fatalf("unexpected capabilities: %+v", support)
	}
	for _, capability := range []string{"metadata.text", "metadata.attribute", "content.groups", "content.remove", "content.default_cleaner", "hosts.exact-www", "hosts.wildcard"} {
		if !slices.Contains(support.Capabilities, capability) {
			t.Fatalf("missing existing capability %q", capability)
		}
	}
	support.Capabilities[0] = "unsupported"
	if DefinitionCapabilities().Capabilities[0] == "unsupported" {
		t.Fatal("capability storage is shared")
	}
}

func TestDefinitionsUsedCapabilitiesAndSite(t *testing.T) {
	snapshot, _ := localDefinitions(t, `schema: 1
site: example
hosts: [www.example.com, "*.example.org"]
metadata:
  title: [{text: h1}]
content:
  groups: [[article]]
  transforms:
    - target: descendants
      selector: p
      algorithm.apply: {name: abendblatt.deobfuscate}
`)
	capabilities, algorithms := snapshot.UsedCapabilities()
	for _, capability := range []string{"content.groups", "hosts.exact-www", "hosts.wildcard", "metadata.text", "transform.algorithm.apply"} {
		if !slices.Contains(capabilities, capability) {
			t.Fatalf("UsedCapabilities() = %v, missing %q", capabilities, capability)
		}
	}
	if slices.Contains(capabilities, "content.remove") || !slices.Equal(algorithms, []string{"abendblatt.deobfuscate"}) {
		t.Fatalf("UsedCapabilities() = %v, %v", capabilities, algorithms)
	}
	support := DefinitionCapabilities()
	for _, capability := range capabilities {
		if !slices.Contains(support.Capabilities, capability) {
			t.Fatalf("used capability %q is outside DefinitionCapabilities", capability)
		}
	}
	capabilities[0] = "mutated"
	if again, _ := snapshot.UsedCapabilities(); again[0] == "mutated" {
		t.Fatal("capability storage is shared")
	}

	for host, want := range map[string]string{"example.com": "example", "WWW.Example.com.": "example", "news.example.org": "example", "example.org": "", "other.test": ""} {
		site, ok := snapshot.Site(host)
		if site != want || ok != (want != "") {
			t.Fatalf("Site(%q) = %q, %t; want %q", host, site, ok, want)
		}
	}
	for _, empty := range []*Definitions{nil, {}} {
		if capabilities, algorithms := empty.UsedCapabilities(); len(capabilities)+len(algorithms) != 0 {
			t.Fatalf("empty snapshot uses %v %v", capabilities, algorithms)
		}
		if site, ok := empty.Site("example.com"); ok || site != "" {
			t.Fatalf("empty snapshot selected %q", site)
		}
	}
}
