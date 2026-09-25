// Package definitions validates local YAML into immutable extraction snapshots.
package definitions

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"gopkg.in/yaml.v3"

	"github.com/BumpyClock/hermes/internal/extractors"
)

const (
	SchemaVersion  = 1
	MaxFiles       = 1024
	MaxFileBytes   = 256 * 1024
	MaxTotalBytes  = 16 * 1024 * 1024
	MaxNodes       = 10000
	MaxDepth       = 16
	MaxListItems   = 128
	MaxStringBytes = 2048
)

// Error describes a rejected definition and preserves its underlying cause.
type Error struct {
	Source       string
	Line, Column int
	Site, Field  string
	Err          error
}

func (e *Error) Error() string {
	return fmt.Sprintf("definition %s:%d:%d site %q field %q: %v", e.Source, e.Line, e.Column, e.Site, e.Field, e.Err)
}
func (e *Error) Unwrap() error { return e.Err }

type rule struct {
	hosts                                        []string
	extractor                                    extractors.DefinitionExtractor
	siteLine, siteColumn, hostsLine, hostsColumn int
}

// Snapshot owns validated rules; Match returns read-only pointers into them.
type Snapshot struct{ rules []rule }

// Match returns the matching rule. The rule is shared with the snapshot and must be treated as read-only.
func (s *Snapshot) Match(host string) *extractors.DefinitionExtractor {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	exact := strings.TrimPrefix(host, "www.")
	best, length := -1, 0
	for i, r := range s.rules {
		for _, pattern := range r.hosts {
			if !strings.HasPrefix(pattern, "*.") {
				if exact == strings.TrimPrefix(pattern, "www.") {
					return &s.rules[i].extractor
				}
			} else {
				suffix := pattern[1:]
				if strings.HasSuffix(host, suffix) && len(host) > len(suffix) && len(suffix) > length {
					best, length = i, len(suffix)
				}
			}
		}
	}
	if best >= 0 {
		return &s.rules[best].extractor
	}
	return nil
}

// LoadDirectory reads only immediate .yaml/.yml regular files. Acceptance is atomic.
func LoadDirectory(dir string) (*Snapshot, error) {
	fail := func(err error) (*Snapshot, error) { return nil, &Error{Source: dir, Err: err} }
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fail(err)
	}
	s := &Snapshot{}
	owners := map[string]string{}
	sites := map[string]bool{}
	total, files := 0, 0
	for _, entry := range entries {
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		files++
		if files > MaxFiles {
			return fail(fmt.Errorf("file count exceeds %d", MaxFiles))
		}
		source := filepath.Join(dir, entry.Name())
		r, size, err := loadFile(source, MaxTotalBytes-total)
		if err != nil {
			return nil, err
		}
		total += size
		if sites[r.extractor.Domain] {
			return nil, &Error{Source: source, Line: r.siteLine, Column: r.siteColumn, Site: r.extractor.Domain, Field: "site", Err: fmt.Errorf("duplicate site identifier")}
		}
		sites[r.extractor.Domain] = true
		for _, host := range r.hosts {
			key := host
			if !strings.HasPrefix(key, "*.") {
				key = strings.TrimPrefix(key, "www.")
			}
			if owner, ok := owners[key]; ok && owner != r.extractor.Domain {
				return nil, &Error{Source: source, Line: r.hostsLine, Column: r.hostsColumn, Site: r.extractor.Domain, Field: "hosts", Err: fmt.Errorf("host %q conflicts with site %q", host, owner)}
			}
			owners[key] = r.extractor.Domain
		}
		s.rules = append(s.rules, r)
	}
	if files == 0 {
		return fail(fmt.Errorf("no YAML definitions"))
	}
	return s, nil
}

func loadFile(source string, remaining int) (rule, int, error) {
	p := validator{source: source}
	info, err := os.Lstat(source)
	if err != nil {
		return rule{}, 0, p.error(nil, "", err)
	}
	if !info.Mode().IsRegular() {
		return rule{}, 0, p.error(nil, "", fmt.Errorf("definition must be a regular file, not a symlink or directory"))
	}
	//nolint:gosec // Explicit local configuration paths are the loader's input contract.
	f, err := os.Open(source)
	if err != nil {
		return rule{}, 0, p.error(nil, "", err)
	}
	limit := min(MaxFileBytes, remaining)
	data, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	err = errors.Join(err, f.Close())
	if err != nil {
		return rule{}, 0, p.error(nil, "", err)
	}
	if len(data) > limit {
		return rule{}, 0, p.error(nil, "", fmt.Errorf("definition byte limit exceeded (file %d, total %d)", MaxFileBytes, MaxTotalBytes))
	}
	var root yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err = decoder.Decode(&root); err != nil {
		return rule{}, 0, p.error(nil, "", err)
	}
	var extra yaml.Node
	if err = decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return rule{}, 0, p.error(&extra, "", err)
		}
		return rule{}, 0, p.bad(&extra, "", "expected exactly one YAML document")
	}
	if len(root.Content) != 1 {
		return rule{}, 0, p.error(&root, "", fmt.Errorf("empty document"))
	}
	if root.Content[0].Kind == yaml.MappingNode {
		items := root.Content[0].Content
		for i := 0; i+1 < len(items); i += 2 {
			if items[i].Value == "site" && items[i+1].Tag == "!!str" {
				p.site = items[i+1].Value
				break
			}
		}
	}
	if err = p.check(root.Content[0], "", 0); err != nil {
		return rule{}, 0, err
	}
	r, err := p.rule(root.Content[0])
	return r, len(data), err
}

type validator struct {
	source, site string
	nodes        int
}

var yamlLine = regexp.MustCompile(`(?:^| )line ([0-9]+):`)

func (p *validator) error(n *yaml.Node, field string, err error) error {
	e := &Error{Source: p.source, Site: p.site, Field: field, Err: err}
	if n != nil {
		e.Line, e.Column = n.Line, n.Column
	}
	if e.Line == 0 {
		if match := yamlLine.FindStringSubmatch(err.Error()); len(match) == 2 {
			e.Line, _ = strconv.Atoi(match[1])
		}
	}
	return e
}
func (p *validator) bad(n *yaml.Node, field, message string) error {
	return p.error(n, field, fmt.Errorf("%s", message))
}
func (p *validator) check(n *yaml.Node, field string, depth int) error {
	p.nodes++
	if depth > MaxDepth || p.nodes > MaxNodes {
		return p.bad(n, field, "YAML structural limit exceeded")
	}
	if n.Anchor != "" || n.Kind == yaml.AliasNode {
		return p.bad(n, field, "anchors and aliases are unsupported")
	}
	if n.Kind == yaml.ScalarNode && len(n.Value) > MaxStringBytes {
		return p.bad(n, field, "scalar byte limit exceeded")
	}
	if n.Kind == yaml.SequenceNode && len(n.Content) > MaxListItems {
		return p.bad(n, field, "list item limit exceeded")
	}
	seen := map[string]bool{}
	for i, child := range n.Content {
		path := field
		if n.Kind == yaml.MappingNode {
			if i%2 == 0 {
				if child.Tag != "!!str" || child.Kind != yaml.ScalarNode {
					return p.bad(child, field, "mapping keys must be strings")
				}
				if seen[child.Value] {
					return p.bad(child, field+"."+child.Value, "duplicate key")
				}
				seen[child.Value] = true
			}
			path = strings.TrimPrefix(field+"."+n.Content[i-i%2].Value, ".")
		}
		if err := p.check(child, path, depth+1); err != nil {
			return err
		}
	}
	return nil
}
func (p *validator) mapping(n *yaml.Node, field string, allowed ...string) (map[string]*yaml.Node, error) {
	if n.Kind != yaml.MappingNode || n.Tag != "!!map" {
		return nil, p.bad(n, field, "expected mapping")
	}
	m := map[string]*yaml.Node{}
	for i := 0; i < len(n.Content); i += 2 {
		k := n.Content[i]
		ok := false
		for _, name := range allowed {
			if k.Value == name {
				ok = true
			}
		}
		if !ok {
			return nil, p.bad(k, strings.TrimPrefix(field+"."+k.Value, "."), "unsupported field")
		}
		m[k.Value] = n.Content[i+1]
	}
	return m, nil
}
func (p *validator) text(n *yaml.Node, field string) (string, error) {
	if n == nil {
		return "", p.bad(n, field, "required field")
	}
	if n.Kind != yaml.ScalarNode || n.Tag != "!!str" || strings.TrimSpace(n.Value) == "" {
		return "", p.bad(n, field, "expected nonempty string")
	}
	return n.Value, nil
}
func (p *validator) list(n *yaml.Node, field string) ([]*yaml.Node, error) {
	if n == nil || n.Kind != yaml.SequenceNode || n.Tag != "!!seq" || len(n.Content) == 0 {
		return nil, p.bad(n, field, "expected nonempty sequence")
	}
	return n.Content, nil
}
func (p *validator) selector(n *yaml.Node, field string) (string, goquery.Matcher, error) {
	s, err := p.text(n, field)
	if err != nil {
		return "", nil, err
	}
	matcher, err := cascadia.Compile(s)
	if err != nil {
		return "", nil, p.error(n, field, err)
	}
	return s, matcher, nil
}

var hostPattern = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
var attributePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.:-]*$`)

func (p *validator) rule(n *yaml.Node) (rule, error) {
	r := rule{}
	m, err := p.mapping(n, "", "schema", "site", "hosts", "metadata", "content")
	if err != nil {
		return r, err
	}
	p.site, err = p.text(m["site"], "site")
	if err != nil {
		return r, err
	}
	r.extractor.Domain = p.site
	r.siteLine, r.siteColumn = m["site"].Line, m["site"].Column
	version := m["schema"]
	if version == nil || version.Tag != "!!int" || version.Value != "1" {
		return r, p.bad(version, "schema", "unsupported schema; expected integer 1")
	}
	hosts, err := p.list(m["hosts"], "hosts")
	if err != nil {
		return r, err
	}
	r.hostsLine, r.hostsColumn = m["hosts"].Line, m["hosts"].Column
	for _, n := range hosts {
		var host string
		host, err = p.text(n, "hosts")
		if err != nil {
			return r, err
		}
		host = strings.ToLower(host)
		base := strings.TrimPrefix(host, "*.")
		if len(base) > 253 || !hostPattern.MatchString(base) {
			return r, p.bad(n, "hosts", "expected DNS hostname or explicit *.hostname")
		}
		r.hosts = append(r.hosts, host)
	}
	if metadata := m["metadata"]; metadata != nil {
		var fields map[string]*yaml.Node
		fields, err = p.mapping(metadata, "metadata", "title", "author", "date_published", "lead_image_url")
		if err != nil {
			return r, err
		}
		for _, name := range []string{"title", "author", "date_published", "lead_image_url"} {
			n := fields[name]
			if n == nil {
				continue
			}
			var f *extractors.FieldExtractor
			f, err = p.field(n, "metadata."+name)
			if err != nil {
				return r, err
			}
			switch name {
			case "title":
				r.extractor.Title = f
			case "author":
				r.extractor.Author = f
			case "date_published":
				r.extractor.DatePublished = f
			case "lead_image_url":
				r.extractor.LeadImageURL = f
			}
		}
	}
	r.extractor.Content, err = p.content(m["content"])
	return r, err
}
func (p *validator) field(n *yaml.Node, path string) (*extractors.FieldExtractor, error) {
	items, err := p.list(n, path)
	if err != nil {
		return nil, err
	}
	f := &extractors.FieldExtractor{}
	for _, item := range items {
		m, err := p.mapping(item, path, "text", "attribute", "text_capture")
		if err != nil {
			return nil, err
		}
		if len(m) != 1 {
			return nil, p.bad(item, path, "expected exactly one text, attribute, or text_capture alternative")
		}
		entry := extractors.SelectorEntry{}
		if text := m["text"]; text != nil {
			entry.Selector, entry.Matcher, err = p.selector(text, path+".text")
		} else if attribute := m["attribute"]; attribute != nil {
			a, e := p.mapping(attribute, path+".attribute", "selector", "name")
			if e != nil {
				return nil, e
			}
			entry.Selector, entry.Matcher, err = p.selector(a["selector"], path+".attribute.selector")
			if err != nil {
				return nil, err
			}
			entry.Attribute, err = p.text(a["name"], path+".attribute.name")
			if err == nil && !attributePattern.MatchString(entry.Attribute) {
				err = p.bad(a["name"], path+".attribute.name", "invalid attribute name")
			}
		} else {
			entry, err = p.textCapture(m["text_capture"], path+".text_capture")
		}
		if err != nil {
			return nil, err
		}
		entry.Matcher = goquery.SingleMatcher(entry.Matcher)
		f.Selectors = append(f.Selectors, entry)
	}
	return f, nil
}

func (p *validator) textCapture(n *yaml.Node, path string) (extractors.SelectorEntry, error) {
	a := p.arguments(n, path, "selector", "pattern", "group")
	entry := extractors.SelectorEntry{Selector: a.string("selector", false)}
	if a.err != nil {
		return entry, a.err
	}
	matcher, err := cascadia.Compile(entry.Selector)
	if err != nil {
		return entry, p.error(a.fields["selector"], path+".selector", err)
	}
	entry.Matcher = matcher
	compiled, index := a.capturePattern()
	if a.err != nil {
		return entry, a.err
	}
	entry.Capture = &extractors.TextCapture{
		Pattern: compiled, Group: index, MaxBytes: MaxValueBytes, MaxNodes: MaxContentNodes, MaxDepth: MaxContentDepth,
	}
	return entry, nil
}
func (p *validator) content(n *yaml.Node) (*extractors.ContentExtractor, error) {
	if n == nil {
		return nil, p.bad(nil, "content", "required field")
	}
	m, err := p.mapping(n, "content", "groups", "remove", "preserve", "default_cleaner", "transforms")
	if err != nil {
		return nil, err
	}
	groups, err := p.list(m["groups"], "content.groups")
	if err != nil {
		return nil, err
	}
	c := &extractors.ContentExtractor{}
	for _, group := range groups {
		items, groupErr := p.list(group, "content.groups")
		if groupErr != nil {
			return nil, groupErr
		}
		var selectors []string
		for _, item := range items {
			s, _, selectorErr := p.selector(item, "content.groups")
			if selectorErr != nil {
				return nil, selectorErr
			}
			selectors = append(selectors, s)
		}
		matcher, selectorErr := cascadia.Compile(strings.Join(selectors, ","))
		if selectorErr != nil {
			return nil, p.error(group, "content.groups", selectorErr)
		}
		c.Selectors = append(c.Selectors, matcher)
	}
	if remove := m["remove"]; remove != nil {
		items, removeErr := p.list(remove, "content.remove")
		if removeErr != nil {
			return nil, removeErr
		}
		for _, item := range items {
			_, matcher, selectorErr := p.selector(item, "content.remove")
			if selectorErr != nil {
				return nil, selectorErr
			}
			c.Clean = append(c.Clean, matcher)
		}
	}
	if preserve := m["preserve"]; preserve != nil {
		items, preserveErr := p.list(preserve, "content.preserve")
		if preserveErr != nil {
			return nil, preserveErr
		}
		for _, item := range items {
			_, matcher, selectorErr := p.selector(item, "content.preserve")
			if selectorErr != nil {
				return nil, selectorErr
			}
			c.Preserve = append(c.Preserve, matcher)
		}
	}
	if enabled := m["default_cleaner"]; enabled != nil {
		if enabled.Tag != "!!bool" || (enabled.Value != "true" && enabled.Value != "false") {
			return nil, p.bad(enabled, "content.default_cleaner", "expected true or false")
		}
		c.DisableDefaultCleaner = enabled.Value == "false"
	}
	if transforms := m["transforms"]; transforms != nil {
		c.OrderedTransforms, err = p.transforms(transforms)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}
