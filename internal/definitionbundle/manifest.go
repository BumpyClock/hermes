// Package definitionbundle validates the independent, local candidate bundle protocol.
package definitionbundle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	Protocol        = 1
	Contract        = "hermes-local-definitions-v1"
	MaxManifestSize = 2 << 20
	MaxArchiveSize  = 80 << 20
	MaxFileSize     = 8 << 20
	MaxPayloadSize  = 64 << 20
	MaxEntries      = 4096
)

// Manifest binds all payload bytes and compatibility requirements, not a download URL.
type Manifest struct {
	Protocol         int           `json:"protocol"`
	Version          string        `json:"version"`
	Scope            string        `json:"scope"`
	Source           Source        `json:"source"`
	DefinitionSchema int           `json:"definition_schema"`
	Engine           Compatibility `json:"engine"`
	Archive          Archive       `json:"archive"`
	Files            []File        `json:"files"`
}

// Source distinguishes a pinned revision from inputs modified relative to it.
type Source struct {
	Repository   string `json:"repository"`
	Revision     string `json:"revision"`
	State        string `json:"state"`
	InputsSHA256 string `json:"inputs_sha256"`
}

// Compatibility is a capability contract, not an ordering of engine releases.
type Compatibility struct {
	Contract   string   `json:"contract"`
	Operations []string `json:"operations"`
	Algorithms []string `json:"algorithms"`
}

// Archive identifies the exact, uncompressed USTAR payload.
type Archive struct {
	Name   string `json:"name"`
	Format string `json:"format"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// File binds an archive entry to its contents.
type File struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

var (
	digestPattern   = regexp.MustCompile(`^[0-9a-f]{64}$`)
	revisionPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	identifier      = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,95}$`)
	component       = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*(?:\.[a-zA-Z0-9_-]+)*$`)
)

// Digest is the lowercase SHA-256 identity used throughout protocol 1.
func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// DecodeJSON rejects duplicate keys, non-exact struct field names, nulls and trailing documents.
func DecodeJSON(data []byte, target any) error {
	if len(data) > MaxManifestSize || !utf8.Valid(data) {
		return fmt.Errorf("JSON must be UTF-8 and at most %d bytes", MaxManifestSize)
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	targetType := reflect.TypeOf(target)
	if targetType == nil || targetType.Kind() != reflect.Pointer {
		return fmt.Errorf("JSON target must be a typed pointer")
	}
	if err := checkJSON(d, 0, targetType); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("expected one JSON document")
	}
	d = json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return fmt.Errorf("JSON type/field: %w", err)
	}
	return nil
}

func checkJSON(d *json.Decoder, depth int, expected reflect.Type) error {
	if depth > 20 {
		return fmt.Errorf("JSON nesting exceeds 20")
	}
	token, err := d.Token()
	if err != nil {
		return err
	}
	if token == nil {
		return fmt.Errorf("JSON null is unsupported")
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delim != '{' && delim != '[' {
		return fmt.Errorf("unexpected JSON delimiter")
	}
	for expected != nil && expected.Kind() == reflect.Pointer {
		expected = expected.Elem()
	}
	var fields map[string]reflect.Type
	var element reflect.Type
	if expected != nil {
		switch expected.Kind() {
		case reflect.Struct:
			fields = map[string]reflect.Type{}
			for i := 0; i < expected.NumField(); i++ {
				field := expected.Field(i)
				name := strings.Split(field.Tag.Get("json"), ",")[0]
				if field.PkgPath != "" || name == "-" {
					continue
				}
				if name == "" {
					name = field.Name
				}
				fields[name] = field.Type
			}
		case reflect.Map, reflect.Slice, reflect.Array:
			element = expected.Elem()
		}
	}
	seen := map[string]bool{}
	for d.More() {
		childType := element
		if delim == '{' {
			key, keyErr := d.Token()
			if keyErr != nil {
				return keyErr
			}
			name, valid := key.(string)
			if !valid || seen[name] {
				return fmt.Errorf("duplicate or invalid JSON key %q", key)
			}
			seen[name] = true
			if fields != nil {
				var exists bool
				childType, exists = fields[name]
				if !exists {
					return fmt.Errorf("unknown or non-exact JSON field %q in %s", name, expected.Name())
				}
			}
		}
		if err = checkJSON(d, depth+1, childType); err != nil {
			return err
		}
	}
	_, err = d.Token()
	return err
}

// ParseManifest validates the complete envelope before archive processing.
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := DecodeJSON(data, &m); err != nil {
		return nil, err
	}
	if m.Protocol != Protocol || m.DefinitionSchema != 1 || m.Engine.Contract != Contract {
		return nil, fmt.Errorf("unsupported bundle protocol, definition schema or engine contract")
	}
	if !identifier.MatchString(m.Version) || (m.Scope != "pilot" && m.Scope != "complete") {
		return nil, fmt.Errorf("invalid version or scope")
	}
	if m.Source.Repository != "https://github.com/BumpyClock/hermes-definitions" ||
		!revisionPattern.MatchString(m.Source.Revision) || !digestPattern.MatchString(m.Source.InputsSHA256) ||
		(m.Source.State != "pristine" && m.Source.State != "overlay") {
		return nil, fmt.Errorf("invalid source identity")
	}
	if m.Archive.Format != "ustar-v1" || !digestPattern.MatchString(m.Archive.SHA256) ||
		m.Archive.Name != "definitions-"+m.Version+"-"+m.Archive.SHA256+".tar" ||
		m.Archive.Size < 1024 || m.Archive.Size > MaxArchiveSize {
		return nil, fmt.Errorf("invalid archive identity, format or size")
	}
	if err := sortedNames(m.Engine.Operations, false); err != nil {
		return nil, fmt.Errorf("operations: %w", err)
	}
	if err := sortedNames(m.Engine.Algorithms, true); err != nil {
		return nil, fmt.Errorf("algorithms: %w", err)
	}
	if len(m.Files) == 0 || len(m.Files) > MaxEntries {
		return nil, fmt.Errorf("invalid archive entry count")
	}
	var total int64
	previous, definitions, cases := "", 0, 0
	paths := make([]string, 0, len(m.Files))
	for _, file := range m.Files {
		if !validPath(file.Path) || file.Path <= previous || file.Size <= 0 ||
			file.Size > MaxFileSize || !digestPattern.MatchString(file.SHA256) {
			return nil, fmt.Errorf("invalid or unordered file record %q", file.Path)
		}
		if strings.HasPrefix(file.Path, "definitions/") {
			definitions++
		}
		if file.Path == "conformance.json" {
			cases++
		}
		total += file.Size
		previous = file.Path
		paths = append(paths, file.Path)
	}
	if err := validatePayloadPaths(paths); err != nil {
		return nil, err
	}
	if total > MaxPayloadSize || definitions == 0 || cases != 1 {
		return nil, fmt.Errorf("payload bounds or required definitions/conformance entries missing")
	}
	return &m, nil
}

func validatePayloadPaths(paths []string) error {
	seen := map[string]string{}
	for _, name := range paths {
		if !validPath(name) {
			return fmt.Errorf("invalid payload path %q", name)
		}
		folded := strings.ToLower(name)
		if previous, exists := seen[folded]; exists {
			return fmt.Errorf("case-folded payload path collision: %q and %q", previous, name)
		}
		seen[folded] = name
	}
	return nil
}

func validPath(name string) bool {
	if len(name) > 99 || path.Clean(name) != name {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if !component.MatchString(part) {
			return false
		}
	}
	parts := strings.Split(name, "/")
	return name == "conformance.json" || name == "coverage.json" ||
		(len(parts) == 2 && parts[0] == "definitions" && strings.HasSuffix(name, ".yaml")) ||
		(len(parts) == 3 && parts[0] == "fixtures" && strings.HasSuffix(name, ".html"))
}

func sortedNames(names []string, emptyAllowed bool) error {
	if names == nil || (!emptyAllowed && len(names) == 0) || len(names) > 128 || !slices.IsSorted(names) {
		return fmt.Errorf("expected explicit sorted capability array")
	}
	for i, name := range names {
		if !identifier.MatchString(name) || (i > 0 && name == names[i-1]) {
			return fmt.Errorf("invalid or duplicate capability %q", name)
		}
	}
	return nil
}

// CheckSupport checks both declared namespaces independently of schema equality.
func (m *Manifest) CheckSupport(schema int, operations, algorithms []string) error {
	if schema != m.DefinitionSchema {
		return fmt.Errorf("engine schema %d does not support %d", schema, m.DefinitionSchema)
	}
	for _, pair := range []struct {
		kind     string
		required []string
		support  []string
	}{{"operation", m.Engine.Operations, operations}, {"algorithm", m.Engine.Algorithms, algorithms}} {
		for _, name := range pair.required {
			if !slices.Contains(pair.support, name) {
				return fmt.Errorf("unsupported required %s %q", pair.kind, name)
			}
		}
	}
	return nil
}
