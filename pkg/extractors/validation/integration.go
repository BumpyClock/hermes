// ABOUTME: Integration utilities for connecting validation framework with parser extraction pipeline
// ABOUTME: Provides seamless validation integration for extracted content fields

package validation

import (
	"time"
	
	"github.com/PuerkitoBio/goquery"
)

// ValidatedResult represents a parser result with validation information
type ValidatedResult struct {
	// Standard parser result fields
	Title          string                 `json:"title,omitempty"`
	Content        string                 `json:"content,omitempty"`
	Author         string                 `json:"author,omitempty"`
	DatePublished  *time.Time            `json:"date_published,omitempty"`
	LeadImageURL   string                 `json:"lead_image_url,omitempty"`
	Dek            string                 `json:"dek,omitempty"`
	URL            string                 `json:"url,omitempty"`
	Domain         string                 `json:"domain,omitempty"`
	Excerpt        string                 `json:"excerpt,omitempty"`
	WordCount      int                    `json:"word_count,omitempty"`
	
	// Extended fields
	ExtendedFields map[string]interface{} `json:"extended_fields,omitempty"`
	
	// Validation information
	ValidationResults map[string]ValidationResult `json:"validation_results,omitempty"`
	ValidationSummary ValidationSummary           `json:"validation_summary"`
}

// ValidationResult represents the validation outcome for a single field
type ValidationResult struct {
	Valid       bool                   `json:"valid"`
	Errors      []string              `json:"errors,omitempty"`
	Warnings    []string              `json:"warnings,omitempty"`
	Confidence  float64               `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationSummary provides an overview of validation results
type ValidationSummary struct {
	TotalFields    int     `json:"total_fields"`
	ValidFields    int     `json:"valid_fields"`
	InvalidFields  int     `json:"invalid_fields"`
	WarningFields  int     `json:"warning_fields"`
	OverallValid   bool    `json:"overall_valid"`
	Confidence     float64 `json:"confidence"`
}

// FieldValidator combines validation and transformation for a specific field
type FieldValidator struct {
	pipeline    *ValidationPipeline
	transformer FieldTransformer
	required    bool
	fieldName   string
}

// NewFieldValidator creates a new field validator
func NewFieldValidator(fieldName string) *FieldValidator {
	return &FieldValidator{
		pipeline:  NewValidationPipeline(),
		fieldName: fieldName,
	}
}

// AddValidator adds a validator to the field's pipeline
func (fv *FieldValidator) AddValidator(name string, validator ValidatorInterface) *FieldValidator {
	fv.pipeline.AddValidator(name, validator)
	return fv
}

// SetTransformer sets the field transformer
func (fv *FieldValidator) SetTransformer(transformer FieldTransformer) *FieldValidator {
	fv.transformer = transformer
	return fv
}

// SetRequired sets whether the field is required
func (fv *FieldValidator) SetRequired(required bool) *FieldValidator {
	fv.required = required
	return fv
}

// ValidateAndTransform validates and transforms a field value
func (fv *FieldValidator) ValidateAndTransform(value interface{}) (interface{}, ValidationResult) {
	startTime := time.Now()
	
	result := ValidationResult{
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
		Metadata: map[string]interface{}{
			"field_name":      fv.fieldName,
			"validation_time": time.Since(startTime),
		},
	}
	
	// Check required constraint
	if fv.required && (value == nil || value == "") {
		result.Errors = append(result.Errors, "field is required but missing")
		result.Valid = false
		result.Confidence = 0.0
		return value, result
	}
	
	// Skip validation if value is empty and not required
	if (value == nil || value == "") && !fv.required {
		result.Valid = true
		result.Confidence = 1.0
		return value, result
	}
	
	// Apply transformer if available
	transformedValue := value
	if fv.transformer != nil {
		transformedValue = fv.transformer.Transform(value)
		result.Metadata["transformed"] = true
		result.Metadata["original_value"] = value
	}
	
	// Run validation pipeline
	if err := fv.pipeline.Validate(transformedValue); err != nil {
		if validationErr, ok := err.(*ValidationError); ok {
			for _, subErr := range validationErr.Errors {
				result.Errors = append(result.Errors, subErr.Error())
			}
		} else {
			result.Errors = append(result.Errors, err.Error())
		}
		result.Valid = false
		result.Confidence = 0.0
	} else {
		result.Valid = true
		result.Confidence = 1.0
	}
	
	// Record metrics
	RecordGlobalValidation(fv.fieldName, result.Valid, time.Since(startTime))
	
	return transformedValue, result
}

// ParserValidator integrates validation with the parser system
type ParserValidator struct {
	fieldValidators map[string]*FieldValidator
	config         *ValidationConfig
	enabled        bool
}

// NewParserValidator creates a new parser validator
func NewParserValidator() *ParserValidator {
	return &ParserValidator{
		fieldValidators: make(map[string]*FieldValidator),
		config:         GetGlobalConfig(),
		enabled:        true,
	}
}

// RegisterFieldValidator registers a validator for a specific field
func (pv *ParserValidator) RegisterFieldValidator(fieldName string, validator *FieldValidator) {
	pv.fieldValidators[fieldName] = validator
}

// SetEnabled enables or disables validation
func (pv *ParserValidator) SetEnabled(enabled bool) {
	pv.enabled = enabled
}

// ValidateResult validates a complete parser result
func (pv *ParserValidator) ValidateResult(result interface{}) *ValidatedResult {
	if !pv.enabled {
		// Return basic validated result without validation
		return pv.createBasicValidatedResult(result)
	}
	
	validatedResult := &ValidatedResult{
		ValidationResults: make(map[string]ValidationResult),
		ExtendedFields:    make(map[string]interface{}),
	}
	
	// Extract and validate standard fields
	pv.validateStandardFields(result, validatedResult)
	
	// Calculate validation summary
	pv.calculateValidationSummary(validatedResult)
	
	return validatedResult
}

// validateStandardFields validates the standard parser result fields
func (pv *ParserValidator) validateStandardFields(result interface{}, validatedResult *ValidatedResult) {
	// Use reflection or type assertion to extract fields from result
	// This is a simplified implementation - in practice, you'd handle the actual parser result type
	
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return
	}
	
	// Validate each field
	fieldMappings := map[string]func(interface{}){
		"title": func(value interface{}) {
			transformed, result := pv.validateField("title", value)
			if result.Valid {
				validatedResult.Title = transformed.(string)
			}
			validatedResult.ValidationResults["title"] = result
		},
		"content": func(value interface{}) {
			transformed, result := pv.validateField("content", value)
			if result.Valid {
				validatedResult.Content = transformed.(string)
			}
			validatedResult.ValidationResults["content"] = result
		},
		"author": func(value interface{}) {
			transformed, result := pv.validateField("author", value)
			if result.Valid {
				validatedResult.Author = transformed.(string)
			}
			validatedResult.ValidationResults["author"] = result
		},
		"url": func(value interface{}) {
			transformed, result := pv.validateField("url", value)
			if result.Valid {
				validatedResult.URL = transformed.(string)
			}
			validatedResult.ValidationResults["url"] = result
		},
		"lead_image_url": func(value interface{}) {
			transformed, result := pv.validateField("lead_image_url", value)
			if result.Valid {
				validatedResult.LeadImageURL = transformed.(string)
			}
			validatedResult.ValidationResults["lead_image_url"] = result
		},
	}
	
	// Apply field validations
	for fieldName, validateFunc := range fieldMappings {
		if value, exists := resultMap[fieldName]; exists {
			validateFunc(value)
		}
	}
}

// validateField validates a single field using registered validators
func (pv *ParserValidator) validateField(fieldName string, value interface{}) (interface{}, ValidationResult) {
	if validator, exists := pv.fieldValidators[fieldName]; exists {
		return validator.ValidateAndTransform(value)
	}
	
	// No specific validator - return as valid
	return value, ValidationResult{
		Valid:      true,
		Confidence: 0.5, // Lower confidence for unvalidated fields
		Metadata: map[string]interface{}{
			"validated": false,
			"reason":    "no validator configured",
		},
	}
}

// calculateValidationSummary calculates the overall validation summary
func (pv *ParserValidator) calculateValidationSummary(result *ValidatedResult) {
	summary := ValidationSummary{
		TotalFields: len(result.ValidationResults),
	}
	
	var confidenceSum float64
	
	for _, validation := range result.ValidationResults {
		if validation.Valid {
			summary.ValidFields++
		} else {
			summary.InvalidFields++
		}
		
		if len(validation.Warnings) > 0 {
			summary.WarningFields++
		}
		
		confidenceSum += validation.Confidence
	}
	
	if summary.TotalFields > 0 {
		summary.Confidence = confidenceSum / float64(summary.TotalFields)
		summary.OverallValid = summary.InvalidFields == 0
	}
	
	result.ValidationSummary = summary
}

// createBasicValidatedResult creates a validated result without validation when disabled
func (pv *ParserValidator) createBasicValidatedResult(result interface{}) *ValidatedResult {
	// Convert result to ValidatedResult format without validation
	validatedResult := &ValidatedResult{
		ValidationResults: make(map[string]ValidationResult),
		ExtendedFields:    make(map[string]interface{}),
		ValidationSummary: ValidationSummary{
			TotalFields:   0,
			ValidFields:   0,
			InvalidFields: 0,
			OverallValid:  true,
			Confidence:    1.0,
		},
	}
	
	// Copy fields from result (simplified implementation)
	if resultMap, ok := result.(map[string]interface{}); ok {
		if title, exists := resultMap["title"].(string); exists {
			validatedResult.Title = title
		}
		if content, exists := resultMap["content"].(string); exists {
			validatedResult.Content = content
		}
		// ... copy other fields as needed
	}
	
	return validatedResult
}

// SetupDefaultValidators configures default validators for common fields
func SetupDefaultValidators(parserValidator *ParserValidator) {
	// Title validator
	titleValidator := NewFieldValidator("title").
		AddValidator("length", NewStringValidator(StringOptions{
			MinLength: 1,
			MaxLength: 200,
		})).
		SetRequired(true)
	
	parserValidator.RegisterFieldValidator("title", titleValidator)
	
	// URL validator
	urlValidator := NewFieldValidator("url").
		AddValidator("url", NewURLValidator(URLOptions{
			RequireHTTPS: false,
		})).
		SetRequired(true)
	
	parserValidator.RegisterFieldValidator("url", urlValidator)
	
	// Author validator
	authorValidator := NewFieldValidator("author").
		AddValidator("length", NewStringValidator(StringOptions{
			MinLength: 1,
			MaxLength: 100,
		})).
		SetRequired(false)
	
	parserValidator.RegisterFieldValidator("author", authorValidator)
	
	// Content validator
	contentValidator := NewFieldValidator("content").
		AddValidator("length", NewStringValidator(StringOptions{
			MinLength: 10,
			MaxLength: 50000,
		})).
		SetRequired(true)
	
	parserValidator.RegisterFieldValidator("content", contentValidator)
	
	// Lead image URL validator
	imageValidator := NewFieldValidator("lead_image_url").
		AddValidator("image", NewImageValidator(ImageOptions{
			RequireHTTPS:   false,
			AllowedFormats: []string{"jpg", "jpeg", "png", "gif", "webp"},
		})).
		SetRequired(false)
	
	parserValidator.RegisterFieldValidator("lead_image_url", imageValidator)
}

// ValidateDocument validates extracted content from a goquery document
func ValidateDocument(doc *goquery.Document, url string) *ValidatedResult {
	parserValidator := NewParserValidator()
	SetupDefaultValidators(parserValidator)
	
	// Extract basic fields for validation
	title := doc.Find("title").First().Text()
	
	// Create a simple result map
	result := map[string]interface{}{
		"title": title,
		"url":   url,
	}
	
	return parserValidator.ValidateResult(result)
}

// Import FieldTransformer from fields package
// type FieldTransformer is defined in interface.go