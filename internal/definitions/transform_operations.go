package definitions

import (
	"fmt"
	stdhtml "html"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type attributeInput struct {
	name     string
	required bool
}

func (a attributeInput) read(e *execution, n *html.Node) (string, bool, error) {
	for _, attr := range n.Attr {
		if err := inspectAttribute(e, attr); err != nil {
			return "", false, err
		}
		if attr.Namespace == "" && attr.Key == a.name {
			if err := checkValueSize(attr.Val); err != nil {
				return "", false, err
			}
			if a.required && strings.TrimSpace(attr.Val) == "" {
				return "", false, fmt.Errorf("required attribute %q is empty", a.name)
			}
			return attr.Val, true, nil
		}
	}
	if a.required {
		return "", false, fmt.Errorf("required attribute %q is absent", a.name)
	}
	return "", false, nil
}

func setAttribute(e *execution, n *html.Node, name, value string) error {
	for _, attr := range n.Attr {
		if err := inspectAttribute(e, attr); err != nil {
			return err
		}
	}
	if err := e.spend(len(name) + 1); err != nil {
		return err
	}
	if err := e.value(value); err != nil {
		return err
	}
	for i := range n.Attr {
		if n.Attr[i].Namespace == "" && n.Attr[i].Key == name {
			n.Attr[i].Val = value
			return nil
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: name, Val: value})
	return nil
}

type attributeCopy struct {
	input attributeInput
	to    string
}

func (o attributeCopy) apply(e *execution, n *html.Node) error {
	value, exists, err := o.input.read(e, n)
	if err != nil || !exists {
		return err
	}
	return setAttribute(e, n, o.to, value)
}

type attributeSet struct{ name, value string }

func (o attributeSet) apply(e *execution, n *html.Node) error {
	return setAttribute(e, n, o.name, o.value)
}

type attributeSetFrom struct {
	name  string
	value scalarSource
}

func (o attributeSetFrom) apply(e *execution, n *html.Node) error {
	value, exists, err := o.value.read(e, n)
	if err != nil || !exists {
		return err
	}
	return setAttribute(e, n, o.name, value)
}

type attributeRemove struct{ name string }

func (o attributeRemove) apply(e *execution, n *html.Node) error {
	for _, attr := range n.Attr {
		if err := inspectAttribute(e, attr); err != nil {
			return err
		}
	}
	kept := n.Attr[:0]
	for _, attr := range n.Attr {
		if attr.Namespace != "" || attr.Key != o.name {
			kept = append(kept, attr)
		}
	}
	n.Attr = kept
	return nil
}

func inspectAttribute(e *execution, attribute html.Attribute) error {
	return e.spend(len(attribute.Key) + len(attribute.Val) + 1)
}

type stringReplace struct {
	input    attributeInput
	old, new string
}

func (o stringReplace) apply(e *execution, n *html.Node) error {
	value, exists, err := o.input.read(e, n)
	if err != nil || !exists {
		return err
	}
	size := len(value) + strings.Count(value, o.old)*(len(o.new)-len(o.old))
	if size > MaxValueBytes {
		return fmt.Errorf("%w: replacement exceeds %d bytes", ErrTransformLimit, MaxValueBytes)
	}
	return setAttribute(e, n, o.input.name, strings.ReplaceAll(value, o.old, o.new))
}

type regexCapture struct {
	input   attributeInput
	pattern *regexp.Regexp
	group   int
}

func (o regexCapture) apply(e *execution, n *html.Node) error {
	value, exists, err := o.input.read(e, n)
	if err != nil || !exists {
		return err
	}
	match := o.pattern.FindStringSubmatchIndex(value)
	if match == nil || match[o.group*2] < 0 {
		if o.input.required {
			return fmt.Errorf("required pattern capture %d did not match attribute %q", o.group, o.input.name)
		}
		return nil
	}
	captured := value[match[o.group*2]:match[o.group*2+1]]
	if o.input.required && strings.TrimSpace(captured) == "" {
		return fmt.Errorf("required pattern capture %d is empty", o.group)
	}
	return setAttribute(e, n, o.input.name, captured)
}

type urlResolve struct {
	input attributeInput
	to    string
	base  scalarSource
}

func (o urlResolve) apply(e *execution, n *html.Node) error {
	value, exists, err := o.input.read(e, n)
	if err != nil || !exists {
		return err
	}
	var base *url.URL
	if o.base == nil {
		base, err = e.defaultResolutionBase()
	} else {
		var baseValue string
		baseValue, exists, err = o.base.read(e, n)
		if err != nil || !exists {
			return err
		}
		base, err = e.resolutionBase(baseValue)
	}
	if err != nil {
		return err
	}
	ref, err := url.Parse(value)
	if err != nil {
		return err
	}
	resolved := base.ResolveReference(ref).String()
	if _, err := httpURL(resolved); err != nil {
		return err
	}
	return setAttribute(e, n, o.to, resolved)
}

type scalarSource interface {
	read(*execution, *html.Node) (string, bool, error)
}

type literalSource string

func (s literalSource) read(e *execution, _ *html.Node) (string, bool, error) {
	return string(s), true, e.value(string(s))
}

type queryValue struct {
	name   string
	source scalarSource
}

type urlBuild struct {
	attribute string
	base      *url.URL
	path      []scalarSource
	query     []queryValue
}

func (o urlBuild) apply(e *execution, n *html.Node) error {
	// The copy keeps the shared validated base unchanged; it never has user info.
	u := *o.base
	for _, source := range o.path {
		value, exists, err := source.read(e, n)
		if err != nil || !exists {
			return err
		}
		if value == "." || value == ".." || value == "" {
			return fmt.Errorf("URL path segment must not be empty, '.' or '..'")
		}
		raw := strings.TrimRight(u.EscapedPath(), "/") + "/" + url.PathEscape(value)
		path, err := url.PathUnescape(raw)
		if err != nil {
			return err
		}
		u.Path = path
		u.RawPath = raw
		if err := e.value(u.String()); err != nil {
			return err
		}
	}
	query := u.Query()
	for _, part := range o.query {
		value, exists, err := part.source.read(e, n)
		if err != nil || !exists {
			return err
		}
		query.Set(part.name, value)
		u.RawQuery = query.Encode()
		if err := e.value(u.String()); err != nil {
			return err
		}
	}
	return setAttribute(e, n, o.attribute, u.String())
}

type elementRename struct{ tag string }

func (o elementRename) apply(e *execution, n *html.Node) error {
	if err := e.value(o.tag); err != nil {
		return err
	}
	n.Data = o.tag
	n.DataAtom = atom.Lookup([]byte(o.tag))
	return nil
}

type elementUnwrap struct{}

func (elementUnwrap) apply(_ *execution, n *html.Node) error {
	parent := n.Parent
	if parent == nil {
		return nil
	}
	for n.FirstChild != nil {
		child := n.FirstChild
		n.RemoveChild(child)
		parent.InsertBefore(child, n)
	}
	parent.RemoveChild(n)
	return nil
}

type elementRemove struct{}

func (elementRemove) apply(_ *execution, n *html.Node) error {
	if n.Parent != nil {
		n.Parent.RemoveChild(n)
	}
	return nil
}

type elementRetain struct {
	selection selectorTarget
	required  bool
}

func (o elementRetain) apply(e *execution, n *html.Node) error {
	nodes, err := selectedDescendants(e, n, o.selection)
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		if o.required {
			return fmt.Errorf("required retained descendants are absent")
		}
		return nil
	}
	nodes = outermostNodes(nodes)
	for _, node := range nodes {
		node.Parent.RemoveChild(node)
	}
	for n.FirstChild != nil {
		n.RemoveChild(n.FirstChild)
	}
	for _, node := range nodes {
		n.AppendChild(node)
	}
	return nil
}

type elementMove struct {
	source, target selectorTarget
	position       string
	required       bool
}

func (o elementMove) apply(e *execution, n *html.Node) error {
	sources, err := selectedDescendants(e, n, o.source)
	if err != nil {
		return err
	}
	targets, err := selectedDescendants(e, n, o.target)
	if err != nil {
		return err
	}
	if len(sources) == 0 || len(targets) == 0 {
		if o.required {
			return fmt.Errorf("required move source or target is absent")
		}
		return nil
	}
	sources = outermostNodes(sources)
	if hasSelfOrAncestorIn(targets, nodeSet(sources)) ||
		(o.position == "replace" && hasSelfOrAncestorIn(sources, nodeSet(targets))) {
		return fmt.Errorf("move destination cannot be the source or a descendant containing it")
	}
	added := 0
	for _, source := range sources {
		count, countErr := e.countNodes(source)
		if countErr != nil {
			return countErr
		}
		added += count * (len(targets) - 1)
	}
	if err := e.allowGenerated(n, added); err != nil {
		return err
	}
	for index, target := range targets {
		if !isAncestor(n, target) {
			continue
		}
		nodes := sources
		if index > 0 {
			nodes = cloneNodes(sources)
		}
		if err := insertNodes(target, nodes, o.position); err != nil {
			return err
		}
	}
	return nil
}

type elementCreate struct {
	target   selectorTarget
	position string
	node     constructedNode
}

func (o elementCreate) apply(e *execution, n *html.Node) error {
	targets, err := selectedDescendants(e, n, o.target)
	if err != nil {
		return err
	}
	attributes := make([]html.Attribute, 0, len(o.node.attributes))
	for _, attribute := range o.node.attributes {
		value, exists, valueErr := attribute.source.read(e, n)
		if valueErr != nil {
			return valueErr
		}
		if !exists {
			return nil
		}
		attributes = append(attributes, html.Attribute{Key: attribute.name, Val: value})
	}
	if o.node.text != "" {
		if err := e.value(o.node.text); err != nil {
			return err
		}
	}
	added := len(targets)
	if o.node.text != "" {
		added += len(targets)
	}
	if err := e.allowGenerated(n, added); err != nil {
		return err
	}
	for _, target := range targets {
		if !isAncestor(n, target) {
			continue
		}
		node := &html.Node{Type: html.ElementNode, Data: o.node.tag, DataAtom: atom.Lookup([]byte(o.node.tag))}
		node.Attr = append(node.Attr, attributes...)
		if o.node.text != "" {
			node.AppendChild(&html.Node{Type: html.TextNode, Data: o.node.text})
		}
		if err := insertNodes(target, []*html.Node{node}, o.position); err != nil {
			return err
		}
	}
	return nil
}

type noscriptRecover struct {
	source, target selectorTarget
	position       string
	required       bool
}

func (o noscriptRecover) apply(e *execution, n *html.Node) error {
	sources, err := selectedDescendants(e, n, o.source)
	if err != nil {
		return err
	}
	targets, err := selectedDescendants(e, n, o.target)
	if err != nil {
		return err
	}
	if len(sources) == 0 || len(targets) == 0 {
		if o.required {
			return fmt.Errorf("required noscript source or target is absent")
		}
		return nil
	}
	var images []*html.Node
	for _, source := range sources {
		if source.DataAtom != atom.Noscript && source.Data != "noscript" {
			continue
		}
		recovered, recoverErr := recoverNoscriptImages(e, source)
		if recoverErr != nil {
			return recoverErr
		}
		images = append(images, recovered...)
	}
	if len(images) == 0 {
		if o.required {
			return fmt.Errorf("required noscript image is absent")
		}
		return nil
	}
	if err := e.allowGenerated(n, len(images)*len(targets)); err != nil {
		return err
	}
	for _, target := range targets {
		if !isAncestor(n, target) {
			continue
		}
		if err := insertNodes(target, cloneNodes(images), o.position); err != nil {
			return err
		}
	}
	for _, source := range sources {
		if source.Parent != nil {
			source.Parent.RemoveChild(source)
		}
	}
	return nil
}

func (e *execution) countNodes(root *html.Node) (int, error) {
	count := 0
	err := e.walk(root, func(*html.Node, int) error {
		count++
		return nil
	})
	return count, err
}

func (e *execution) allowGenerated(scope *html.Node, added int) error {
	count, err := e.countNodes(scope)
	if err != nil {
		return err
	}
	if added > MaxContentNodes-count {
		return fmt.Errorf("%w: generated content exceeds %d nodes", ErrTransformLimit, MaxContentNodes)
	}
	return nil
}

func selectedDescendants(e *execution, root *html.Node, target selectorTarget) ([]*html.Node, error) {
	if target.self {
		if err := e.spend(1); err != nil {
			return nil, err
		}
		return []*html.Node{root}, nil
	}
	var matches []*html.Node
	err := e.walk(root, func(node *html.Node, _ int) error {
		if node != root && node.Type == html.ElementNode && target.selector.Match(node) {
			matches = append(matches, node)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	switch target.mode {
	case "first":
		if len(matches) > 1 {
			matches = matches[:1]
		}
	case "last":
		if len(matches) > 1 {
			matches = matches[len(matches)-1:]
		}
	}
	return matches, nil
}

// outermostNodes drops nodes below another selected node and preserves input
// order. Callers pass document-order selections, so the outermost selected
// ancestor of every dropped node is kept.
func outermostNodes(nodes []*html.Node) []*html.Node {
	selected := nodeSet(nodes)
	result := make([]*html.Node, 0, len(nodes))
	for _, node := range nodes {
		contained := false
		for parent := node.Parent; parent != nil; parent = parent.Parent {
			if _, ok := selected[parent]; ok {
				contained = true
				break
			}
		}
		if !contained {
			result = append(result, node)
		}
	}
	return result
}

func nodeSet(nodes []*html.Node) map[*html.Node]struct{} {
	set := make(map[*html.Node]struct{}, len(nodes))
	for _, node := range nodes {
		set[node] = struct{}{}
	}
	return set
}

// hasSelfOrAncestorIn reports whether any node or one of its ancestors is in set.
func hasSelfOrAncestorIn(nodes []*html.Node, set map[*html.Node]struct{}) bool {
	for _, node := range nodes {
		for ; node != nil; node = node.Parent {
			if _, ok := set[node]; ok {
				return true
			}
		}
	}
	return false
}

func isAncestor(ancestor, node *html.Node) bool {
	for node != nil {
		if node == ancestor {
			return true
		}
		node = node.Parent
	}
	return false
}

func insertNodes(target *html.Node, nodes []*html.Node, position string) error {
	if len(nodes) == 0 {
		return nil
	}
	for _, node := range nodes {
		if node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
	switch position {
	case "append":
		for _, node := range nodes {
			target.AppendChild(node)
		}
	case "prepend":
		anchor := target.FirstChild
		for _, node := range nodes {
			target.InsertBefore(node, anchor)
		}
	case "before":
		if target.Parent == nil {
			return fmt.Errorf("move target is detached")
		}
		for _, node := range nodes {
			target.Parent.InsertBefore(node, target)
		}
	case "after":
		if target.Parent == nil {
			return fmt.Errorf("move target is detached")
		}
		anchor := target.NextSibling
		for _, node := range nodes {
			target.Parent.InsertBefore(node, anchor)
		}
	case "replace":
		if target.Parent == nil {
			return fmt.Errorf("move target is detached")
		}
		for _, node := range nodes {
			target.Parent.InsertBefore(node, target)
		}
		target.Parent.RemoveChild(target)
	default:
		return fmt.Errorf("unsupported insertion position %q", position)
	}
	return nil
}

func cloneNodes(nodes []*html.Node) []*html.Node {
	clones := make([]*html.Node, len(nodes))
	for i, node := range nodes {
		clones[i] = cloneNode(node)
	}
	return clones
}

func cloneNode(node *html.Node) *html.Node {
	clone := *node
	clone.Parent, clone.PrevSibling, clone.NextSibling = nil, nil, nil
	clone.FirstChild, clone.LastChild = nil, nil
	clone.Attr = append([]html.Attribute(nil), node.Attr...)
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		clone.AppendChild(cloneNode(child))
	}
	return &clone
}

func recoverNoscriptImages(e *execution, source *html.Node) ([]*html.Node, error) {
	var parsed []*html.Node
	var encoded strings.Builder
	for child := source.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			if err := e.value(child.Data); err != nil {
				return nil, err
			}
			if encoded.Len()+len(child.Data) > MaxValueBytes {
				return nil, fmt.Errorf("%w: noscript text exceeds %d bytes", ErrTransformLimit, MaxValueBytes)
			}
			encoded.WriteString(child.Data)
		}
		if err := e.walk(child, func(node *html.Node, _ int) error {
			if node.Type == html.ElementNode && (node.DataAtom == atom.Img || node.Data == "img") {
				parsed = append(parsed, cloneNode(node))
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	if len(parsed) != 0 || encoded.Len() == 0 {
		return parsed, nil
	}
	context := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := html.ParseFragment(strings.NewReader(stdhtml.UnescapeString(encoded.String())), context)
	if err != nil {
		return nil, fmt.Errorf("parse encoded noscript content: %w", err)
	}
	for _, node := range nodes {
		if err := e.walk(node, func(child *html.Node, _ int) error {
			if child.Type == html.ElementNode && (child.DataAtom == atom.Img || child.Data == "img") {
				parsed = append(parsed, cloneNode(child))
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return parsed, nil
}
