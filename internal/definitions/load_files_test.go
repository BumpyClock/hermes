package definitions

import (
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// loadBoth loads files from a directory and from memory.
func loadBoth(t *testing.T, files map[string]string) (dir string, fromDir, fromMemory *Snapshot, dirErr, memoryErr error) {
	t.Helper()
	dir = t.TempDir()
	memory := make(map[string][]byte, len(files))
	for name, data := range files {
		write(t, dir, name, data)
		memory[name] = []byte(data)
	}
	fromDir, dirErr = LoadDirectory(dir)
	fromMemory, memoryErr = LoadFiles(memory)
	return dir, fromDir, fromMemory, dirErr, memoryErr
}

// equalIgnoringFuncs is reflect.DeepEqual, except that two non-nil funcs are
// equal. Compiled cascadia selectors are closures, which DeepEqual never equates.
func equalIgnoringFuncs(a, b reflect.Value) bool {
	if a.IsValid() != b.IsValid() || (a.IsValid() && a.Type() != b.Type()) {
		return false
	}
	if !a.IsValid() {
		return true
	}
	switch a.Kind() {
	case reflect.Func:
		return a.IsNil() == b.IsNil()
	case reflect.Pointer, reflect.Interface:
		if a.IsNil() || b.IsNil() {
			return a.IsNil() == b.IsNil()
		}
		return equalIgnoringFuncs(a.Elem(), b.Elem())
	case reflect.Struct:
		for i := range a.NumField() {
			if !equalIgnoringFuncs(a.Field(i), b.Field(i)) {
				return false
			}
		}
		return true
	case reflect.Slice, reflect.Array:
		if a.Kind() == reflect.Slice && a.IsNil() != b.IsNil() {
			return false
		}
		if a.Len() != b.Len() {
			return false
		}
		for i := range a.Len() {
			if !equalIgnoringFuncs(a.Index(i), b.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Map:
		if a.IsNil() != b.IsNil() || a.Len() != b.Len() {
			return false
		}
		for iter := a.MapRange(); iter.Next(); {
			if !equalIgnoringFuncs(iter.Value(), b.MapIndex(iter.Key())) {
				return false
			}
		}
		return true
	case reflect.String:
		return a.String() == b.String()
	case reflect.Bool:
		return a.Bool() == b.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return a.Int() == b.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return a.Uint() == b.Uint()
	case reflect.Float32, reflect.Float64:
		return a.Float() == b.Float()
	default:
		panic(fmt.Sprintf("equalIgnoringFuncs: unsupported kind %s", a.Kind()))
	}
}

func TestLoadFilesMatchesLoadDirectory(t *testing.T) {
	wild := strings.Replace(strings.Replace(valid, "site: test", "site: wild", 1), "[example.com]", "['*.example.net']", 1)
	upper := strings.Replace(strings.Replace(valid, "site: test", "site: upper", 1), "[example.com]", "[upper.example.org]", 1)
	files := map[string]string{
		"features.yaml": featureDefinition,
		"plain.yml":     plainDefinition,
		"upper.YAML":    upper,
		"wild.yaml":     wild,
		// Neither loader reads names without a .yaml or .yml extension.
		"notes.txt":    "not: [yaml",
		"fixture.html": "<html>",
		"yaml":         "not: [yaml",
	}
	dir, fromDir, fromMemory, dirErr, memoryErr := loadBoth(t, files)
	if dirErr != nil || memoryErr != nil {
		t.Fatalf("directory error %v, memory error %v", dirErr, memoryErr)
	}
	if len(fromDir.rules) != 4 || len(fromMemory.rules) != len(fromDir.rules) {
		t.Fatalf("rule counts: directory %d, memory %d", len(fromDir.rules), len(fromMemory.rules))
	}
	// Transform programs keep their diagnostic source: the file path when loaded
	// from a directory and the key when loaded from memory. Nothing else may differ.
	for i, r := range fromDir.rules {
		dirProgram, _ := r.extractor.Content.OrderedTransforms.(*program)
		memoryProgram, _ := fromMemory.rules[i].extractor.Content.OrderedTransforms.(*program)
		if dirProgram == nil || memoryProgram == nil {
			if dirProgram != memoryProgram {
				t.Fatalf("rule %d transform presence differs", i)
			}
			continue
		}
		if dirProgram.source != filepath.Join(dir, r.source) || memoryProgram.source != r.source {
			t.Fatalf("transform sources: directory %q, memory %q", dirProgram.source, memoryProgram.source)
		}
		dirProgram.source = memoryProgram.source
	}
	for i := range fromDir.rules {
		if !equalIgnoringFuncs(reflect.ValueOf(fromDir.rules[i]), reflect.ValueOf(fromMemory.rules[i])) {
			t.Errorf("rule %d (%s) differs between loaders", i, fromDir.rules[i].source)
		}
	}
	// The comparison must distinguish rules, or the parity check above is vacuous.
	if equalIgnoringFuncs(reflect.ValueOf(fromDir.rules[0]), reflect.ValueOf(fromMemory.rules[1])) {
		t.Fatal("comparison does not distinguish different rules")
	}
	wantSites := map[string]string{"features.yaml": "features", "plain.yml": "plain", "upper.YAML": "upper", "wild.yaml": "wild"}
	if sites := fromMemory.Sites(); !maps.Equal(sites, wantSites) || !maps.Equal(fromDir.Sites(), wantSites) {
		t.Fatalf("sites: memory %v, directory %v, want %v", sites, fromDir.Sites(), wantSites)
	}
	dirOperations, dirAlgorithms := fromDir.UsedCapabilities()
	memoryOperations, memoryAlgorithms := fromMemory.UsedCapabilities()
	if !slices.Equal(dirOperations, memoryOperations) || !slices.Equal(dirAlgorithms, memoryAlgorithms) {
		t.Fatalf("capabilities: directory %v %v, memory %v %v", dirOperations, dirAlgorithms, memoryOperations, memoryAlgorithms)
	}
	for _, host := range []string{"example.com", "www.example.com", "a.example.net", "upper.example.org", "example.org"} {
		if got, want := fromMemory.Match(host), fromDir.Match(host); (got == nil) != (want == nil) || (got != nil && got.Domain != want.Domain) {
			t.Errorf("Match(%q): memory %v, directory %v", host, got, want)
		}
	}
}

// numbered returns count valid definitions with distinct sites and hosts,
// each padded to size bytes when size is nonzero.
func numbered(count, size int) map[string]string {
	files := make(map[string]string, count)
	for i := range count {
		source := strings.Replace(strings.Replace(valid, "site: test", fmt.Sprintf("site: test%d", i), 1), "example.com", fmt.Sprintf("site%d.example.com", i), 1)
		if size > 0 {
			source += "#" + strings.Repeat(" ", size-len(source)-1)
		}
		files[fmt.Sprintf("%04d.yaml", i)] = source
	}
	return files
}

func TestLoadFilesRejectsLikeLoadDirectory(t *testing.T) {
	second := strings.Replace(valid, "site: test", "site: second", 1)
	tests := map[string]struct {
		files map[string]string
		// setLevel errors name the whole set instead of one file.
		setLevel bool
		message  string
	}{
		"too many files": {numbered(MaxFiles+1, 0), true, fmt.Sprintf("file count exceeds %d", MaxFiles)},
		"oversized file": {map[string]string{"a.yaml": valid, "z.yaml": valid + "#" + strings.Repeat(" ", MaxFileBytes-len(valid))}, false,
			fmt.Sprintf("definition byte limit exceeded (file %d, total %d)", MaxFileBytes, MaxTotalBytes)},
		"total size": {numbered(MaxTotalBytes/MaxFileBytes+1, MaxFileBytes), false,
			fmt.Sprintf("definition byte limit exceeded (file %d, total %d)", MaxFileBytes, MaxTotalBytes)},
		"no YAML":        {map[string]string{"notes.txt": valid, "site.yaml.bak": valid}, true, "no YAML definitions"},
		"syntax":         {map[string]string{"a.yaml": valid, "z.yaml": "schema: 1\nhosts: [unterminated\n"}, false, ""},
		"schema":         {map[string]string{"a.yaml": valid, "z.yaml": strings.Replace(second, "schema: 1", "schema: 2", 1)}, false, "unsupported schema; expected integer 1"},
		"duplicate site": {map[string]string{"a.yaml": valid, "b.yml": strings.Replace(valid, "example.com", "other.com", 1)}, false, "duplicate site identifier"},
		"host conflict":  {map[string]string{"a.yaml": valid, "b.yaml": strings.Replace(second, "example.com", "WWW.EXAMPLE.COM", 1)}, false, `host "www.example.com" conflicts with site "test"`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			dir, fromDir, fromMemory, dirErr, memoryErr := loadBoth(t, test.files)
			var dirDiagnostic, memoryDiagnostic *Error
			if fromDir != nil || fromMemory != nil || !errors.As(dirErr, &dirDiagnostic) || !errors.As(memoryErr, &memoryDiagnostic) {
				t.Fatalf("directory %v, memory %v", dirErr, memoryErr)
			}
			if test.message != "" && memoryDiagnostic.Err.Error() != test.message {
				t.Fatalf("memory error %q, want %q", memoryDiagnostic.Err, test.message)
			}
			wantSource := filepath.Base(dirDiagnostic.Source)
			if test.setLevel {
				wantSource = fileSetSource
				if dirDiagnostic.Source != dir {
					t.Fatalf("directory error source %q, want %q", dirDiagnostic.Source, dir)
				}
			}
			if memoryDiagnostic.Source != wantSource {
				t.Fatalf("memory error source %q, want %q", memoryDiagnostic.Source, wantSource)
			}
			dirDiagnostic.Source = memoryDiagnostic.Source
			if dirDiagnostic.Error() != memoryDiagnostic.Error() {
				t.Fatalf("errors differ beyond source:\n directory %v\n memory    %v", dirDiagnostic, memoryDiagnostic)
			}
		})
	}
}

func TestLoadFilesSelectsOnlyBaseNamedYAML(t *testing.T) {
	for _, files := range []map[string][]byte{nil, {"notes.txt": []byte(valid), "yaml": []byte(valid)}} {
		_, err := LoadFiles(files)
		var diagnostic *Error
		if !errors.As(err, &diagnostic) || diagnostic.Source != fileSetSource || diagnostic.Err.Error() != "no YAML definitions" {
			t.Fatalf("LoadFiles(%v) error = %v, want no YAML definitions", slices.Collect(maps.Keys(files)), err)
		}
	}
	// A path-like key without a YAML extension is ignored like any other non-YAML name.
	snapshot, err := LoadFiles(map[string][]byte{"site.yaml": []byte(valid), "fixtures/site/page.html": []byte("<html>")})
	if err != nil || !maps.Equal(snapshot.Sites(), map[string]string{"site.yaml": "test"}) {
		t.Fatalf("snapshot %v, error %v", snapshot, err)
	}
	for _, name := range []string{"definitions/site.yaml", `definitions\site.yaml`, "/site.yml"} {
		snapshot, err := LoadFiles(map[string][]byte{name: []byte(valid)})
		var diagnostic *Error
		if snapshot != nil || !errors.As(err, &diagnostic) || diagnostic.Source != name || diagnostic.Err.Error() != "definition name must be a base file name" {
			t.Fatalf("LoadFiles accepted path-like key %q: %v", name, err)
		}
	}
}
