package definitions

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestTransformValidation(t *testing.T) {
	tests := []string{
		`{target: anywhere, element.remove: {}}`,
		`{target: root, selector: p, element.remove: {}}`,
		`{target: descendants, element.remove: {}}`,
		`{target: descendants, selector: '[', element.remove: {}}`,
		`{target: root}`,
		`{target: root, element.remove: {}, element.unwrap: {}}`,
		`{target: root, script: "run()"}`,
		`{target: root, element.rename: div}`,
		`{target: root, element.rename: {tag: script}}`,
		`{target: root, element.rename: {tag: img}}`,
		`{target: root, element.unwrap: {unexpected: true}}`,
		`{target: root, element.retain: {}}`,
		`{target: root, element.retain: {selector: img, select: random}}`,
		`{target: root, element.move: {source: {selector: img}, position: append}}`,
		`{target: root, element.move: {source: {self: true}, target: {selector: figure}, position: append}}`,
		`{target: root, element.move: {source: {selector: img}, target: {selector: figure, select: random}, position: append}}`,
		`{target: root, element.create: {target: {selector: figure}, position: append}}`,
		`{target: root, element.create: {target: {self: false}, position: append, node: {tag: p}}}`,
		`{target: root, element.create: {target: {self: true, select: first}, position: append, node: {tag: p}}}`,
		`{target: root, element.create: {target: {selector: figure}, position: append, node: {tag: script}}}`,
		`{target: root, element.create: {target: {selector: figure}, position: append, node: {tag: img, text: x}}}`,
		`{target: root, element.create: {target: {selector: figure}, position: append, node: {tag: p, attributes: {onclick: 1}}}}`,
		`{target: root, noscript.recover: {source: {selector: noscript}, position: replace}}`,
		`{target: root, noscript.recover: {source: {selector: noscript}, target: {selector: .slot}, position: move}}`,
		`{target: root, noscript.recover: {source: {self: true}, target: {selector: .slot}, position: replace}}`,
		`{target: root, noscript.recover: {source: {self: true}, target: {self: true}, position: append}}`,
		`{target: root, attribute.set_from: {name: src}}`,
		`{target: root, attribute.set_from: {name: src, value: {json: {attribute: data-props, path: []}}}}`,
		`{target: root, attribute.set_from: {name: src, value: {json: {attribute: data-props, path: [{field: sources, index: 0}]}}}}`,
		`{target: root, attribute.set_from: {name: src, value: {json: {attribute: data-props, path: [{index: -1}]}}}}`,
		`{target: root, attribute.set_from: {name: src, value: {descendant_attribute: {selector: img, name: src, select: all}}}}`,
		`{target: root, algorithm.apply: {name: unknown.decoder}}`,
		`{target: root, algorithm.apply: {name: abendblatt.deobfuscate, extra: true}}`,
		`{target: root, attribute.copy: {from: src}}`,
		`{target: root, attribute.copy: {from: src, to: 'bad name'}}`,
		`{target: root, attribute.copy: {from: src, to: alt, required: 'true'}}`,
		`{target: root, attribute.set: {name: alt, value: 640}}`,
		`{target: root, attribute.set: {name: alt}}`,
		`{target: root, attribute.set: {name: alt, value: new, required: true}}`,
		`{target: root, attribute.remove: {name: null}}`,
		`{target: root, string.replace: {attribute: src, old: '', new: x}}`,
		`{target: root, string.replace: {attribute: src, old: x}}`,
		`{target: root, regex.capture: {attribute: src, pattern: '(', group: 1}}`,
		`{target: root, regex.capture: {attribute: src, pattern: '(a)', group: 0}}`,
		`{target: root, regex.capture: {attribute: src, pattern: '(a)', group: 2}}`,
		`{target: root, regex.capture: {attribute: src, pattern: '(a)', group: '1'}}`,
		`{target: root, regex.capture: {attribute: src, pattern: '(a)'}}`,
		`{target: root, regex.capture: {attribute: src, pattern: '(a)\1', group: 1}}`,
		`{target: root, regex.capture: {attribute: src, pattern: '(?=a)(a)', group: 1}}`,
		`{target: root, url.resolve: {attribute: src, base: 'https://example.com'}}`,
		`{target: root, url.build: {attribute: src, base: '/relative'}}`,
		`{target: root, url.build: {attribute: src, base: 'file:///example'}}`,
		`{target: root, url.build: {attribute: src, base: 'https://user:pass@example.com'}}`,
		`{target: root, url.build: {attribute: src, base: 'https://example.com/?q=%zz'}}`,
		`{target: root, url.build: {attribute: src, base: 'https://example.com', path: [id]}}`,
		`{target: root, url.build: {attribute: src, base: 'https://example.com', path: [{literal: '..'}]}}`,
		`{target: root, url.build: {attribute: src, base: 'https://example.com', path: [{literal: a, attribute: title}]}}`,
		`{target: root, url.build: {attribute: src, base: 'https://example.com', query: [{name: id}]}}`,
		`{target: root, url.build: {attribute: src, base: 'https://example.com', query: [{name: id, value: {literal: a}}, {name: id, value: {literal: b}}]}}`,
		`{target: root, element.remove: {}, when: []}`,
		`{target: root, element.remove: {}, when: [{exists: {attribute: src, present: 'true'}}]}`,
		`{target: root, element.remove: {}, when: [{expression: '1 + 1'}]}`,
		`{target: root, element.remove: {}, when: [{exists: {attribute: src}, text: {equals: x}}]}`,
		`{target: root, element.remove: {}, when: [{text: {comparison: regex, value: x}}]}`,
		`{target: root, element.remove: {}, when: [{text: {comparison: equals, value: x, required: true}}]}`,
		`{target: root, element.remove: {}, when: [{number: {attribute: width, comparison: gt, value: '5'}}]}`,
		`{target: root, element.remove: {}, when: [{number: {attribute: width, comparison: gt, value: .nan}}]}`,
		`{target: root, element.remove: {}, when: [{number: {attribute: width, comparison: gt, value: .inf}}]}`,
		`{target: root, element.remove: {}, when: [{descendant: {selector: '['}}]}`,
	}
	tests = append(tests,
		fmt.Sprintf("{target: root, regex.capture: {attribute: src, pattern: '%s', group: 1}}", strings.Repeat("(a)", MaxCaptureGroups+1)),
		fmt.Sprintf("{target: root, regex.capture: {attribute: src, pattern: '(%s)', group: 1}}", strings.Repeat("a", MaxPatternBytes)),
		"{target: root, element.remove: {}, when: ["+strings.Repeat("{exists: {attribute: src}},", MaxConditions)+"{exists: {attribute: src}}]}",
		"{target: root, attribute.set_from: {name: src, value: {json: {attribute: data-props, path: ["+strings.Repeat("{field: a},", MaxJSONTraversal)+"{field: a}], required: true}}}}",
	)
	for i, step := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			dir := t.TempDir()
			write(t, dir, "site.yaml", valid+"  transforms:\n    - "+step+"\n")
			s, err := LoadDirectory(dir)
			var diagnostic *Error
			if s != nil || !errors.As(err, &diagnostic) || diagnostic.Site != "test" || !strings.Contains(diagnostic.Field, "content.transforms") || diagnostic.Line == 0 {
				t.Fatalf("invalid step accepted or context absent: %s: %v", step, err)
			}
		})
	}
}

func TestTransformLoadBoundaries(t *testing.T) {
	for _, test := range []struct {
		name, step string
		count      int
		valid      bool
	}{
		{"max steps", "{target: root, element.unwrap: {}}", MaxTransformSteps, true},
		{"excess steps", "{target: root, element.unwrap: {}}", MaxTransformSteps + 1, false},
		{"max pattern", "{target: root, regex.capture: {attribute: title, pattern: '(" + strings.Repeat("a", MaxPatternBytes-2) + ")', group: 1}}", 1, true},
		{"max groups", "{target: root, regex.capture: {attribute: title, pattern: '" + strings.Repeat("(a)", MaxCaptureGroups) + "', group: 16}}", 1, true},
		{"max conditions", "{target: root, element.unwrap: {}, when: [" + strings.Repeat("{exists: {attribute: title}},", MaxConditions-1) + "{exists: {attribute: title}}]}", 1, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			write(t, dir, "site.yaml", valid+"  transforms:\n"+strings.Repeat("    - "+test.step+"\n", test.count))
			_, err := LoadDirectory(dir)
			if (err == nil) != test.valid {
				t.Fatalf("boundary: %v", err)
			}
		})
	}
}

type cancellingContext struct {
	context.Context
	calls, cancelAt int
	cause           error
}

func (c *cancellingContext) Err() error {
	c.calls++
	if c.calls >= c.cancelAt {
		return c.cause
	}
	return nil
}

func TestTransformCancellationAndSourceOwnership(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "site.yaml", valid+`  transforms:
    - target: descendants
      selector: p
      attribute.set: {name: title, value: Changed}
`)
	s, err := LoadDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(filepath.Join(dir, "site.yaml")); err != nil {
		t.Fatal(err)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<article>" + strings.Repeat("<p>Original</p>", 100) + "</article>"))
	if err != nil {
		t.Fatal(err)
	}
	program := s.Match("example.com").Content.OrderedTransforms
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		// Initial inspection and step matching finish before this cancellation point.
		ctx := &cancellingContext{Context: context.Background(), cancelAt: 800, cause: cause}
		result, err := program.CopyAndExecute(ctx, doc.Find("article"), "https://example.com")
		var diagnostic *OperationError
		if result != nil || !errors.Is(err, cause) || !errors.As(err, &diagnostic) || diagnostic.Step != 1 {
			t.Fatalf("cancellation not preserved: %v, calls=%d", err, ctx.calls)
		}
		if doc.Find("[title]").Length() != 0 || doc.Find("p").Length() != 100 {
			t.Fatal("partial transform execution modified source")
		}
	}
}
