package definitions

import (
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/andybalholm/cascadia"
	"gopkg.in/yaml.v3"
)

var operationNames = []string{
	"attribute.copy", "attribute.set", "attribute.remove", "string.replace", "regex.capture",
	"url.resolve", "url.build", "element.rename", "element.unwrap", "element.remove",
	"element.retain", "element.move", "element.create", "noscript.recover",
	"attribute.set_from", "algorithm.apply",
}

var conditionNames = []string{"exists", "text", "number", "descendant"}
var algorithmNames = []string{"abendblatt.deobfuscate"}
var featureCapabilities = []string{
	"transform.url.resolve.base",
	"transform.url.resolve.to",
	"element.create.typed_attributes",
	"element.create.self_target",
	"transform.noscript.recover.self",
	"value.descendant_attribute",
	"value.json",
}

// TransformCapabilities returns independent capability identifiers for executable operations.
func TransformCapabilities() []string {
	ids := make([]string, 0, len(operationNames)+len(conditionNames)+len(featureCapabilities))
	for _, name := range operationNames {
		ids = append(ids, "transform."+name)
	}
	for _, name := range conditionNames {
		ids = append(ids, "condition."+name)
	}
	ids = append(ids, featureCapabilities...)
	return ids
}

// AlgorithmCapabilities returns independently owned named algorithm identifiers.
func AlgorithmCapabilities() []string {
	return append([]string(nil), algorithmNames...)
}

func (p *validator) transforms(n *yaml.Node) (*program, error) {
	items, err := p.list(n, "content.transforms")
	if err != nil {
		return nil, err
	}
	if len(items) > MaxTransformSteps {
		return nil, p.bad(n, "content.transforms", "transform step limit exceeded")
	}
	result := &program{source: p.source, site: p.site}
	for i, item := range items {
		path := fmt.Sprintf("content.transforms[%d]", i+1)
		m, err := p.mapping(item, path, append([]string{"target", "selector", "when"}, operationNames...)...)
		if err != nil {
			return nil, err
		}
		a := arguments{p: p, node: item, fields: m, path: path}
		target := a.choice("target", "root", "descendants")
		step := transformStep{root: target == "root", line: item.Line, col: item.Column}
		if step.root {
			if m["selector"] != nil {
				return nil, p.bad(m["selector"], path+".selector", "root target does not accept a selector")
			}
		} else {
			step.selector = a.selector("selector")
		}
		if a.err != nil {
			return nil, a.err
		}
		for _, name := range operationNames {
			if params := m[name]; params != nil {
				if step.operation != nil {
					return nil, p.bad(item, path, "expected exactly one operation")
				}
				step.operation, err = p.operation(params, path+"."+name, name)
				if err != nil {
					return nil, err
				}
				step.name = name
				p.use("transform." + name)
			}
		}
		if step.operation == nil {
			return nil, p.bad(item, path, "expected exactly one operation")
		}
		if when := m["when"]; when != nil {
			step.conditions, err = p.conditions(when, path+".when")
			if err != nil {
				return nil, err
			}
		}
		result.steps = append(result.steps, step)
	}
	return result, nil
}

type arguments struct {
	p      *validator
	node   *yaml.Node
	fields map[string]*yaml.Node
	path   string
	err    error
}

func (p *validator) arguments(n *yaml.Node, path string, allowed ...string) *arguments {
	fields, err := p.mapping(n, path, allowed...)
	return &arguments{p: p, node: n, fields: fields, path: path, err: err}
}

func (a *arguments) reject(name, reason string) {
	if a.err == nil {
		node := a.fields[name]
		if node == nil {
			node = a.node
		}
		a.err = a.p.bad(node, a.path+"."+name, reason)
	}
}

func (a *arguments) string(name string, empty bool) string {
	if a.err != nil {
		return ""
	}
	n := a.fields[name]
	if n == nil || n.Kind != yaml.ScalarNode || n.Tag != "!!str" || (!empty && strings.TrimSpace(n.Value) == "") {
		a.reject(name, "expected "+map[bool]string{true: "string", false: "nonempty string"}[empty])
		return ""
	}
	return n.Value
}

func (a *arguments) attribute(name string) string {
	s := a.string(name, false)
	if a.err == nil && !attributePattern.MatchString(s) {
		a.reject(name, "invalid attribute name")
	}
	return strings.ToLower(s)
}

func (a *arguments) boolean(name string, fallback bool) bool {
	n := a.fields[name]
	if n == nil {
		return fallback
	}
	if n.Tag != "!!bool" || (n.Value != "true" && n.Value != "false") {
		a.reject(name, "expected true or false")
	}
	return n.Value == "true"
}

func (a *arguments) choice(name string, choices ...string) string {
	s := a.string(name, false)
	for _, choice := range choices {
		if s == choice {
			return s
		}
	}
	a.reject(name, "expected one of "+strings.Join(choices, ", "))
	return s
}

func (a *arguments) optionalChoice(name, fallback string, choices ...string) string {
	if a.fields[name] == nil {
		return fallback
	}
	return a.choice(name, choices...)
}

func (a *arguments) selector(name string) cascadia.Selector {
	s := a.string(name, false)
	if a.err != nil {
		return nil
	}
	selector, err := cascadia.Compile(s)
	if err != nil {
		a.err = a.p.error(a.fields[name], a.path+"."+name, err)
		return nil
	}
	return selector
}

func (a *arguments) input(name string) attributeInput {
	return attributeInput{name: a.attribute(name), required: a.boolean("required", false)}
}

func (a *arguments) target() selectorTarget {
	return selectorTarget{selector: a.selector("selector"), mode: a.optionalChoice("select", "all", "first", "last", "all")}
}

func (a *arguments) require(names ...string) error {
	for _, name := range names {
		if a.fields[name] == nil {
			return a.p.bad(a.node, a.path+"."+name, "required field")
		}
	}
	return nil
}

func (a *arguments) capturePattern() (*regexp.Regexp, int) {
	pattern := a.string("pattern", false)
	if len(pattern) > MaxPatternBytes {
		a.reject("pattern", "pattern byte limit exceeded")
	}
	if a.err != nil {
		return nil, 0
	}
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		a.err = a.p.error(a.fields["pattern"], a.path+".pattern", err)
		return nil, 0
	}
	if compiled.NumSubexp() > MaxCaptureGroups {
		a.reject("pattern", "capture group limit exceeded")
		return nil, 0
	}
	group := a.fields["group"]
	if group == nil || group.Tag != "!!int" {
		a.reject("group", "expected integer capture group")
		return nil, 0
	}
	index, err := strconv.Atoi(group.Value)
	if err != nil || index < 1 || index > compiled.NumSubexp() {
		a.reject("group", "capture group must exist and be at least 1")
		return nil, 0
	}
	return compiled, index
}

func (p *validator) operation(n *yaml.Node, path, name string) (operation, error) {
	switch name {
	case "attribute.copy":
		a := p.arguments(n, path, "from", "to", "required")
		o := attributeCopy{input: a.input("from"), to: a.attribute("to")}
		return o, a.err
	case "attribute.set":
		a := p.arguments(n, path, "name", "value")
		o := attributeSet{name: a.attribute("name"), value: a.string("value", true)}
		return o, a.err
	case "attribute.remove":
		a := p.arguments(n, path, "name")
		o := attributeRemove{name: a.attribute("name")}
		return o, a.err
	case "string.replace":
		a := p.arguments(n, path, "attribute", "old", "new", "required")
		o := stringReplace{input: a.input("attribute"), old: a.string("old", true), new: a.string("new", true)}
		if o.old == "" {
			a.reject("old", "replacement search string must not be empty")
		}
		return o, a.err
	case "regex.capture":
		return p.capture(n, path)
	case "url.resolve":
		a := p.arguments(n, path, "attribute", "to", "base", "required")
		o := urlResolve{input: a.input("attribute"), to: a.attribute("attribute")}
		if a.fields["to"] != nil {
			p.use("transform.url.resolve.to")
			o.to = a.attribute("to")
		}
		if a.fields["base"] != nil {
			p.use("transform.url.resolve.base")
			var err error
			o.base, err = p.scalarSource(a.fields["base"], path+".base")
			if err != nil {
				return nil, err
			}
		}
		return o, a.err
	case "url.build":
		return p.buildURL(n, path)
	case "element.rename":
		a := p.arguments(n, path, "tag")
		tag := a.choice("tag", safeRenameTags...)
		return elementRename{tag: tag}, a.err
	case "element.unwrap":
		a := p.arguments(n, path)
		return elementUnwrap{}, a.err
	case "element.remove":
		a := p.arguments(n, path)
		return elementRemove{}, a.err
	case "element.retain":
		return p.retain(n, path)
	case "element.move":
		return p.move(n, path)
	case "element.create":
		return p.create(n, path)
	case "noscript.recover":
		return p.recoverNoscript(n, path)
	case "attribute.set_from":
		return p.setFrom(n, path)
	case "algorithm.apply":
		return p.algorithm(n, path)
	default:
		return nil, p.bad(n, path, "unsupported operation")
	}
}

var safeRenameTags = strings.Fields("a article aside b blockquote caption code dd del div dl dt em figcaption figure h1 h2 h3 h4 h5 h6 i li ol p pre s section span strong table tbody td tfoot th thead tr u ul")
var safeConstructTags = append(append([]string(nil), safeRenameTags...), "br", "img")

type selectorTarget struct {
	selector cascadia.Selector
	mode     string
	self     bool
}

func (p *validator) selectorTarget(n *yaml.Node, path string, allowSelf bool) (selectorTarget, error) {
	if n == nil {
		return selectorTarget{}, p.bad(nil, path, "required field")
	}
	a := p.arguments(n, path, "selector", "select", "self")
	if a.fields["self"] != nil {
		if !allowSelf {
			return selectorTarget{}, p.bad(a.fields["self"], path+".self", "self target is unsupported")
		}
		if len(a.fields) != 1 || !a.boolean("self", false) {
			return selectorTarget{}, p.bad(n, path, "self target must be exactly {self: true}")
		}
		return selectorTarget{self: true}, a.err
	}
	target := a.target()
	return target, a.err
}

func (p *validator) retain(n *yaml.Node, path string) (operation, error) {
	a := p.arguments(n, path, "selector", "select", "required")
	o := elementRetain{selection: a.target(), required: a.boolean("required", false)}
	return o, a.err
}

func (p *validator) move(n *yaml.Node, path string) (operation, error) {
	a := p.arguments(n, path, "source", "target", "position", "required")
	if err := a.require("source", "target"); err != nil {
		return nil, err
	}
	source, err := p.selectorTarget(a.fields["source"], path+".source", false)
	if err != nil {
		return nil, err
	}
	target, err := p.selectorTarget(a.fields["target"], path+".target", false)
	if err != nil {
		return nil, err
	}
	o := elementMove{
		source:   source,
		target:   target,
		position: a.choice("position", "append", "prepend", "before", "after", "replace"),
		required: a.boolean("required", false),
	}
	return o, a.err
}

func (p *validator) create(n *yaml.Node, path string) (operation, error) {
	a := p.arguments(n, path, "target", "position", "node")
	if err := a.require("target", "node"); err != nil {
		return nil, err
	}
	target, err := p.selectorTarget(a.fields["target"], path+".target", true)
	if err != nil {
		return nil, err
	}
	if target.self {
		p.use("element.create.self_target")
	}
	node, err := p.constructNode(a.fields["node"], path+".node")
	if err != nil {
		return nil, err
	}
	return elementCreate{
		target:   target,
		position: a.choice("position", "append", "prepend", "before", "after", "replace"),
		node:     node,
	}, a.err
}

func (p *validator) recoverNoscript(n *yaml.Node, path string) (operation, error) {
	a := p.arguments(n, path, "source", "target", "position", "required")
	if err := a.require("source", "target"); err != nil {
		return nil, err
	}
	source, err := p.selectorTarget(a.fields["source"], path+".source", true)
	if err != nil {
		return nil, err
	}
	target, err := p.selectorTarget(a.fields["target"], path+".target", true)
	if err != nil {
		return nil, err
	}
	position := a.choice("position", "append", "prepend", "before", "after", "replace")
	if a.err != nil {
		return nil, a.err
	}
	if source.self != target.self {
		return nil, p.bad(n, path, "noscript self source and target must be used together")
	}
	if source.self && position != "replace" {
		return nil, p.bad(a.fields["position"], path+".position", "noscript self recovery requires replace")
	}
	if source.self {
		p.use("transform.noscript.recover.self")
	}
	return noscriptRecover{
		source:   source,
		target:   target,
		position: position,
		required: a.boolean("required", false),
	}, a.err
}

func (p *validator) setFrom(n *yaml.Node, path string) (operation, error) {
	a := p.arguments(n, path, "name", "value")
	if err := a.require("value"); err != nil {
		return nil, err
	}
	value, err := p.scalarSource(a.fields["value"], path+".value")
	if err != nil {
		return nil, err
	}
	return attributeSetFrom{name: a.attribute("name"), value: value}, a.err
}

func (p *validator) algorithm(n *yaml.Node, path string) (operation, error) {
	a := p.arguments(n, path, "name")
	name := a.choice("name", "abendblatt.deobfuscate")
	p.algorithms[name] = true
	return namedAlgorithm{name: name}, a.err
}

type constructedNode struct {
	tag        string
	text       string
	attributes []constructedAttribute
}

type constructedAttribute struct {
	name   string
	source scalarSource
}

func (p *validator) constructNode(n *yaml.Node, path string) (constructedNode, error) {
	if n == nil {
		return constructedNode{}, p.bad(nil, path, "required field")
	}
	a := p.arguments(n, path, "tag", "text", "attributes")
	node := constructedNode{tag: a.choice("tag", safeConstructTags...)}
	if a.fields["text"] != nil {
		node.text = a.string("text", true)
	}
	if node.tag == "img" && node.text != "" {
		a.reject("text", "img nodes cannot have text")
	}
	attributes, err := p.constructedAttributes(a.fields["attributes"], path+".attributes")
	if err != nil {
		return constructedNode{}, err
	}
	node.attributes = attributes
	return node, a.err
}

func (p *validator) constructedAttributes(n *yaml.Node, path string) ([]constructedAttribute, error) {
	if n == nil {
		return nil, nil
	}
	if n.Kind != yaml.MappingNode || n.Tag != "!!map" {
		return nil, p.bad(n, path, "expected mapping")
	}
	attributes := make([]constructedAttribute, 0, len(n.Content)/2)
	for i := 0; i < len(n.Content); i += 2 {
		key, value := n.Content[i], n.Content[i+1]
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || !attributePattern.MatchString(key.Value) {
			return nil, p.bad(key, path, "invalid attribute name")
		}
		var source scalarSource
		var err error
		if value.Kind == yaml.ScalarNode && value.Tag == "!!str" {
			source = literalSource(value.Value)
		} else {
			p.use("element.create.typed_attributes")
			source, err = p.scalarSource(value, path+"."+key.Value)
			if err != nil {
				return nil, err
			}
		}
		attributes = append(attributes, constructedAttribute{name: strings.ToLower(key.Value), source: source})
	}
	return attributes, nil
}

func (p *validator) capture(n *yaml.Node, path string) (operation, error) {
	a := p.arguments(n, path, "attribute", "pattern", "group", "required")
	o := regexCapture{input: a.input("attribute")}
	o.pattern, o.group = a.capturePattern()
	if a.err != nil {
		return nil, a.err
	}
	return o, nil
}

func (p *validator) scalarSource(n *yaml.Node, path string) (scalarSource, error) {
	a := p.arguments(n, path, "literal", "attribute", "descendant_attribute", "json", "required")
	sources := 0
	for _, name := range []string{"literal", "attribute", "descendant_attribute", "json"} {
		if a.fields[name] != nil {
			sources++
		}
	}
	if sources != 1 {
		return nil, p.bad(n, path, "expected exactly one scalar source")
	}
	if a.fields["literal"] != nil {
		if len(a.fields) != 1 {
			return nil, p.bad(n, path, "literal source accepts only literal")
		}
		s := literalSource(a.string("literal", true))
		return s, a.err
	}
	if a.fields["attribute"] != nil {
		s := a.input("attribute")
		return s, a.err
	}
	if a.fields["descendant_attribute"] != nil {
		p.use("value.descendant_attribute")
		if a.fields["required"] != nil {
			return nil, p.bad(a.fields["required"], path+".required", "required belongs inside descendant_attribute")
		}
		return p.descendantAttributeSource(a.fields["descendant_attribute"], path+".descendant_attribute")
	}
	p.use("value.json")
	if a.fields["required"] != nil {
		return nil, p.bad(a.fields["required"], path+".required", "required belongs inside json")
	}
	return p.jsonSource(a.fields["json"], path+".json")
}

func (p *validator) descendantAttributeSource(n *yaml.Node, path string) (scalarSource, error) {
	a := p.arguments(n, path, "selector", "name", "select", "required")
	return descendantAttributeSource{
		selection: selectorTarget{selector: a.selector("selector"), mode: a.optionalChoice("select", "first", "first", "last")},
		input:     attributeInput{name: a.attribute("name"), required: a.boolean("required", false)},
	}, a.err
}

func (p *validator) jsonSource(n *yaml.Node, path string) (scalarSource, error) {
	a := p.arguments(n, path, "attribute", "path", "required")
	if err := a.require("path"); err != nil {
		return nil, err
	}
	items, err := p.list(a.fields["path"], path+".path")
	if err != nil {
		return nil, err
	}
	if len(items) > MaxJSONTraversal {
		return nil, p.bad(a.fields["path"], path+".path", "JSON traversal limit exceeded")
	}
	source := jsonValueSource{input: attributeInput{name: a.attribute("attribute"), required: a.boolean("required", false)}}
	for _, item := range items {
		m, itemErr := p.mapping(item, path+".path", "field", "index")
		if itemErr != nil {
			return nil, itemErr
		}
		if len(m) != 1 {
			return nil, p.bad(item, path+".path", "expected exactly one field or index segment")
		}
		if field := m["field"]; field != nil {
			value, valueErr := p.text(field, path+".path.field")
			if valueErr != nil {
				return nil, valueErr
			}
			source.path = append(source.path, jsonPathSegment{field: value})
			continue
		}
		index := m["index"]
		if index == nil || index.Tag != "!!int" {
			return nil, p.bad(index, path+".path.index", "expected nonnegative integer index")
		}
		value, valueErr := strconv.Atoi(index.Value)
		if valueErr != nil || value < 0 {
			return nil, p.bad(index, path+".path.index", "expected nonnegative integer index")
		}
		source.path = append(source.path, jsonPathSegment{index: value, isIndex: true})
	}
	return source, a.err
}

func (p *validator) buildURL(n *yaml.Node, path string) (operation, error) {
	a := p.arguments(n, path, "attribute", "base", "path", "query")
	o := urlBuild{attribute: a.attribute("attribute"), base: a.string("base", false)}
	if a.err != nil {
		return nil, a.err
	}
	base, err := httpURL(o.base)
	if err != nil {
		return nil, p.error(a.fields["base"], path+".base", err)
	}
	if _, err := url.ParseQuery(base.RawQuery); err != nil {
		return nil, p.error(a.fields["base"], path+".base", err)
	}
	if segments := a.fields["path"]; segments != nil {
		items, err := p.list(segments, path+".path")
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			s, err := p.scalarSource(item, path+".path")
			if err != nil {
				return nil, err
			}
			if literal, ok := s.(literalSource); ok && (literal == "" || literal == "." || literal == "..") {
				return nil, p.bad(item, path+".path", "invalid literal path segment")
			}
			o.path = append(o.path, s)
		}
	}
	if query := a.fields["query"]; query != nil {
		items, err := p.list(query, path+".query")
		if err != nil {
			return nil, err
		}
		names := map[string]bool{}
		for _, item := range items {
			q := p.arguments(item, path+".query", "name", "value")
			name := q.string("name", false)
			if q.err != nil {
				return nil, q.err
			}
			if names[name] {
				return nil, p.bad(item, path+".query", "duplicate query parameter name")
			}
			names[name] = true
			if q.fields["value"] == nil {
				return nil, p.bad(item, path+".query.value", "required field")
			}
			s, err := p.scalarSource(q.fields["value"], path+".query.value")
			if err != nil {
				return nil, err
			}
			o.query = append(o.query, queryValue{name, s})
		}
	}
	return o, nil
}

func (p *validator) conditions(n *yaml.Node, path string) ([]condition, error) {
	items, err := p.list(n, path)
	if err != nil {
		return nil, err
	}
	if len(items) > MaxConditions {
		return nil, p.bad(n, path, "condition count limit exceeded")
	}
	var result []condition
	for _, item := range items {
		m, err := p.mapping(item, path, conditionNames...)
		if err != nil {
			return nil, err
		}
		if len(m) != 1 {
			return nil, p.bad(item, path, "expected exactly one condition")
		}
		for name, params := range m {
			p.use("condition." + name)
			c, err := p.condition(params, path+"."+name, name)
			if err != nil {
				return nil, err
			}
			result = append(result, c)
		}
	}
	return result, nil
}

func (p *validator) condition(n *yaml.Node, path, name string) (condition, error) {
	switch name {
	case "exists":
		a := p.arguments(n, path, "attribute", "present")
		c := existsCondition{attribute: a.attribute("attribute"), present: a.boolean("present", true)}
		return c, a.err
	case "descendant":
		a := p.arguments(n, path, "selector", "present")
		c := descendantCondition{selector: a.selector("selector"), present: a.boolean("present", true)}
		return c, a.err
	case "text":
		a := p.arguments(n, path, "attribute", "required", "comparison", "value")
		c := textCondition{comparison: a.choice("comparison", "equals", "contains", "prefix", "suffix"), value: a.string("value", true)}
		if a.fields["attribute"] != nil {
			c.input = a.input("attribute")
		} else if a.fields["required"] != nil {
			a.reject("required", "required applies only to attribute inputs")
		}
		return c, a.err
	case "number":
		a := p.arguments(n, path, "attribute", "required", "comparison", "value")
		c := numberCondition{input: a.input("attribute"), comparison: a.choice("comparison", "eq", "ne", "lt", "le", "gt", "ge")}
		value := a.fields["value"]
		if value == nil || (value.Tag != "!!float" && value.Tag != "!!int") {
			a.reject("value", "expected finite number")
		} else {
			number, err := strconv.ParseFloat(value.Value, 64)
			if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
				a.reject("value", "expected finite number")
			}
			c.value = number
		}
		return c, a.err
	default:
		return nil, p.bad(n, path, "unsupported condition")
	}
}
