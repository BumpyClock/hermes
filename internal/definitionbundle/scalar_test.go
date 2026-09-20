package definitionbundle_test

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	bundle "github.com/BumpyClock/hermes/internal/definitionbundle"

	hermes "github.com/BumpyClock/hermes"
)

func TestScalarRequirementsUseValidatedOperationAndConditionKeys(t *testing.T) {
	support := hermes.DefinitionCapabilities()
	if !slices.Contains(support.Capabilities, "transform.attribute.copy") {
		t.Skip("selected engine does not implement scalar transforms")
	}
	m, files := pinned(t)
	definition := "definitions/example.yaml"
	files[definition] = []byte(`schema: 1
site: synthetic-example
hosts: [example.com]
content:
  groups: [[article]]
  transforms:
    - target: root
      when:
        - exists: {attribute: id}
        - text: {comparison: contains, value: report}
        - number: {attribute: width, comparison: ge, value: 1}
        - descendant: {selector: p}
      attribute.copy: {from: id, to: title}
    - target: root
      attribute.set: {name: title, value: report}
    - target: root
      attribute.remove: {name: title}
    - target: root
      string.replace: {attribute: title, old: a, new: b}
    - target: root
      regex.capture: {attribute: title, pattern: '(report)', group: 1}
    - target: root
      url.resolve: {attribute: href}
    - target: root
      url.build: {attribute: href, base: 'https://example.com/'}
    - target: root
      element.rename: {tag: section}
    - target: descendants
      selector: span
      element.unwrap: {}
    - target: descendants
      selector: aside
      element.remove: {}
`)
	suite, err := bundle.ParseSuite(files)
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(t.TempDir(), "definitions")
	if err = bundle.WriteDefinitions(files, directory); err != nil {
		t.Fatal(err)
	}
	m.Engine.Operations = slices.Clone(support.Capabilities)
	slices.Sort(m.Engine.Operations)
	if err = m.AuditRequirements(files, suite, directory); err != nil {
		t.Fatalf("all implemented scalar operations and conditions should be auditable: %v", err)
	}
	for _, capability := range []string{
		"transform.attribute.copy", "transform.attribute.set", "transform.attribute.remove",
		"transform.string.replace", "transform.regex.capture", "transform.url.resolve",
		"transform.url.build", "transform.element.rename", "transform.element.unwrap",
		"transform.element.remove", "condition.exists", "condition.text",
		"condition.number", "condition.descendant",
	} {
		t.Run(capability, func(t *testing.T) {
			original := m.Engine.Operations
			m.Engine.Operations = slices.DeleteFunc(slices.Clone(original), func(name string) bool { return name == capability })
			if err := m.AuditRequirements(files, suite, directory); err == nil || !strings.Contains(err.Error(), capability) {
				t.Fatalf("missing scalar declaration not identified: %v", err)
			}
			m.Engine.Operations = original
		})
	}
}

func TestSuiteAcceptsExplicitFormatScopedContains(t *testing.T) {
	_, files := pinned(t)
	var suite map[string]any
	if err := json.Unmarshal(files["conformance.json"], &suite); err != nil {
		t.Fatal(err)
	}
	c := suite["cases"].([]any)[0].(map[string]any)
	c["formats"] = []string{"html", "markdown"}
	c["contains"] = map[string][]string{"html": {"https://example.com/image.jpg"}, "markdown": {"https://example.com/image.jpg"}}
	c["expect"], c["exactly_once"] = map[string]string{}, []string{}
	data, err := json.Marshal(suite)
	if err != nil {
		t.Fatal(err)
	}
	files["conformance.json"] = data
	if _, err = bundle.ParseSuite(files); err != nil {
		t.Fatalf("format-scoped substring assertions should load: %v", err)
	}
}

func TestContainsAssertionsRespectOutputFormat(t *testing.T) {
	c := bundle.Case{
		Expect:      map[string]string{"title": "Expected"},
		ExactlyOnce: []string{"Report prose"}, Exclude: []string{"advertisement"},
		Contains: map[string][]string{"html": {"https://example.com/image.jpg"}, "markdown": {"https://example.com/image.jpg"}},
	}
	fields := map[string]string{"title": "Expected", "content": "Report prose"}
	if failures := c.Compare(fields, "text"); len(failures) != 0 {
		t.Fatalf("URL assertions must not falsely apply to text: %v", failures)
	}
	for _, format := range []string{"html", "markdown"} {
		if failures := c.Compare(fields, format); len(failures) != 1 || !strings.Contains(failures[0], "required substring") {
			t.Fatalf("missing URL substring not identified in %s: %v", format, failures)
		}
	}
	fields["content"] += " https://example.com/image.jpg https://example.com/image.jpg"
	if failures := c.Compare(fields, "html"); len(failures) != 0 {
		t.Fatalf("contains means presence, not exactly-once: %v", failures)
	}
}

func TestFormatContractRejectsMalformedAssertions(t *testing.T) {
	for _, extra := range []string{
		`"formats":[]`, `"formats":["HTML"]`, `"formats":["html","html"]`, `"formats":["text","html"]`,
		`"formats":["html"],"contains":{"markdown":["fragment"]}`,
		`"formats":["html"],"contains":{"HTML":["fragment"]}`,
		`"formats":["html"],"contains":{"html":[]}`,
		`"formats":["html"],"contains":{"html":[true]}`,
		`"formats":["html"],"Contains":{"html":["fragment"]}`,
	} {
		_, files := pinned(t)
		files["conformance.json"] = []byte(strings.Replace(string(files["conformance.json"]), `"id":`, extra+`,"id":`, 1))
		if _, err := bundle.ParseSuite(files); err == nil {
			t.Errorf("malformed case accepted: %s", extra)
		}
	}
}
