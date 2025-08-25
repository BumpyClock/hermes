// ABOUTME: Type-safe transform interface to replace JavaScript callback patterns
// ABOUTME: Provides better performance and compile-time safety vs map[string]interface{}

package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Transform defines the interface for content transformation operations
// Replaces JavaScript transform callback functions with proper Go interfaces
type Transform interface {
	// Transform applies the transformation to the given selection
	Transform(selection *goquery.Selection) *goquery.Selection
	
	// Name returns the name/type of this transform
	Name() string
}

// TransformRegistry holds type-safe transform implementations
// Replaces map[string]interface{} with proper typed registry
type TransformRegistry map[string]Transform

// TagRenameTransform renames HTML tags (e.g., h1 -> h2)
type TagRenameTransform struct {
	OriginalTag string
	NewTag      string
}

func (t *TagRenameTransform) Transform(selection *goquery.Selection) *goquery.Selection {
	// Rename tag while preserving attributes and content
	selection.Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == t.OriginalTag {
			// Get all attributes
			attrs := make(map[string]string)
			for _, attr := range s.Get(0).Attr {
				attrs[attr.Key] = attr.Val
			}
			
			// Get content  
			htmlContent, err := s.Html()
			
			// Create new element with new tag
			newElem := fmt.Sprintf("<%s", t.NewTag)
			for key, val := range attrs {
				newElem += fmt.Sprintf(` %s="%s"`, key, val)
			}
			newElem += ">"
			if err == nil {
				newElem += htmlContent
			}
			newElem += fmt.Sprintf("</%s>", t.NewTag)
			
			// Replace current element
			s.ReplaceWithHtml(newElem)
		}
	})
	return selection
}

func (t *TagRenameTransform) Name() string {
	return fmt.Sprintf("rename_%s_to_%s", t.OriginalTag, t.NewTag)
}

// ClassAddTransform adds a CSS class to elements
type ClassAddTransform struct {
	ClassName string
}

func (t *ClassAddTransform) Transform(selection *goquery.Selection) *goquery.Selection {
	selection.AddClass(t.ClassName)
	return selection
}

func (t *ClassAddTransform) Name() string {
	return fmt.Sprintf("add_class_%s", t.ClassName)
}

// ClassRemoveTransform removes a CSS class from elements
type ClassRemoveTransform struct {
	ClassName string
}

func (t *ClassRemoveTransform) Transform(selection *goquery.Selection) *goquery.Selection {
	selection.RemoveClass(t.ClassName)
	return selection
}

func (t *ClassRemoveTransform) Name() string {
	return fmt.Sprintf("remove_class_%s", t.ClassName)
}

// AttributeSetTransform sets an attribute value
type AttributeSetTransform struct {
	AttributeName  string
	AttributeValue string
}

func (t *AttributeSetTransform) Transform(selection *goquery.Selection) *goquery.Selection {
	selection.SetAttr(t.AttributeName, t.AttributeValue)
	return selection
}

func (t *AttributeSetTransform) Name() string {
	return fmt.Sprintf("set_%s_%s", t.AttributeName, t.AttributeValue)
}

// AttributeRemoveTransform removes an attribute
type AttributeRemoveTransform struct {
	AttributeName string
}

func (t *AttributeRemoveTransform) Transform(selection *goquery.Selection) *goquery.Selection {
	selection.RemoveAttr(t.AttributeName)
	return selection
}

func (t *AttributeRemoveTransform) Name() string {
	return fmt.Sprintf("remove_attr_%s", t.AttributeName)
}

// TextReplaceTransform replaces text content using string replacement
type TextReplaceTransform struct {
	OldText string
	NewText string
}

func (t *TextReplaceTransform) Transform(selection *goquery.Selection) *goquery.Selection {
	selection.Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, t.OldText) {
			newText := strings.ReplaceAll(text, t.OldText, t.NewText)
			s.SetText(newText)
		}
	})
	return selection
}

func (t *TextReplaceTransform) Name() string {
	return fmt.Sprintf("replace_text_%s_to_%s", t.OldText, t.NewText)
}

// CustomFunctionTransform wraps a custom function for compatibility
// Used during migration period when some transforms are still functions
type CustomFunctionTransform struct {
	TransformName     string
	TransformFunction func(*goquery.Selection) *goquery.Selection
}

func (t *CustomFunctionTransform) Transform(selection *goquery.Selection) *goquery.Selection {
	if t.TransformFunction != nil {
		return t.TransformFunction(selection)
	}
	return selection
}

func (t *CustomFunctionTransform) Name() string {
	return t.TransformName
}

// NewTransformRegistry creates a registry with common transforms
func NewTransformRegistry() TransformRegistry {
	registry := make(TransformRegistry)
	
	// Register common transform patterns from JavaScript parser
	registry["h1_to_h2"] = &TagRenameTransform{OriginalTag: "h1", NewTag: "h2"}
	registry["h2_to_h3"] = &TagRenameTransform{OriginalTag: "h2", NewTag: "h3"}
	registry["h3_to_h4"] = &TagRenameTransform{OriginalTag: "h3", NewTag: "h4"}
	registry["h4_to_h5"] = &TagRenameTransform{OriginalTag: "h4", NewTag: "h5"}
	registry["h5_to_h6"] = &TagRenameTransform{OriginalTag: "h5", NewTag: "h6"}
	
	registry["add_hermes_keep"] = &ClassAddTransform{ClassName: "hermes-parser-keep"}
	registry["remove_hermes_keep"] = &ClassRemoveTransform{ClassName: "hermes-parser-keep"}
	
	return registry
}

// ConvertLegacyTransforms converts map[string]interface{} to TransformRegistry
// This enables gradual migration from JavaScript patterns to Go interfaces
func ConvertLegacyTransforms(legacy map[string]interface{}) TransformRegistry {
	registry := NewTransformRegistry()
	
	for name, transform := range legacy {
		switch t := transform.(type) {
		case Transform:
			// Already a proper transform
			registry[name] = t
		case func(*goquery.Selection) *goquery.Selection:
			// Function-based transform - wrap it
			registry[name] = &CustomFunctionTransform{
				TransformName:     name,
				TransformFunction: t,
			}
		case string:
			// String-based transform - parse common patterns
			if strings.HasPrefix(t, "rename_") {
				parts := strings.Split(t, "_")
				if len(parts) >= 4 && parts[2] == "to" {
					registry[name] = &TagRenameTransform{
						OriginalTag: parts[1],
						NewTag:      parts[3],
					}
				}
			}
		default:
			// Unknown type - create a no-op transform
			registry[name] = &CustomFunctionTransform{
				TransformName: name,
				TransformFunction: func(s *goquery.Selection) *goquery.Selection {
					return s // No-op
				},
			}
		}
	}
	
	return registry
}

// ApplyTransforms applies all transforms in the registry to a selection
func ApplyTransforms(selection *goquery.Selection, transforms TransformRegistry) *goquery.Selection {
	if len(transforms) == 0 {
		return selection
	}
	
	result := selection
	for _, transform := range transforms {
		result = transform.Transform(result)
	}
	
	return result
}

// GetTransformNames returns all transform names in the registry
func (tr TransformRegistry) GetTransformNames() []string {
	names := make([]string, 0, len(tr))
	for name := range tr {
		names = append(names, name)
	}
	return names
}

// HasTransform checks if a transform exists in the registry
func (tr TransformRegistry) HasTransform(name string) bool {
	_, exists := tr[name]
	return exists
}

// GetTransform retrieves a specific transform by name
func (tr TransformRegistry) GetTransform(name string) (Transform, bool) {
	transform, exists := tr[name]
	return transform, exists
}