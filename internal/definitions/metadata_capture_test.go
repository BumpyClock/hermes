package definitions

import (
	"errors"
	"strings"
	"testing"
)

func TestMetadataTextCaptureValidation(t *testing.T) {
	tests := []string{
		`text_capture: {selector: .author, pattern: '(', group: 1}`,
		`text_capture: {selector: .author, pattern: '(date)', group: 0}`,
		`text_capture: {selector: .author, pattern: '(date)', group: 2}`,
		`text_capture: {selector: .author, pattern: '(date)', group: '1'}`,
		`text_capture: {selector: .author, pattern: '(date)'}`,
		`text_capture: {selector: '[', pattern: '(date)', group: 1}`,
	}
	tests = append(tests,
		"text_capture: {selector: .author, pattern: '("+strings.Repeat("a", MaxPatternBytes)+")', group: 1}",
		"text_capture: {selector: .author, pattern: '"+strings.Repeat("(a)", MaxCaptureGroups+1)+"', group: 1}",
	)
	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			dir := t.TempDir()
			write(t, dir, "site.yaml", `schema: 1
site: metadata-capture
hosts: [example.com]
metadata:
  date_published:
    - `+test+`
content:
  groups: [[article]]
`)
			snapshot, err := LoadDirectory(dir)
			var diagnostic *Error
			if snapshot != nil || !errors.As(err, &diagnostic) || diagnostic.Line == 0 || !strings.Contains(diagnostic.Field, "metadata.date_published") {
				t.Fatalf("invalid metadata capture accepted: %v", err)
			}
		})
	}
}
