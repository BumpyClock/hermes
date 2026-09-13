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

	"github.com/andybalholm/cascadia"
	"gopkg.in/yaml.v3"

	"github.com/BumpyClock/hermes/internal/extractors/custom"
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
	extractor                                    custom.CustomExtractor
	siteLine, siteColumn, hostsLine, hostsColumn int
}

// Snapshot owns validated rules; none of its storage is exported.
type Snapshot struct{ rules []rule }

// Match returns a private copy for the extraction pipeline.
func (s *Snapshot) Match(host string) *custom.CustomExtractor {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	exact := strings.TrimPrefix(host, "www.")
	best, length := -1, 0
	for i, r := range s.rules {
		for _, pattern := range r.hosts {
			if !strings.HasPrefix(pattern, "*.") {
				if exact == strings.TrimPrefix(pattern, "www.") {
					return clone(r.extractor)
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
		return clone(s.rules[best].extractor)
	}
	return nil
}

func clone(e custom.CustomExtractor) *custom.CustomExtractor {
	copyField := func(f *custom.FieldExtractor) *custom.FieldExtractor {
		if f == nil {
			return nil
		}
		v := *f
		v.Selectors = append([]custom.SelectorEntry(nil), f.Selectors...)
		return &v
	}
	e.Title, e.Author = copyField(e.Title), copyField(e.Author)
	e.DatePublished, e.LeadImageURL = copyField(e.DatePublished), copyField(e.LeadImageURL)
	if e.Content != nil {
		c := *e.Content
		c.Clean = append([]string(nil), c.Clean...)
		c.Selectors = make([]custom.ContentSelectorGroup, len(e.Content.Selectors))
		for i, group := range e.Content.Selectors {
			c.Selectors[i] = append(custom.ContentSelectorGroup(nil), group...)
		}
		e.Content = &c
	}
	return &e
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
func (p *validator) selector(n *yaml.Node, field string) (string, error) {
	s, err := p.text(n, field)
	if err != nil {
		return "", err
	}
	if _, err := cascadia.Compile(s); err != nil {
		return "", p.error(n, field, err)
	}
	return s, nil
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
		for name, n := range fields {
			var f *custom.FieldExtractor
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
func (p *validator) field(n *yaml.Node, path string) (*custom.FieldExtractor, error) {
	items, err := p.list(n, path)
	if err != nil {
		return nil, err
	}
	f := &custom.FieldExtractor{}
	for _, item := range items {
		m, err := p.mapping(item, path, "text", "attribute")
		if err != nil {
			return nil, err
		}
		if len(m) != 1 {
			return nil, p.bad(item, path, "expected exactly one text or attribute alternative")
		}
		entry := custom.SelectorEntry{}
		if text := m["text"]; text != nil {
			entry.Selector, err = p.selector(text, path+".text")
		} else {
			a, e := p.mapping(m["attribute"], path+".attribute", "selector", "name")
			if e != nil {
				return nil, e
			}
			entry.Selector, err = p.selector(a["selector"], path+".attribute.selector")
			if err != nil {
				return nil, err
			}
			entry.Attribute, err = p.text(a["name"], path+".attribute.name")
			if err == nil && !attributePattern.MatchString(entry.Attribute) {
				err = p.bad(a["name"], path+".attribute.name", "invalid attribute name")
			}
		}
		if err != nil {
			return nil, err
		}
		f.Selectors = append(f.Selectors, entry)
	}
	return f, nil
}
func (p *validator) content(n *yaml.Node) (*custom.ContentExtractor, error) {
	if n == nil {
		return nil, p.bad(nil, "content", "required field")
	}
	m, err := p.mapping(n, "content", "groups", "remove", "default_cleaner")
	if err != nil {
		return nil, err
	}
	groups, err := p.list(m["groups"], "content.groups")
	if err != nil {
		return nil, err
	}
	c := &custom.ContentExtractor{}
	for _, group := range groups {
		items, err := p.list(group, "content.groups")
		if err != nil {
			return nil, err
		}
		var selectors custom.ContentSelectorGroup
		for _, item := range items {
			s, err := p.selector(item, "content.groups")
			if err != nil {
				return nil, err
			}
			selectors = append(selectors, s)
		}
		c.Selectors = append(c.Selectors, selectors)
	}
	if remove := m["remove"]; remove != nil {
		items, err := p.list(remove, "content.remove")
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			s, err := p.selector(item, "content.remove")
			if err != nil {
				return nil, err
			}
			c.Clean = append(c.Clean, s)
		}
	}
	if enabled := m["default_cleaner"]; enabled != nil {
		if enabled.Tag != "!!bool" || (enabled.Value != "true" && enabled.Value != "false") {
			return nil, p.bad(enabled, "content.default_cleaner", "expected true or false")
		}
		c.DisableDefaultCleaner = enabled.Value == "false"
	}
	return c, nil
}
