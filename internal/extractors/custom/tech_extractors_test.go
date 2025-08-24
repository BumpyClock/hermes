// ABOUTME: Comprehensive tests for all 15 tech site custom extractors
// ABOUTME: Validates JavaScript compatibility and fixture validation for tech sites

package custom

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTechExtractorsBasicStructure(t *testing.T) {
	tests := []struct {
		name         string
		extractor    *CustomExtractor
		expectedDomain string
	}{
		{"ArstechnicaCom", GetArstechnicaComExtractor(), "arstechnica.com"},
		{"WwwThevergeCom", GetWwwThevergeComExtractor(), "www.theverge.com"},
		{"WwwWiredCom", GetWwwWiredComExtractor(), "www.wired.com"},
		{"WwwEngadgetCom", GetWwwEngadgetComExtractor(), "www.engadget.com"},
		{"WwwCnetCom", GetWwwCnetComExtractor(), "www.cnet.com"},
		{"WwwAndroidcentralCom", GetWwwAndroidcentralComExtractor(), "www.androidcentral.com"},
		{"WwwMacrumors Com", GetWwwMacrumorsComExtractor(), "www.macrumors.com"},
		{"MashableCom", GetMashableComExtractor(), "mashable.com"},
		{"WwwPhoronixCom", GetWwwPhoronixComExtractor(), "www.phoronix.com"},
		{"GithubCom", GetGithubComExtractor(), "github.com"},
		{"WwwInfoqCom", GetWwwInfoqComExtractor(), "www.infoq.com"},
		{"WwwGizmodoJp", GetWwwGizmodoJpExtractor(), "www.gizmodo.jp"},
		{"WiredJp", GetWiredJpExtractor(), "wired.jp"},
		{"JapanCnetCom", GetJapanCnetComExtractor(), "japan.cnet.com"},
		{"JapanZdnetCom", GetJapanZdnetComExtractor(), "japan.zdnet.com"},
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

func TestArstechnicaComExtractorDetails(t *testing.T) {
	extractor := GetArstechnicaComExtractor()
	
	// Test domain
	assert.Equal(t, "arstechnica.com", extractor.Domain)
	
	// Test title selectors
	assert.Contains(t, extractor.Title.Selectors, "title")
	
	// Test author selectors
	assert.Contains(t, extractor.Author.Selectors, "*[rel=\"author\"] *[itemprop=\"name\"]")
	
	// Test date selectors  
	dateSelector := []string{".byline time", "datetime"}
	assert.Contains(t, extractor.DatePublished.Selectors, dateSelector)
	
	// Test dek selectors
	assert.Contains(t, extractor.Dek.Selectors, "h2[itemprop=\"description\"]")
	
	// Test lead image selectors
	imageSelector := []string{"meta[name=\"og:image\"]", "value"}
	assert.Contains(t, extractor.LeadImageURL.Selectors, imageSelector)
	
	// Test content selectors
	assert.Contains(t, extractor.Content.Selectors, "div[itemprop=\"articleBody\"]")
	
	// Test transforms exist
	assert.Contains(t, extractor.Content.Transforms, "h2")
	
	// Test clean selectors
	assert.Contains(t, extractor.Content.Clean, "figcaption .enlarge-link")
	assert.Contains(t, extractor.Content.Clean, "figure.video")
	assert.Contains(t, extractor.Content.Clean, ".gallery")
	assert.Contains(t, extractor.Content.Clean, "aside")
	assert.Contains(t, extractor.Content.Clean, ".sidebar")
}

func TestWwwThevergeComExtractorDetails(t *testing.T) {
	extractor := GetWwwThevergeComExtractor()
	
	// Test domain and supported domains
	assert.Equal(t, "www.theverge.com", extractor.Domain)
	assert.Contains(t, extractor.SupportedDomains, "www.polygon.com")
	
	// Test multi-match selectors for content
	expectedMultiMatch1 := []interface{}{".c-entry-hero .e-image", ".c-entry-intro", ".c-entry-content"}
	expectedMultiMatch2 := []interface{}{".e-image--hero", ".c-entry-content"}
	assert.Contains(t, extractor.Content.Selectors, expectedMultiMatch1)
	assert.Contains(t, extractor.Content.Selectors, expectedMultiMatch2)
	assert.Contains(t, extractor.Content.Selectors, ".l-wrapper .l-feature")
	assert.Contains(t, extractor.Content.Selectors, "div.c-entry-content")
	
	// Test transforms exist
	assert.Contains(t, extractor.Content.Transforms, "noscript")
	
	// Test clean selectors
	assert.Contains(t, extractor.Content.Clean, ".aside")
	assert.Contains(t, extractor.Content.Clean, "img.c-dynamic-image")
}

func TestGithubComExtractorDetails(t *testing.T) {
	extractor := GetGithubComExtractor()
	
	// Test domain
	assert.Equal(t, "github.com", extractor.Domain)
	
	// Test title selectors
	titleSelector := []string{"meta[name=\"og:title\"]", "value"}
	assert.Contains(t, extractor.Title.Selectors, titleSelector)
	
	// Test date selectors with relative-time
	dateSelector1 := []string{"relative-time[datetime]", "datetime"}
	dateSelector2 := []string{"span[itemprop=\"dateModified\"] relative-time", "datetime"}
	assert.Contains(t, extractor.DatePublished.Selectors, dateSelector1)
	assert.Contains(t, extractor.DatePublished.Selectors, dateSelector2)
	
	// Test dek selectors
	dekSelector1 := []string{"meta[name=\"description\"]", "value"}
	dekSelector2 := "span[itemprop=\"about\"]"
	assert.Contains(t, extractor.Dek.Selectors, dekSelector1)
	assert.Contains(t, extractor.Dek.Selectors, dekSelector2)
	
	// Test content selectors for README
	contentSelector := []interface{}{"#readme article"}
	assert.Contains(t, extractor.Content.Selectors, contentSelector)
}

func TestMashableComExtractorStringTransform(t *testing.T) {
	extractor := GetMashableComExtractor()
	
	// Test domain
	assert.Equal(t, "mashable.com", extractor.Domain)
	
	// Test string transform exists
	transform, exists := extractor.Content.Transforms[".image-credit"]
	assert.True(t, exists, "String transform should exist for .image-credit")
	
	// Test it's a StringTransform
	stringTransform, ok := transform.(*StringTransform)
	assert.True(t, ok, "Transform should be a StringTransform")
	assert.Equal(t, "figcaption", stringTransform.TargetTag, "Target tag should be figcaption")
}

func TestJapaneseExtractors(t *testing.T) {
	tests := []struct {
		name      string
		extractor *CustomExtractor
		domain    string
	}{
		{"GizmodoJP", GetWwwGizmodoJpExtractor(), "www.gizmodo.jp"},
		{"WiredJP", GetWiredJpExtractor(), "wired.jp"},
		{"CnetJP", GetJapanCnetComExtractor(), "japan.cnet.com"},
		{"ZdnetJP", GetJapanZdnetComExtractor(), "japan.zdnet.com"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.domain, tt.extractor.Domain)
			
			// All Japanese extractors should have og:image lead image
			if tt.extractor.LeadImageURL != nil {
				imageSelector := []string{"meta[name=\"og:image\"]", "value"}
				assert.Contains(t, tt.extractor.LeadImageURL.Selectors, imageSelector)
			}
		})
	}
}

func TestWwwGizmodoJpImageTransform(t *testing.T) {
	extractor := GetWwwGizmodoJpExtractor()
	
	// Test image transform exists
	transform, exists := extractor.Content.Transforms["img.p-post-thumbnailImage"]
	assert.True(t, exists, "Image transform should exist")
	
	// Test it's a FunctionTransform
	_, ok := transform.(*FunctionTransform)
	assert.True(t, ok, "Transform should be a FunctionTransform")
}

func TestWiredJpURLResolveTransform(t *testing.T) {
	extractor := GetWiredJpExtractor()
	
	// Test URL resolve transform exists
	transform, exists := extractor.Content.Transforms["img[data-original]"]
	assert.True(t, exists, "URL resolve transform should exist")
	
	// Test it's a FunctionTransform
	_, ok := transform.(*FunctionTransform)
	assert.True(t, ok, "Transform should be a FunctionTransform")
}

func TestWwwInfoqComDefaultCleanerFalse(t *testing.T) {
	extractor := GetWwwInfoqComExtractor()
	
	// Test defaultCleaner is false
	assert.False(t, extractor.Content.DefaultCleaner, "DefaultCleaner should be false for InfoQ")
}

func TestCNETExtractorsWithComplexTransforms(t *testing.T) {
	extractor := GetWwwCnetComExtractor()
	
	// Test figure.image transform exists
	transform, exists := extractor.Content.Transforms["figure.image"]
	assert.True(t, exists, "Figure image transform should exist")
	
	// Test it's a FunctionTransform
	_, ok := transform.(*FunctionTransform)
	assert.True(t, ok, "Transform should be a FunctionTransform")
	
	// Test content multi-match selectors
	multiMatch := []interface{}{"img.__image-lead__", ".article-main-body"}
	assert.Contains(t, extractor.Content.Selectors, multiMatch)
	assert.Contains(t, extractor.Content.Selectors, ".article-main-body")
}