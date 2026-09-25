package definitions

import (
	"maps"
	"slices"
	"testing"
)

const featureDefinition = `schema: 1
site: features
hosts: [www.example.com]
metadata:
  title:
    - text: h1
  author:
    - attribute: {selector: 'meta[name=author]', name: content}
  date_published:
    - text_capture: {selector: .byline, pattern: '([0-9]{4})', group: 1}
content:
  groups: [[article]]
  remove: [.ad]
  preserve: [figure]
  transforms:
    - target: descendants
      selector: img
      when:
        - exists: {attribute: data-src}
        - text: {comparison: contains, value: photo}
        - number: {attribute: width, comparison: gt, value: 10}
        - descendant: {selector: span}
      url.resolve:
        attribute: data-src
        to: src
        base: {attribute: data-base}
    - target: descendants
      selector: a
      url.build:
        attribute: href
        base: https://example.com/
        path: [{json: {attribute: data-json, path: [{field: id}]}}]
    - target: descendants
      selector: img
      attribute.set_from:
        name: alt
        value: {descendant_attribute: {selector: span, name: title}}
    - target: descendants
      selector: figure
      element.create:
        target: {self: true}
        position: append
        node: {tag: img, attributes: {src: {attribute: data-src}}}
    - target: descendants
      selector: noscript
      noscript.recover:
        source: {self: true}
        target: {self: true}
        position: replace
    - target: descendants
      selector: p
      algorithm.apply: {name: abendblatt.deobfuscate}
`

// plainDefinition uses the same operations without their conditional parameter features.
const plainDefinition = `schema: 1
site: plain
hosts: ['*.example.org']
content:
  groups: [[article]]
  transforms:
    - target: descendants
      selector: img
      url.resolve: {attribute: src}
    - target: descendants
      selector: a
      url.build:
        attribute: href
        base: https://example.org/
        path: [{literal: a}]
        query: [{name: q, value: {attribute: data-q}}]
    - target: descendants
      selector: img
      attribute.set_from: {name: alt, value: {literal: photo}}
    - target: descendants
      selector: figure
      element.create:
        target: {selector: figcaption}
        position: append
        node: {tag: span, attributes: {class: caption}}
    - target: descendants
      selector: figure
      noscript.recover:
        source: {selector: noscript}
        target: {selector: img}
        position: replace
`

func TestUsedCapabilitiesRecordsValidatedFeatures(t *testing.T) {
	features := []string{
		"condition.descendant", "condition.exists", "condition.number", "condition.text",
		"content.default_cleaner", "content.groups", "content.preserve", "content.remove",
		"element.create.self_target", "element.create.typed_attributes", "hosts.exact-www",
		"metadata.attribute", "metadata.text", "metadata.text_capture",
		"transform.algorithm.apply", "transform.attribute.set_from", "transform.element.create",
		"transform.noscript.recover", "transform.noscript.recover.self", "transform.url.build",
		"transform.url.resolve", "transform.url.resolve.base", "transform.url.resolve.to",
		"value.descendant_attribute", "value.json",
	}
	plain := []string{
		"content.default_cleaner", "content.groups", "hosts.wildcard",
		"transform.attribute.set_from", "transform.element.create", "transform.noscript.recover",
		"transform.url.build", "transform.url.resolve",
	}
	tests := []struct {
		name       string
		files      map[string]string
		operations []string
		algorithms []string
		sites      map[string]string
	}{
		{"conditional features", map[string]string{"features.yaml": featureDefinition},
			features, []string{"abendblatt.deobfuscate"}, map[string]string{"features.yaml": "features"}},
		{"plain parameters", map[string]string{"plain.yaml": plainDefinition},
			plain, nil, map[string]string{"plain.yaml": "plain"}},
		{"union across files", map[string]string{"features.yaml": featureDefinition, "plain.yaml": plainDefinition},
			slices.Sorted(slices.Values(append(slices.Clone(features), "hosts.wildcard"))), []string{"abendblatt.deobfuscate"},
			map[string]string{"features.yaml": "features", "plain.yaml": "plain"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range test.files {
				write(t, dir, name, content)
			}
			snapshot, err := LoadDirectory(dir)
			if err != nil {
				t.Fatal(err)
			}
			operations, algorithms := snapshot.UsedCapabilities()
			if !slices.Equal(operations, test.operations) || !slices.Equal(algorithms, test.algorithms) {
				t.Fatalf("recorded capabilities\n got  %v %v\n want %v %v", operations, algorithms, test.operations, test.algorithms)
			}
			if sites := snapshot.Sites(); !maps.Equal(sites, test.sites) {
				t.Fatalf("sites = %v, want %v", sites, test.sites)
			}
			operations[0] = "mutated"
			if again, _ := snapshot.UsedCapabilities(); !slices.Equal(again, test.operations) {
				t.Fatalf("returned capabilities share storage: %v", again)
			}
		})
	}
}

func TestUsedCapabilitiesOfEmptySnapshot(t *testing.T) {
	var snapshot *Snapshot
	if operations, algorithms := snapshot.UsedCapabilities(); len(operations)+len(algorithms) != 0 {
		t.Fatalf("empty snapshot reported capabilities: %v %v", operations, algorithms)
	}
	if sites := snapshot.Sites(); len(sites) != 0 {
		t.Fatalf("empty snapshot reported sites: %v", sites)
	}
}
