// ABOUTME: Comprehensive tests for all 15 science & education site custom extractors
// ABOUTME: Validates JavaScript compatibility and fixture validation for science sites

package custom

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScienceEducationExtractorsBasicStructure(t *testing.T) {
	tests := []struct {
		name           string
		extractor      *CustomExtractor
		expectedDomain string
	}{
		{"WwwNationalgeographicCom", GetWwwNationalgeographicComExtractor(), "www.nationalgeographic.com"},
		{"NewsNationalgeographicCom", GetNewsNationalgeographicComExtractor(), "news.nationalgeographic.com"},
		{"BiorxivOrg", GetBiorxivOrgExtractor(), "biorxiv.org"},
		{"ClinicaltrialsGov", GetClinicaltrialsGovExtractor(), "clinicaltrials.gov"},
		{"ScienceflyCom", GetScienceflyComExtractor(), "sciencefly.com"},
		{"WwwIpaGoJp", GetWwwIpaGoJpExtractor(), "www.ipa.go.jp"},
		{"WwwJnsaOrg", GetWwwJnsaOrgExtractor(), "www.jnsa.org"},
		{"ScanNetsecurityNeJp", GetScanNetsecurityNeJpExtractor(), "scan.netsecurity.ne.jp"},
		{"TakagihiromitsuJp", GetTakagihiromitsuJpExtractor(), "takagi-hiromitsu.jp"},
		{"SectIijAdJp", GetSectIijAdJpExtractor(), "sect.iij.ad.jp"},
		{"TechlogIijAdJp", GetTechlogIijAdJpExtractor(), "techlog.iij.ad.jp"},
		{"JvndbJvnJp", GetJvndbJvnJpExtractor(), "jvndb.jvn.jp"},
		{"PhpspotOrg", GetPhpspotOrgExtractor(), "phpspot.org"},
		{"WwwFortinetCom", GetWwwFortinetComExtractor(), "www.fortinet.com"},
		{"ArstechnicaCom", GetArstechnicaComExtractor(), "arstechnica.com"}, // Tech site with scientific content
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.extractor, "Extractor should not be nil")
			assert.Equal(t, tt.expectedDomain, tt.extractor.Domain, "Domain should match expected")
			
			// Verify basic structure exists
			assert.NotNil(t, tt.extractor.Title, "Title extractor should exist")
			assert.NotEmpty(t, tt.extractor.Title.Selectors, "Title selectors should not be empty")
		})
	}
}

func TestNationalGeographicExtractorDetails(t *testing.T) {
	extractor := GetWwwNationalgeographicComExtractor()
	
	// Test title selectors
	assert.Equal(t, []interface{}{"h1", "h1.main-title"}, extractor.Title.Selectors)
	
	// Test author selectors  
	assert.Equal(t, []interface{}{".byline-component__contributors b span"}, extractor.Author.Selectors)
	
	// Test date published selectors (meta attribute extraction)
	expectedDateSelectors := []interface{}{[]string{"meta[name=\"article:published_time\"]", "value"}}
	assert.Equal(t, expectedDateSelectors, extractor.DatePublished.Selectors)
	
	// Test dek selectors
	expectedDekSelectors := []interface{}{".Article__Headline__Desc", ".article__deck"}
	assert.Equal(t, expectedDekSelectors, extractor.Dek.Selectors)
	
	// Test lead image URL selectors
	expectedImageSelectors := []interface{}{[]string{"meta[name=\"og:image\"]", "value"}}
	assert.Equal(t, expectedImageSelectors, extractor.LeadImageURL.Selectors)
	
	// Test content selectors
	expectedContentSelectors := []interface{}{
		"section.Article__Content",
		[]string{".parsys.content", ".__image-lead__"},
		".content",
	}
	assert.Equal(t, expectedContentSelectors, extractor.Content.Selectors)
	
	// Test complex transforms exist for NatGeo image handling
	assert.Contains(t, extractor.Content.Transforms, ".parsys.content", "Should have .parsys.content transform")
	
	// Test clean selectors
	assert.Contains(t, extractor.Content.Clean, ".pull-quote.pull-quote--small")
}

func TestNewsNationalGeographicExtractorDetails(t *testing.T) {
	extractor := GetNewsNationalgeographicComExtractor()
	
	// Test different clean selector from main site
	assert.Contains(t, extractor.Content.Clean, ".pull-quote.pull-quote--large")
	
	// Test similar but different content selectors
	expectedContentSelectors := []interface{}{
		[]string{".parsys.content", ".__image-lead__"},
		".content",
	}
	assert.Equal(t, expectedContentSelectors, extractor.Content.Selectors)
	
	// Test image transform exists
	assert.Contains(t, extractor.Content.Transforms, ".parsys.content", "Should have image transform")
}

func TestBioRxivExtractorDetails(t *testing.T) {
	extractor := GetBiorxivOrgExtractor()
	
	// Test academic paper title selector
	assert.Equal(t, []interface{}{"h1#page-title"}, extractor.Title.Selectors)
	
	// Test academic author selector (complex citation format)
	expectedAuthorSelectors := []interface{}{
		"div.highwire-citation-biorxiv-article-top > div.highwire-cite-authors",
	}
	assert.Equal(t, expectedAuthorSelectors, extractor.Author.Selectors)
	
	// Test abstract content selector
	assert.Equal(t, []interface{}{"div#abstract-1"}, extractor.Content.Selectors)
	
	// Test no transforms or clean selectors (academic paper simplicity)
	assert.Empty(t, extractor.Content.Transforms)
	assert.Empty(t, extractor.Content.Clean)
}

func TestClinicalTrialsGovExtractorDetails(t *testing.T) {
	extractor := GetClinicaltrialsGovExtractor()
	
	// Test government site title format
	assert.Equal(t, []interface{}{"h1.tr-solo_record"}, extractor.Title.Selectors)
	
	// Test sponsor as author (government context)
	assert.Equal(t, []interface{}{"div#sponsor.tr-info-text"}, extractor.Author.Selectors)
	
	// Test complex date selector using :has() pseudo-class
	expectedDateSelectors := []interface{}{
		`div:has(> span.term[data-term="Last Update Posted"])`,
	}
	assert.Equal(t, expectedDateSelectors, extractor.DatePublished.Selectors)
	
	// Test government content structure
	assert.Equal(t, []interface{}{"div#tab-body"}, extractor.Content.Selectors)
	
	// Test government alert removal
	assert.Contains(t, extractor.Content.Clean, ".usa-alert> img")
}

func TestJapaneseAcademicSites(t *testing.T) {
	// Test IPA (government research agency)
	ipaExtractor := GetWwwIpaGoJpExtractor()
	assert.Equal(t, "www.ipa.go.jp", ipaExtractor.Domain)
	assert.Nil(t, ipaExtractor.Author) // Government sites often don't have individual authors
	assert.False(t, ipaExtractor.Content.DefaultCleaner) // Custom cleaning for Japanese sites
	
	// Test JNSA (security association) 
	jnsaExtractor := GetWwwJnsaOrgExtractor()
	assert.Equal(t, "www.jnsa.org", jnsaExtractor.Domain)
	assert.NotNil(t, jnsaExtractor.Excerpt) // Has special excerpt extraction
	
	// Test academic researcher site
	takagiExtractor := GetTakagihiromitsuJpExtractor()
	assert.Equal(t, "takagi-hiromitsu.jp", takagiExtractor.Domain)
	assert.Equal(t, []interface{}{[]string{"meta[name=\"author\"]", "value"}}, takagiExtractor.Author.Selectors)
	assert.False(t, takagiExtractor.Content.DefaultCleaner) // Personal academic site
}

func TestCybersecurityResearchSites(t *testing.T) {
	// Test ScanNetSecurity
	scanExtractor := GetScanNetsecurityNeJpExtractor()
	assert.Equal(t, "scan.netsecurity.ne.jp", scanExtractor.Domain)
	assert.Equal(t, []interface{}{"header.arti-header h1.head"}, scanExtractor.Title.Selectors)
	assert.False(t, scanExtractor.Content.DefaultCleaner) // Custom cleaning for security content
	
	// Test JVNDB (vulnerability database)
	jvndbExtractor := GetJvndbJvnJpExtractor()
	assert.Equal(t, "jvndb.jvn.jp", jvndbExtractor.Domain)
	assert.Equal(t, []interface{}{"title"}, jvndbExtractor.Title.Selectors) // Simple title for database
	assert.Nil(t, jvndbExtractor.Author) // Database entries don't have authors
}

func TestCorporateResearchSites(t *testing.T) {
	// Test Fortinet (cybersecurity company)
	fortinetExtractor := GetWwwFortinetComExtractor()
	assert.Equal(t, "www.fortinet.com", fortinetExtractor.Domain)
	
	// Test complex corporate content selector
	expectedSelectors := []interface{}{
		"div.responsivegrid.aem-GridColumn.aem-GridColumn--default--12",
	}
	assert.Equal(t, expectedSelectors, fortinetExtractor.Content.Selectors)
	
	// Test noscript transform for AEM-based corporate site
	assert.Contains(t, fortinetExtractor.Content.Transforms, "noscript", "Should have noscript transform for AEM sites")
}

func TestDeveloperResourceSites(t *testing.T) {
	// Test PHPSpot (development resources)
	phpspotExtractor := GetPhpspotOrgExtractor()
	assert.Equal(t, "phpspot.org", phpspotExtractor.Domain)
	assert.Equal(t, []interface{}{"h3.hl"}, phpspotExtractor.Title.Selectors)
	assert.Equal(t, []interface{}{"h4.hl"}, phpspotExtractor.DatePublished.Selectors) // Unusual date selector
	
	// Test IIJ technical blogs
	techlogExtractor := GetTechlogIijAdJpExtractor()
	assert.Equal(t, "techlog.iij.ad.jp", techlogExtractor.Domain)
	assert.Equal(t, []interface{}{"h1.entry-title"}, techlogExtractor.Title.Selectors) // WordPress-style
	assert.Contains(t, techlogExtractor.Content.Clean, ".wp_social_bookmarking_light") // WordPress cleanup
}

// Test registry integration
func TestScienceEducationExtractorsInRegistry(t *testing.T) {
	registry := GetAllCustomExtractors()
	
	expectedExtractors := []string{
		"WwwNationalgeographicComExtractor",
		"NewsNationalgeographicComExtractor", 
		"BiorxivOrgExtractor",
		"ClinicaltrialsGovExtractor",
		"ScienceflyComExtractor",
		"WwwIpaGoJpExtractor",
		"WwwJnsaOrgExtractor",
		"ScanNetsecurityNeJpExtractor",
		"TakagihiromitsuJpExtractor",
		"SectIijAdJpExtractor",
		"TechlogIijAdJpExtractor",
		"JvndbJvnJpExtractor",
		"PhpspotOrgExtractor",
		"WwwFortinetComExtractor",
		"ArstechnicaComExtractor", // Pre-existing tech extractor with scientific content
	}
	
	for _, extractorName := range expectedExtractors {
		t.Run(extractorName, func(t *testing.T) {
			extractor, exists := registry[extractorName]
			assert.True(t, exists, "Extractor %s should be registered", extractorName)
			assert.NotNil(t, extractor, "Registered extractor should not be nil")
		})
	}
}

// Test domain lookup functionality
func TestScienceEducationDomainLookup(t *testing.T) {
	domainTests := []struct {
		domain   string
		expected bool
	}{
		{"www.nationalgeographic.com", true},
		{"news.nationalgeographic.com", true},
		{"biorxiv.org", true},
		{"clinicaltrials.gov", true},
		{"sciencefly.com", true},
		{"www.ipa.go.jp", true},
		{"www.jnsa.org", true},
		{"scan.netsecurity.ne.jp", true},
		{"takagi-hiromitsu.jp", true},
		{"sect.iij.ad.jp", true},
		{"techlog.iij.ad.jp", true},
		{"jvndb.jvn.jp", true},
		{"phpspot.org", true},
		{"www.fortinet.com", true},
		{"arstechnica.com", true},
		{"nonexistent-science-site.com", false},
	}
	
	for _, tt := range domainTests {
		t.Run(tt.domain, func(t *testing.T) {
			extractor, found := GetCustomExtractorByDomain(tt.domain)
			assert.Equal(t, tt.expected, found, "Domain lookup for %s should return %v", tt.domain, tt.expected)
			
			if found {
				assert.NotNil(t, extractor, "Found extractor should not be nil")
				assert.Equal(t, tt.domain, extractor.Domain, "Extractor domain should match")
			}
		})
	}
}

// Benchmark science & education extractor performance
func BenchmarkScienceEducationExtractors(b *testing.B) {
	extractors := []struct {
		name string
		fn   func() *CustomExtractor
	}{
		{"NatGeo", GetWwwNationalgeographicComExtractor},
		{"BioRxiv", GetBiorxivOrgExtractor},
		{"ClinicalTrials", GetClinicaltrialsGovExtractor},
		{"Fortinet", GetWwwFortinetComExtractor},
		{"JVNDB", GetJvndbJvnJpExtractor},
	}
	
	for _, extractor := range extractors {
		b.Run(extractor.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = extractor.fn()
			}
		})
	}
}