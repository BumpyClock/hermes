// ABOUTME: Documentation and usage guide for the field validation framework
// ABOUTME: Comprehensive package documentation with examples and best practices

/*
Package validation provides a comprehensive field validation framework for extracted fields and extended field support.

# Overview

The validation framework offers:
- Type-specific validators (string, URL, date, image, number)
- Validation pipelines for chaining multiple validators
- Custom validators for domain-specific rules
- Performance optimization with configurable validation modes
- Thread-safe validator registry for dynamic field management
- Extended field types (categories, tags, related articles, sentiment)
- Field transformers for data normalization and format conversion
- Comprehensive configuration and metrics system

# Quick Start

Basic validation example:

	import "github.com/BumpyClock/parser-go/pkg/extractors/validation"
	
	// Create a string validator
	validator := validation.NewStringValidator(validation.StringOptions{
		MinLength: 1,
		MaxLength: 100,
		Required:  true,
	})
	
	// Validate a value
	err := validator.Validate("hello world")
	if err != nil {
		log.Printf("Validation failed: %v", err)
	}

# Validation Pipeline

Chain multiple validators for complex validation:

	pipeline := validation.NewValidationPipeline()
	pipeline.AddValidator("length", validation.NewStringValidator(validation.StringOptions{
		MinLength: 5,
		MaxLength: 50,
	}))
	pipeline.AddValidator("url", validation.NewURLValidator(validation.URLOptions{
		RequireHTTPS: true,
	}))
	
	err := pipeline.Validate("https://example.com")

# Field Registry

Register custom field definitions dynamically:

	fieldDef := validation.FieldDefinition{
		Name:        "custom_field",
		Type:        "string", 
		Description: "A custom field for articles",
		Required:    true,
		Validators: []validation.ValidatorInterface{
			validation.NewStringValidator(validation.StringOptions{MinLength: 1}),
		},
	}
	
	err := validation.RegisterField(fieldDef)
	if err != nil {
		log.Printf("Field registration failed: %v", err)
	}

# Extended Field Types

Work with specialized field types:

	import "github.com/BumpyClock/parser-go/pkg/extractors/fields"
	
	// Extract categories
	categoryExtractor := fields.NewCategoryExtractor()
	result := categoryExtractor.Extract([]string{"technology", "science"})
	categoryField := result.(fields.CategoryField)
	
	// Extract and normalize tags
	tagsExtractor := fields.NewTagsExtractor()
	tags := tagsExtractor.Extract([]string{"Web Development", "Go Programming"})
	normalizedTags := tags.([]string) // ["web-development", "go-programming"]

# Field Transformers

Transform extracted data to standard formats:

	// String normalization
	stringTransformer := fields.NewStringTransformer()
	normalized := stringTransformer.Transform("  hello world  ") // "hello world"
	
	// URL resolution
	urlTransformer := fields.NewURLTransformer("https://example.com")
	resolved := urlTransformer.Transform("/path") // "https://example.com/path"
	
	// Date parsing
	dateTransformer := fields.NewDateTransformer()
	parsed := dateTransformer.Transform("2023-12-25T10:30:00Z") // time.Time

# Configuration and Profiles

Configure validation behavior with profiles:

	// Use strict validation profile
	err := validation.SetValidationProfile("strict")
	if err != nil {
		log.Printf("Failed to set profile: %v", err)
	}
	
	// Create custom configuration
	builder := validation.NewValidationConfigBuilder()
	config, fieldConfigs := builder.
		WithProfile("custom").
		WithPerformanceMode("thorough").
		WithErrorHandling("collect_all").
		WithTimeout(10 * time.Second).
		Build()
	
	validation.SetGlobalConfig(config)

# Performance and Metrics

Monitor validation performance:

	// Get global metrics
	metrics := validation.GetGlobalMetrics()
	log.Printf("Total validations: %d", metrics.TotalValidations)
	log.Printf("Success rate: %.2f%%", 
		float64(metrics.SuccessfulValidations)/float64(metrics.TotalValidations)*100)
	log.Printf("Average time: %v", metrics.AverageValidationTime)

# Thread Safety

All validators and the registry are thread-safe:

	// Safe to use from multiple goroutines
	validator := validation.NewStringValidator(validation.StringOptions{MinLength: 1})
	
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := validator.Validate(fmt.Sprintf("value-%d", id))
			if err != nil {
				log.Printf("Validation %d failed: %v", id, err)
			}
		}(i)
	}
	wg.Wait()

# Custom Validators

Create domain-specific validators:

	emailValidator := validation.NewCustomValidator("email", "string", 
		func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("expected string")
			}
			if !strings.Contains(str, "@") {
				return fmt.Errorf("invalid email format")
			}
			return nil
		})

# Validation Profiles

Built-in profiles:

- "strict": Enables all validations, fails fast on first error
- "lenient": Permissive validation, collects all errors
- "production": Balanced approach, warns on errors but doesn't fail

Custom profiles can be registered:

	customProfile := validation.ValidationProfile{
		Name:                 "custom",
		EnableAllValidations: true,
		ErrorHandling:        "warn_only",
		PerformanceMode:      "fast",
	}
	validation.RegisterValidationProfile("custom", customProfile)

# Integration with Parser

The validation framework integrates seamlessly with the parser:

	import "github.com/BumpyClock/parser-go/pkg/parser"
	
	// Validation is automatically applied during field extraction
	// Configure validation behavior as needed
	validation.SetValidationProfile("production")
	
	// Parse content - validation will be applied to extracted fields
	p := parser.NewParser()
	result, err := p.Parse("https://example.com", parser.ParserOptions{})

# Error Handling

Validation errors provide detailed information:

	err := validator.Validate(invalidValue)
	if validationErr, ok := err.(*validation.ValidationError); ok {
		log.Printf("Field: %s", validationErr.Field)
		log.Printf("Message: %s", validationErr.Message)
		for i, subErr := range validationErr.Errors {
			log.Printf("Error %d: %v", i+1, subErr)
		}
	}

# Best Practices

1. **Use appropriate validation profiles** for different environments (dev/staging/prod)
2. **Register custom fields early** in your application initialization
3. **Monitor validation metrics** to identify performance bottlenecks
4. **Use transformers** to normalize data before validation
5. **Chain validators** for complex validation requirements
6. **Handle validation errors gracefully** in production environments
7. **Test validation rules thoroughly** with edge cases
8. **Configure reasonable timeouts** for validation operations

# Architecture

The validation framework is designed with the following principles:

- **Modularity**: Each validator type is independent and composable
- **Extensibility**: Easy to add new validator types and field definitions
- **Performance**: Minimal overhead when validation is disabled
- **Thread Safety**: All components are safe for concurrent use
- **Configuration**: Flexible configuration system for different environments
- **Observability**: Comprehensive metrics and monitoring support

# Type System

The framework provides type-safe validation for:

- Strings (length, pattern, required constraints)
- URLs (security validation, scheme requirements, domain filtering)
- Dates (format parsing, age constraints, future/past requirements)
- Images (format validation, size constraints)
- Numbers (range validation, integer requirements)
- Custom types (domain-specific validation logic)

All validators implement the ValidatorInterface for consistent behavior.
*/
package validation