// ABOUTME: Core validation interfaces and types for field validation framework
// ABOUTME: Defines the base validator interface and core validation abstractions

package validation

import (
	"fmt"
	"sync"
	"time"
)

// ValidatorInterface defines the contract for all field validators
type ValidatorInterface interface {
	// Validate checks if the given value is valid according to validator rules
	Validate(value interface{}) error
	
	// Name returns the validator name for identification and error reporting
	Name() string
	
	// Type returns the data type this validator handles (string, url, date, etc.)
	Type() string
	
	// SetEnabled allows enabling/disabling validation for performance control
	SetEnabled(enabled bool)
	
	// IsEnabled returns whether validation is currently enabled
	IsEnabled() bool
}

// ValidationPipeline chains multiple validators together
type ValidationPipeline struct {
	validators        map[string]ValidatorInterface
	validatorOrder    []string
	errorAggregation  bool
	mu                sync.RWMutex
}

// NewValidationPipeline creates a new validation pipeline
func NewValidationPipeline() *ValidationPipeline {
	return &ValidationPipeline{
		validators:       make(map[string]ValidatorInterface),
		validatorOrder:   make([]string, 0),
		errorAggregation: false,
	}
}

// AddValidator adds a validator to the pipeline
func (vp *ValidationPipeline) AddValidator(name string, validator ValidatorInterface) {
	vp.mu.Lock()
	defer vp.mu.Unlock()
	
	vp.validators[name] = validator
	vp.validatorOrder = append(vp.validatorOrder, name)
}

// SetErrorAggregation enables or disables error aggregation
func (vp *ValidationPipeline) SetErrorAggregation(enabled bool) {
	vp.mu.Lock()
	defer vp.mu.Unlock()
	vp.errorAggregation = enabled
}

// Validate runs all validators in the pipeline
func (vp *ValidationPipeline) Validate(value interface{}) error {
	vp.mu.RLock()
	defer vp.mu.RUnlock()
	
	var errors []error
	
	for _, name := range vp.validatorOrder {
		validator := vp.validators[name]
		if !validator.IsEnabled() {
			continue
		}
		
		if err := validator.Validate(value); err != nil {
			if vp.errorAggregation {
				errors = append(errors, fmt.Errorf("validator '%s': %w", name, err))
			} else {
				// Fail fast - return first error
				return fmt.Errorf("validator '%s': %w", name, err)
			}
		}
	}
	
	if len(errors) > 0 {
		return &ValidationError{
			Message: "Multiple validation failures",
			Errors:  errors,
		}
	}
	
	return nil
}

// ValidationError represents validation failures
type ValidationError struct {
	Message string
	Errors  []error
	Field   string
}

func (ve *ValidationError) Error() string {
	if len(ve.Errors) == 1 {
		return fmt.Sprintf("validation error for field '%s': %s", ve.Field, ve.Errors[0].Error())
	}
	return fmt.Sprintf("validation error for field '%s': %s (%d errors)", ve.Field, ve.Message, len(ve.Errors))
}

// FieldDefinition describes a field type with its validation rules
type FieldDefinition struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Validators  []ValidatorInterface
	Transformer FieldTransformer
	Examples    []interface{}
	
	// Metadata for documentation generation
	Category    string
	Version     string
	Deprecated  bool
}

// FieldTransformer converts extracted data to standard formats
type FieldTransformer interface {
	Transform(value interface{}) interface{}
	TargetType() string
}

// ValidationProfile defines validation behavior settings
type ValidationProfile struct {
	Name                 string
	EnableAllValidations bool
	ErrorHandling        string // "fail_fast", "collect_all", "warn_only"
	PerformanceMode      string // "fast", "thorough"
	CustomRules          map[string]interface{}
}

// Validation configuration and profiles
var (
	validationProfiles = map[string]ValidationProfile{
		"strict": {
			Name:                 "strict",
			EnableAllValidations: true,
			ErrorHandling:        "fail_fast",
			PerformanceMode:      "thorough",
		},
		"lenient": {
			Name:                 "lenient",
			EnableAllValidations: false,
			ErrorHandling:        "collect_all",
			PerformanceMode:      "fast",
		},
		"production": {
			Name:                 "production",
			EnableAllValidations: true,
			ErrorHandling:        "warn_only",
			PerformanceMode:      "fast",
		},
	}
	profileMutex sync.RWMutex
)

// RegisterValidationProfile registers a new validation profile
func RegisterValidationProfile(name string, profile ValidationProfile) {
	profileMutex.Lock()
	defer profileMutex.Unlock()
	validationProfiles[name] = profile
}

// GetValidationProfile retrieves a validation profile by name
func GetValidationProfile(name string) ValidationProfile {
	profileMutex.RLock()
	defer profileMutex.RUnlock()
	
	if profile, exists := validationProfiles[name]; exists {
		return profile
	}
	
	// Return default profile if not found
	return validationProfiles["lenient"]
}

// Field registry for dynamic field management
var (
	fieldRegistry = make(map[string]FieldDefinition)
	registryMutex sync.RWMutex
)

// RegisterField registers a new field definition
func RegisterField(field FieldDefinition) error {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	
	if field.Name == "" {
		return fmt.Errorf("field name cannot be empty")
	}
	
	if field.Type == "" {
		return fmt.Errorf("field type cannot be empty")
	}
	
	fieldRegistry[field.Name] = field
	return nil
}

// GetFieldDefinition retrieves a field definition by name
func GetFieldDefinition(name string) (FieldDefinition, bool) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	
	field, exists := fieldRegistry[name]
	return field, exists
}

// DiscoverFields returns all registered field definitions
func DiscoverFields() []FieldDefinition {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	
	fields := make([]FieldDefinition, 0, len(fieldRegistry))
	for _, field := range fieldRegistry {
		fields = append(fields, field)
	}
	
	return fields
}

// GetFieldsByCategory returns fields filtered by category
func GetFieldsByCategory(category string) []FieldDefinition {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	
	var fields []FieldDefinition
	for _, field := range fieldRegistry {
		if field.Category == category {
			fields = append(fields, field)
		}
	}
	
	return fields
}

// BaseValidator provides common functionality for all validators
type BaseValidator struct {
	name    string
	vType   string
	enabled bool
	mu      sync.RWMutex
}

// NewBaseValidator creates a new base validator
func NewBaseValidator(name, vType string) BaseValidator {
	return BaseValidator{
		name:    name,
		vType:   vType,
		enabled: true,
	}
}

// Name returns the validator name
func (bv *BaseValidator) Name() string {
	return bv.name
}

// Type returns the validator type
func (bv *BaseValidator) Type() string {
	return bv.vType
}

// SetEnabled sets the enabled state
func (bv *BaseValidator) SetEnabled(enabled bool) {
	bv.mu.Lock()
	defer bv.mu.Unlock()
	bv.enabled = enabled
}

// IsEnabled returns the enabled state
func (bv *BaseValidator) IsEnabled() bool {
	bv.mu.RLock()
	defer bv.mu.RUnlock()
	return bv.enabled
}

// Common validation options structures
type StringOptions struct {
	MinLength    int
	MaxLength    int
	Required     bool
	Pattern      string // Regex pattern
	AllowEmpty   bool
	TrimSpaces   bool
}

type URLOptions struct {
	RequireHTTPS   bool
	AllowQuery     bool
	AllowFragment  bool
	AllowedDomains []string
	BlockedDomains []string
}

type DateOptions struct {
	RequireFuture bool
	RequirePast   bool
	MinAge        time.Duration
	MaxAge        time.Duration
	AllowedFormats []string
}

type ImageOptions struct {
	RequireHTTPS   bool
	AllowedFormats []string
	MaxFileSize    int64
	MinWidth       int
	MinHeight      int
}

type NumberOptions struct {
	Min          float64
	Max          float64
	RequireInt   bool
	AllowNegative bool
}