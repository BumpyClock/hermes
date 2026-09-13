package definitions

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDirectoryNumericalLimits(t *testing.T) {
	for _, kind := range []string{"files", "total bytes"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			count := MaxFiles + 1
			if kind == "total bytes" {
				count = MaxTotalBytes/MaxFileBytes + 1
			}
			for i := 0; i < count; i++ {
				source := strings.Replace(strings.Replace(valid, "site: test", fmt.Sprintf("site: test%d", i), 1), "example.com", fmt.Sprintf("site%d.example.com", i), 1)
				if kind == "total bytes" {
					source += "#" + strings.Repeat(" ", MaxFileBytes-len(source)-1)
				}
				write(t, dir, fmt.Sprintf("%04d.yaml", i), source)
			}
			s, err := LoadDirectory(dir)
			if s != nil || err == nil {
				t.Fatalf("accepted excessive %s", kind)
			}
		})
	}
}

func TestUnreadableSourceAndSyntaxCause(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "site.yaml")
	write(t, dir, "site.yaml", valid)
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0600) })
	s, err := LoadDirectory(dir)
	if s != nil || !errors.Is(err, os.ErrPermission) {
		t.Fatalf("expected unreadable cause: %v", err)
	}
	if err = os.Chmod(path, 0600); err != nil {
		t.Fatal(err)
	}
	write(t, dir, "site.yaml", "schema: 1\nhosts: [unterminated\n")
	_, err = LoadDirectory(dir)
	var diagnostic *Error
	if !errors.As(err, &diagnostic) || diagnostic.Line == 0 || diagnostic.Unwrap() == nil {
		t.Fatalf("syntax diagnostic: %#v", diagnostic)
	}
}

const valid = `schema: 1
site: test
hosts: [example.com]
metadata:
  title:
    - text: h1
content:
  groups: [[article]]
`

func write(t *testing.T, dir, name, data string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadRejectsAtomically(t *testing.T) {
	tests := map[string]string{
		"empty":                   "",
		"null":                    "null",
		"unknown":                 valid + "transforms: []\n",
		"unsupported capability":  valid + "capabilities: [scalar.transforms]\n",
		"duplicate":               valid + "site: other\n",
		"wrong version type":      strings.Replace(valid, "schema: 1", "schema: \"1\"", 1),
		"unsupported schema":      strings.Replace(valid, "schema: 1", "schema: 2", 1),
		"wrong host type":         strings.Replace(valid, "[example.com]", "[42]", 1),
		"bad host":                strings.Replace(valid, "example.com", "foo.*.com", 1),
		"selector":                strings.Replace(valid, "[[article]]", "[['[']]", 1),
		"scalar content":          strings.Replace(valid, "[[article]]", "[article]", 1),
		"metadata tuple":          strings.Replace(valid, "- text: h1", "- [h1, text]", 1),
		"multiple metadata kinds": strings.Replace(valid, "- text: h1", "- text: h1\n      attribute: {selector: h1, name: title}", 1),
		"bad attribute":           strings.Replace(valid, "- text: h1", "- attribute: {selector: h1, name: 'bad name'}", 1),
		"missing attribute":       strings.Replace(valid, "- text: h1", "- attribute: {selector: h1}", 1),
		"legacy option":           valid + "  disableDefaultCleaner: true\n",
		"wrong cleaner type":      valid + "  default_cleaner: 'false'\n",
		"unsupported operation":   valid + "  transforms: []\n",
		"multiple documents":      valid + "---\n" + valid,
		"anchor":                  strings.Replace(valid, "[example.com]", "&hosts [example.com]", 1),
		"long scalar":             strings.Replace(valid, "example.com", strings.Repeat("a", MaxStringBytes+1), 1),
		"too many items":          strings.Replace(valid, "[example.com]", "["+strings.Repeat("example.com,", MaxListItems)+"example.com]", 1),
		"too many bytes":          strings.Repeat("# padding\n", MaxFileBytes/10+1),
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			write(t, dir, "a.yaml", valid)
			write(t, dir, "z.yaml", input)
			s, err := LoadDirectory(dir)
			var diagnostic *Error
			if s != nil || !errors.As(err, &diagnostic) || diagnostic.Source == "" || errors.Unwrap(err) == nil {
				t.Fatalf("snapshot=%v error=%v", s, err)
			}
		})
	}
}

func TestSourcesAndConflicts(t *testing.T) {
	for _, kind := range []string{"missing", "empty", "directory", "symlink", "conflict", "www conflict", "wildcard conflict", "duplicate site"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			switch kind {
			case "missing":
				dir = filepath.Join(dir, "absent")
			case "directory":
				if err := os.Mkdir(filepath.Join(dir, "bad.yaml"), 0700); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				write(t, dir, "source", valid)
				if err := os.Symlink(filepath.Join(dir, "source"), filepath.Join(dir, "bad.yaml")); err != nil {
					t.Fatal(err)
				}
			case "conflict", "www conflict", "wildcard conflict", "duplicate site":
				first, second := valid, strings.Replace(valid, "site: test", "site: second", 1)
				switch kind {
				case "www conflict":
					second = strings.Replace(second, "example.com", "WWW.EXAMPLE.COM", 1)
				case "wildcard conflict":
					first = strings.Replace(first, "example.com", "'*.example.com'", 1)
					second = strings.Replace(second, "example.com", "'*.example.com'", 1)
				case "duplicate site":
					second = strings.Replace(valid, "example.com", "other.com", 1)
				}
				write(t, dir, "a.yaml", first)
				write(t, dir, "b.yaml", second)
			}
			if s, err := LoadDirectory(dir); s != nil || err == nil {
				t.Fatalf("snapshot=%v err=%v", s, err)
			}
		})
	}
}

func TestMatchingAndIsolation(t *testing.T) {
	dir := t.TempDir()
	for name, hosts := range map[string]string{"exact": "[example.com]", "wild": "['*.example.com']", "deep": "['*.news.example.com']", "specific": "[a.news.example.com]"} {
		write(t, dir, name+".yaml", strings.Replace(strings.Replace(valid, "site: test", "site: "+name, 1), "[example.com]", hosts, 1))
	}
	s, err := LoadDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	for host, want := range map[string]string{"EXAMPLE.COM": "exact", "www.example.com": "exact", "a.example.com": "wild", "x.y.example.com": "wild", "news.example.com": "wild", "b.news.example.com": "deep", "a.news.example.com": "specific", "www.a.news.example.com": "specific", "badexample.com": "", "example.net": ""} {
		got := s.Match(host)
		if want == "" {
			if got != nil {
				t.Errorf("%s matched", host)
			}
			continue
		}
		if got == nil || got.Domain != want {
			t.Fatalf("%s: %v want %s", host, got, want)
		}
		got.Content.Selectors[0][0] = "changed"
		if s.Match(host).Content.Selectors[0][0] != "article" {
			t.Fatal("snapshot mutated")
		}
	}
}

func TestContextAndStructuralLimits(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "bad.yaml", valid+"  default_cleaner: invalid\n")
	_, err := LoadDirectory(dir)
	var e *Error
	if !errors.As(err, &e) || e.Site != "test" || e.Field != "content.default_cleaner" || e.Line == 0 || e.Column == 0 {
		t.Fatalf("missing context: %#v", e)
	}
	for _, input := range []string{
		strings.Repeat("[", MaxDepth+1) + "x" + strings.Repeat("]", MaxDepth+1),
		"[" + strings.Repeat("["+strings.Repeat("x,", 100)+"x],", 100) + "x]",
	} {
		write(t, dir, "bad.yaml", input)
		if _, err := LoadDirectory(dir); err == nil {
			t.Fatal("accepted excessive structure")
		}
	}
}
