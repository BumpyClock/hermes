// ABOUTME: Custom extractor for www.apartmenttherapy.com - Home design and lifestyle site
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.apartmenttherapy.com/index.js ApartmentTherapyExtractor

package custom

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
)

// GetWwwApartmenttherapyComExtractor returns the custom extractor for www.apartmenttherapy.com
func GetWwwApartmenttherapyComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.apartmenttherapy.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1.headline",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				".PostByline__name",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{".PostByline__timestamp[datetime]", "datetime"},
			},
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"og:image\"]", "value"},
			},
		},
		
		Dek: &FieldExtractor{
			Selectors: []interface{}{
				// Empty array in JavaScript version
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					"div.post__content",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// Complex JSON-based transform for lazy-loaded images
				"div[data-render-react-id=\"images/LazyPicture\"]": &FunctionTransform{
					Fn: transformApartmentTherapyLazyPicture,
				},
			},
			
			Clean: []string{
				// No clean selectors in JavaScript version
			},
		},
	}
}

// transformApartmentTherapyLazyPicture handles lazy-loaded images with JSON data-props
// JavaScript equivalent: 'div[data-render-react-id="images/LazyPicture"]': ($node, $) => { ... }
func transformApartmentTherapyLazyPicture(selection *goquery.Selection) error {
	// Get data-props attribute
	dataProps, exists := selection.Attr("data-props")
	if !exists {
		return nil
	}
	
	// Parse JSON data
	var data struct {
		Sources []struct {
			Src string `json:"src"`
		} `json:"sources"`
	}
	
	if err := json.Unmarshal([]byte(dataProps), &data); err != nil {
		return nil // Ignore parsing errors, JavaScript version continues silently
	}
	
	// Check if we have sources
	if len(data.Sources) == 0 {
		return nil
	}
	
	// Get src from first source
	src := data.Sources[0].Src
	if src == "" {
		return nil
	}
	
	// Create img element and replace the div
	imgHtml := "<img src=\"" + src + "\" />"
	selection.ReplaceWithHtml(imgHtml)
	
	return nil
}