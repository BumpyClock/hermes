package custom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildDomainMapPreservesDomainPrecedence(t *testing.T) {
	first := &CustomExtractor{
		Domain:           "EXAMPLE.com",
		SupportedDomains: []string{"alias.example.com", "EXAMPLE.com"},
	}
	second := &CustomExtractor{
		Domain:           "ALIAS.example.com",
		SupportedDomains: []string{"example.COM", "other.example.com", ""},
	}
	domains := buildDomainMap(map[string]*CustomExtractor{
		"ZSecond": second,
		"AFirst":  first,
		"Nil":     nil,
		"Empty":   {},
	})

	assert.Len(t, domains, 4)
	assert.Same(t, first, domains["example.com"])
	assert.Same(t, first, domains["alias.example.com"])
	assert.Same(t, second, domains["other.example.com"])
	assert.Same(t, second, domains[""])
}
