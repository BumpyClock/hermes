package definitionbundle_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	hermes "github.com/BumpyClock/hermes"
	bundle "github.com/BumpyClock/hermes/internal/definitionbundle"
)

func TestStructuredRequirementsIncludeMetadataAndNamedAlgorithms(t *testing.T) {
	m, files := pinned(t)
	files["definitions/example.yaml"] = []byte(`schema: 1
site: synthetic-example
hosts: [example.com]
metadata:
  date_published:
    - text_capture:
        selector: .byline
        pattern: '([0-9]{4}-[0-9]{2}-[0-9]{2})'
        group: 1
content:
  groups: [[article]]
  transforms:
    - target: descendants
      selector: p.obfuscated
      algorithm.apply: {name: abendblatt.deobfuscate}
`)
	suite, err := bundle.ParseSuite(files)
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(t.TempDir(), "definitions")
	if err := bundle.WriteDefinitions(files, directory); err != nil {
		t.Fatal(err)
	}
	support := hermes.DefinitionCapabilities()
	m.Engine.Operations = slices.Clone(support.Capabilities)
	m.Engine.Algorithms = slices.Clone(support.Algorithms)
	if err := m.AuditRequirements(files, suite, directory); err != nil {
		t.Fatal(err)
	}
	for _, capability := range []string{"metadata.text_capture", "transform.algorithm.apply"} {
		t.Run(capability, func(t *testing.T) {
			original := m.Engine.Operations
			m.Engine.Operations = slices.DeleteFunc(slices.Clone(original), func(name string) bool {
				return name == capability
			})
			defer func() { m.Engine.Operations = original }()
			if err := m.AuditRequirements(files, suite, directory); err == nil || !strings.Contains(err.Error(), capability) {
				t.Fatalf("undeclared metadata/algorithm operation accepted: %v", err)
			}
		})
	}
	m.Engine.Algorithms = []string{}
	if err := m.AuditRequirements(files, suite, directory); err == nil || !strings.Contains(err.Error(), "abendblatt.deobfuscate") {
		t.Fatalf("undeclared named algorithm accepted: %v", err)
	}
}

func TestParameterExtensionsRequireTheirOwnCapabilities(t *testing.T) {
	m, files := pinned(t)
	files["definitions/example.yaml"] = []byte(`schema: 1
site: synthetic-example
hosts: [example.com]
content:
  groups: [[article]]
  transforms:
    - target: descendants
      selector: img
      url.resolve:
        attribute: data-original
        to: src
        base: {attribute: src}
    - target: root
      element.create:
        target: {self: true}
        position: append
        node:
          tag: img
          attributes:
            src:
              json:
                attribute: data-props
                path: [{field: image}]
            alt:
              descendant_attribute:
                selector: .caption
                name: title
    - target: descendants
      selector: noscript
      noscript.recover:
        source: {self: true}
        target: {self: true}
        position: replace
`)
	suite, err := bundle.ParseSuite(files)
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(t.TempDir(), "definitions")
	if err := bundle.WriteDefinitions(files, directory); err != nil {
		t.Fatal(err)
	}
	support := hermes.DefinitionCapabilities()
	m.Engine.Operations = slices.Clone(support.Capabilities)
	if err := m.AuditRequirements(files, suite, directory); err != nil {
		t.Fatal(err)
	}
	for _, capability := range []string{
		"transform.url.resolve.base", "transform.url.resolve.to",
		"element.create.self_target", "element.create.typed_attributes",
		"value.json", "value.descendant_attribute", "transform.noscript.recover.self",
	} {
		t.Run(capability, func(t *testing.T) {
			original := m.Engine.Operations
			m.Engine.Operations = slices.DeleteFunc(slices.Clone(original), func(name string) bool {
				return name == capability
			})
			defer func() { m.Engine.Operations = original }()
			if err := m.AuditRequirements(files, suite, directory); err == nil || !strings.Contains(err.Error(), capability) {
				t.Fatalf("undeclared parameter extension accepted: %v", err)
			}
		})
	}
}
