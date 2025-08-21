// ABOUTME: Test suite for field validation framework
// ABOUTME: Comprehensive tests for all validator types and validation pipeline components

package validation

import (
	"fmt"
	"testing"
	"time"
	
	"github.com/BumpyClock/parser-go/pkg/extractors/fields"
)

func TestValidatorInterface(t *testing.T) {
	t.Run("StringValidator validates strings correctly", func(t *testing.T) {
		validator := NewStringValidator(StringOptions{
			MinLength: 1,
			MaxLength: 100,
			Required:  true,
		})

		// Test valid string
		err := validator.Validate("valid string")
		if err != nil {
			t.Errorf("Expected valid string to pass validation, got error: %v", err)
		}

		// Test empty string with required
		err = validator.Validate("")
		if err == nil {
			t.Error("Expected empty required string to fail validation")
		}

		// Test string too long
		longString := make([]byte, 101)
		for i := range longString {
			longString[i] = 'a'
		}
		err = validator.Validate(string(longString))
		if err == nil {
			t.Error("Expected overly long string to fail validation")
		}
	})

	t.Run("URLValidator validates URLs correctly", func(t *testing.T) {
		validator := NewURLValidator(URLOptions{
			RequireHTTPS: false,
			AllowQuery:   true,
		})

		// Test valid URL
		err := validator.Validate("https://example.com/path")
		if err != nil {
			t.Errorf("Expected valid URL to pass validation, got error: %v", err)
		}

		// Test invalid URL
		err = validator.Validate("not-a-url")
		if err == nil {
			t.Error("Expected invalid URL to fail validation")
		}

		// Test HTTPS requirement
		httpsValidator := NewURLValidator(URLOptions{RequireHTTPS: true})
		err = httpsValidator.Validate("http://example.com")
		if err == nil {
			t.Error("Expected HTTP URL to fail HTTPS validation")
		}
	})

	t.Run("DateValidator validates dates correctly", func(t *testing.T) {
		validator := NewDateValidator(DateOptions{
			RequireFuture: false,
			MaxAge:        2 * 365 * 24 * time.Hour, // 2 years - more lenient
		})

		// Test valid ISO date (recent date)
		recentDate := time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339) // 30 days ago
		err := validator.Validate(recentDate)
		if err != nil {
			t.Errorf("Expected valid recent date to pass validation, got error: %v", err)
		}

		// Test invalid date format
		err = validator.Validate("not-a-date")
		if err == nil {
			t.Error("Expected invalid date to fail validation")
		}

		// Test future requirement
		futureValidator := NewDateValidator(DateOptions{RequireFuture: true})
		pastDate := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
		err = futureValidator.Validate(pastDate)
		if err == nil {
			t.Error("Expected past date to fail future validation")
		}
	})

	t.Run("ImageValidator validates images correctly", func(t *testing.T) {
		validator := NewImageValidator(ImageOptions{
			RequireHTTPS:    false,
			AllowedFormats:  []string{"jpg", "jpeg", "png", "gif", "webp"},
			MaxFileSize:     5 * 1024 * 1024, // 5MB
		})

		// Test valid image URL
		err := validator.Validate("https://example.com/image.jpg")
		if err != nil {
			t.Errorf("Expected valid image URL to pass validation, got error: %v", err)
		}

		// Test unsupported format
		err = validator.Validate("https://example.com/image.bmp")
		if err == nil {
			t.Error("Expected unsupported image format to fail validation")
		}
	})

	t.Run("NumberValidator validates numbers correctly", func(t *testing.T) {
		validator := NewNumberValidator(NumberOptions{
			Min: 0,
			Max: 1000,
		})

		// Test valid number
		err := validator.Validate(500)
		if err != nil {
			t.Errorf("Expected valid number to pass validation, got error: %v", err)
		}

		// Test number out of range
		err = validator.Validate(1500)
		if err == nil {
			t.Error("Expected out-of-range number to fail validation")
		}

		// Test invalid type
		err = validator.Validate("not-a-number")
		if err == nil {
			t.Error("Expected non-numeric value to fail validation")
		}
	})
}

func TestValidationPipeline(t *testing.T) {
	t.Run("Pipeline chains multiple validators", func(t *testing.T) {
		pipeline := NewValidationPipeline()

		// Add validators to pipeline
		pipeline.AddValidator("string", NewStringValidator(StringOptions{MinLength: 1, MaxLength: 50}))
		pipeline.AddValidator("length", NewStringValidator(StringOptions{MinLength: 5}))

		// Test with valid value
		err := pipeline.Validate("valid string")
		if err != nil {
			t.Errorf("Expected valid value to pass pipeline, got error: %v", err)
		}

		// Test with invalid value (too short for second validator)
		err = pipeline.Validate("abc")
		if err == nil {
			t.Error("Expected short string to fail pipeline validation")
		}
	})

	t.Run("Pipeline aggregates errors correctly", func(t *testing.T) {
		pipeline := NewValidationPipeline()
		pipeline.SetErrorAggregation(true)

		// Add failing validators
		pipeline.AddValidator("length", NewStringValidator(StringOptions{MinLength: 10}))
		pipeline.AddValidator("required", NewStringValidator(StringOptions{Required: true}))

		// Test with empty string (should fail both)
		err := pipeline.Validate("")
		if err == nil {
			t.Error("Expected empty string to fail multiple validators")
		}

		// Check if error contains multiple validation failures
		aggErr, ok := err.(*ValidationError)
		if !ok {
			t.Error("Expected ValidationError type")
		}
		if len(aggErr.Errors) < 2 {
			t.Errorf("Expected multiple errors, got %d", len(aggErr.Errors))
		}
	})
}

func TestValidationConfiguration(t *testing.T) {
	t.Run("Validation profiles work correctly", func(t *testing.T) {
		// Test strict profile
		strictProfile := GetValidationProfile("strict")
		if !strictProfile.EnableAllValidations || strictProfile.ErrorHandling != "fail_fast" {
			t.Error("Strict profile should enable all validations and fail fast")
		}

		// Test lenient profile
		lenientProfile := GetValidationProfile("lenient")
		if lenientProfile.EnableAllValidations || lenientProfile.ErrorHandling != "collect_all" {
			t.Error("Lenient profile should be permissive and collect all errors")
		}
	})

	t.Run("Custom validation profiles can be created", func(t *testing.T) {
		customProfile := ValidationProfile{
			Name:                "custom",
			EnableAllValidations: true,
			ErrorHandling:       "warn_only",
			PerformanceMode:     "thorough",
		}

		RegisterValidationProfile("custom", customProfile)
		retrieved := GetValidationProfile("custom")

		if retrieved.Name != "custom" || retrieved.ErrorHandling != "warn_only" {
			t.Error("Custom profile was not registered correctly")
		}
	})
}

func TestFieldRegistry(t *testing.T) {
	t.Run("Fields can be registered dynamically", func(t *testing.T) {
		// Create a custom field definition
		customField := FieldDefinition{
			Name:        "custom_field",
			Type:        "string",
			Description: "A custom field for testing",
			Required:    true,
			Validators:  []ValidatorInterface{NewStringValidator(StringOptions{MinLength: 1})},
		}

		// Register the field
		err := RegisterField(customField)
		if err != nil {
			t.Errorf("Failed to register custom field: %v", err)
		}

		// Retrieve and validate the field
		retrieved, exists := GetFieldDefinition("custom_field")
		if !exists {
			t.Error("Custom field was not registered correctly")
		}
		if retrieved.Name != "custom_field" {
			t.Error("Retrieved field has incorrect name")
		}
	})

	t.Run("Field discovery works correctly", func(t *testing.T) {
		// Register multiple fields
		fields := []FieldDefinition{
			{Name: "test_field_1", Type: "string"},
			{Name: "test_field_2", Type: "url"},
			{Name: "test_field_3", Type: "date"},
		}

		for _, field := range fields {
			RegisterField(field)
		}

		// Discover all fields
		discovered := DiscoverFields()
		if len(discovered) < 3 {
			t.Errorf("Expected at least 3 discovered fields, got %d", len(discovered))
		}

		// Check that our test fields are present
		found := 0
		for _, field := range discovered {
			if field.Name == "test_field_1" || field.Name == "test_field_2" || field.Name == "test_field_3" {
				found++
			}
		}
		if found != 3 {
			t.Errorf("Expected 3 test fields in discovery, found %d", found)
		}
	})
}

func TestExtendedFields(t *testing.T) {
	t.Run("Category field extraction works", func(t *testing.T) {
		extractor := fields.NewCategoryExtractor()
		
		// Test valid categories
		result := extractor.Extract([]string{"technology", "science", "news"})
		categoryField, ok := result.(fields.CategoryField)
		if !ok {
			t.Errorf("Expected CategoryField, got %T", result)
		}
		if categoryField.Primary == "" {
			t.Error("Expected non-empty primary category")
		}
	})

	t.Run("Tags field extraction works", func(t *testing.T) {
		extractor := fields.NewTagsExtractor()
		
		// Test tag normalization
		result := extractor.Extract([]string{"Web Development", "go-lang", "API_Design"})
		tags, ok := result.([]string)
		if !ok {
			t.Errorf("Expected []string, got %T", result)
		}
		
		if len(tags) == 0 {
			t.Error("Expected non-empty tags slice")
		}
	})

	t.Run("Related articles field extraction works", func(t *testing.T) {
		extractor := fields.NewRelatedArticlesExtractor()
		
		// Mock related articles data
		mockData := []map[string]interface{}{
			{"title": "Related Article 1", "url": "https://example.com/article1"},
			{"title": "Related Article 2", "url": "https://example.com/article2"},
		}
		
		result := extractor.Extract(mockData)
		articles, ok := result.([]fields.RelatedArticle)
		if !ok {
			t.Errorf("Expected []RelatedArticle, got %T", result)
		}
		if len(articles) != 2 {
			t.Errorf("Expected 2 related articles, got %d", len(articles))
		}
	})
}

func TestFieldTransformers(t *testing.T) {
	t.Run("String transformer normalizes correctly", func(t *testing.T) {
		transformer := fields.NewStringTransformer()
		
		input := "  Test String  \n\t"
		expected := "Test String"
		result := transformer.Transform(input)
		
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("URL transformer resolves correctly", func(t *testing.T) {
		transformer := fields.NewURLTransformer("https://example.com")
		
		// Test relative URL resolution
		relative := "/path/to/resource"
		expected := "https://example.com/path/to/resource"
		result := transformer.Transform(relative)
		
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("Date transformer parses correctly", func(t *testing.T) {
		transformer := fields.NewDateTransformer()
		
		// Test various date formats
		testCases := []string{
			"2023-12-25T10:30:00Z",
			"December 25, 2023",
			"2023/12/25",
		}
		
		for _, testCase := range testCases {
			result := transformer.Transform(testCase)
			if _, ok := result.(time.Time); !ok {
				t.Errorf("Expected time.Time for input %q, got %T", testCase, result)
			}
		}
	})
}

func TestPerformanceAndThreadSafety(t *testing.T) {
	t.Run("Validator registry is thread-safe", func(t *testing.T) {
		// Test concurrent registration and retrieval
		done := make(chan bool, 10)
		
		for i := 0; i < 10; i++ {
			go func(id int) {
				field := FieldDefinition{
					Name: fmt.Sprintf("concurrent_field_%d", id),
					Type: "string",
				}
				RegisterField(field)
				
				// Try to retrieve immediately
				_, exists := GetFieldDefinition(field.Name)
				if !exists {
					t.Errorf("Field %s was not found after registration", field.Name)
				}
				done <- true
			}(i)
		}
		
		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("Validation has minimal performance impact when disabled", func(t *testing.T) {
		// Create validator but disable validation
		validator := NewStringValidator(StringOptions{MinLength: 1})
		validator.SetEnabled(false)
		
		// Measure performance (should be near-zero when disabled)
		start := time.Now()
		for i := 0; i < 1000; i++ {
			validator.Validate("test string")
		}
		duration := time.Since(start)
		
		// Should complete very quickly when disabled
		if duration > time.Millisecond {
			t.Errorf("Disabled validation took too long: %v", duration)
		}
	})
}