// ABOUTME: Validation configuration system for customizing validation behavior and performance settings
// ABOUTME: Provides validation profiles, custom rules, and performance optimization options

package validation

import (
	"fmt"
	"sync"
	"time"
)

// ValidationConfig holds the global validation configuration
type ValidationConfig struct {
	Enabled            bool
	Profile            string
	CustomRules        map[string]interface{}
	PerformanceMode    string
	ErrorHandling      string
	MaxValidationTime  time.Duration
	ConcurrentValidation bool
	mu                sync.RWMutex
}

// DefaultValidationConfig returns the default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		Enabled:            true,
		Profile:            "lenient",
		CustomRules:        make(map[string]interface{}),
		PerformanceMode:    "fast",
		ErrorHandling:      "collect_all",
		MaxValidationTime:  5 * time.Second,
		ConcurrentValidation: true,
	}
}

// Global configuration instance
var globalConfig = DefaultValidationConfig()
var configMutex sync.RWMutex

// SetGlobalConfig sets the global validation configuration
func SetGlobalConfig(config *ValidationConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig = config
}

// GetGlobalConfig returns the current global validation configuration
func GetGlobalConfig() *ValidationConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	
	// Return a copy to prevent external modification
	return &ValidationConfig{
		Enabled:            globalConfig.Enabled,
		Profile:            globalConfig.Profile,
		CustomRules:        copyMap(globalConfig.CustomRules),
		PerformanceMode:    globalConfig.PerformanceMode,
		ErrorHandling:      globalConfig.ErrorHandling,
		MaxValidationTime:  globalConfig.MaxValidationTime,
		ConcurrentValidation: globalConfig.ConcurrentValidation,
	}
}

// EnableValidation enables or disables validation globally
func EnableValidation(enabled bool) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.Enabled = enabled
}

// IsValidationEnabled returns whether validation is globally enabled
func IsValidationEnabled() bool {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	return globalConfig.Enabled
}

// SetValidationProfile sets the current validation profile
func SetValidationProfile(profileName string) error {
	if _, exists := validationProfiles[profileName]; !exists {
		return fmt.Errorf("validation profile '%s' does not exist", profileName)
	}
	
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.Profile = profileName
	
	// Apply profile settings
	profile := GetValidationProfile(profileName)
	globalConfig.ErrorHandling = profile.ErrorHandling
	globalConfig.PerformanceMode = profile.PerformanceMode
	
	return nil
}

// GetCurrentProfile returns the current validation profile
func GetCurrentProfile() ValidationProfile {
	globalConfig.mu.RLock()
	profileName := globalConfig.Profile
	globalConfig.mu.RUnlock()
	
	return GetValidationProfile(profileName)
}

// ValidationRuleConfig represents configuration for specific validation rules
type ValidationRuleConfig struct {
	FieldName   string
	RuleType    string
	Enabled     bool
	Severity    string // "error", "warning", "info"
	Parameters  map[string]interface{}
	CustomMessage string
}

// FieldValidationConfig holds validation configuration for a specific field
type FieldValidationConfig struct {
	FieldName    string
	Required     bool
	Rules        []ValidationRuleConfig
	Transformers []string // Names of transformers to apply
	Metadata     map[string]interface{}
}

// ValidationConfigBuilder helps build complex validation configurations
type ValidationConfigBuilder struct {
	config *ValidationConfig
	fields map[string]FieldValidationConfig
}

// NewValidationConfigBuilder creates a new configuration builder
func NewValidationConfigBuilder() *ValidationConfigBuilder {
	return &ValidationConfigBuilder{
		config: DefaultValidationConfig(),
		fields: make(map[string]FieldValidationConfig),
	}
}

// WithProfile sets the validation profile
func (vcb *ValidationConfigBuilder) WithProfile(profileName string) *ValidationConfigBuilder {
	vcb.config.Profile = profileName
	return vcb
}

// WithPerformanceMode sets the performance mode
func (vcb *ValidationConfigBuilder) WithPerformanceMode(mode string) *ValidationConfigBuilder {
	vcb.config.PerformanceMode = mode
	return vcb
}

// WithErrorHandling sets the error handling strategy
func (vcb *ValidationConfigBuilder) WithErrorHandling(strategy string) *ValidationConfigBuilder {
	vcb.config.ErrorHandling = strategy
	return vcb
}

// WithTimeout sets the maximum validation time
func (vcb *ValidationConfigBuilder) WithTimeout(timeout time.Duration) *ValidationConfigBuilder {
	vcb.config.MaxValidationTime = timeout
	return vcb
}

// WithCustomRule adds a custom validation rule
func (vcb *ValidationConfigBuilder) WithCustomRule(key string, value interface{}) *ValidationConfigBuilder {
	vcb.config.CustomRules[key] = value
	return vcb
}

// AddFieldConfig adds configuration for a specific field
func (vcb *ValidationConfigBuilder) AddFieldConfig(config FieldValidationConfig) *ValidationConfigBuilder {
	vcb.fields[config.FieldName] = config
	return vcb
}

// Build creates the final validation configuration
func (vcb *ValidationConfigBuilder) Build() (*ValidationConfig, map[string]FieldValidationConfig) {
	return vcb.config, vcb.fields
}

// ValidationMetrics tracks validation performance and statistics
type ValidationMetrics struct {
	TotalValidations  int64
	SuccessfulValidations int64
	FailedValidations int64
	AverageValidationTime time.Duration
	ValidationsByType map[string]int64
	ErrorsByType      map[string]int64
	mu               sync.RWMutex
}

// NewValidationMetrics creates a new metrics tracker
func NewValidationMetrics() *ValidationMetrics {
	return &ValidationMetrics{
		ValidationsByType: make(map[string]int64),
		ErrorsByType:     make(map[string]int64),
	}
}

// RecordValidation records a validation attempt
func (vm *ValidationMetrics) RecordValidation(validationType string, success bool, duration time.Duration) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	
	vm.TotalValidations++
	vm.ValidationsByType[validationType]++
	
	if success {
		vm.SuccessfulValidations++
	} else {
		vm.FailedValidations++
		vm.ErrorsByType[validationType]++
	}
	
	// Update average validation time
	if vm.TotalValidations > 0 {
		totalTime := vm.AverageValidationTime * time.Duration(vm.TotalValidations-1)
		vm.AverageValidationTime = (totalTime + duration) / time.Duration(vm.TotalValidations)
	}
}

// GetMetrics returns a copy of the current metrics
func (vm *ValidationMetrics) GetMetrics() ValidationMetrics {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	
	return ValidationMetrics{
		TotalValidations:     vm.TotalValidations,
		SuccessfulValidations: vm.SuccessfulValidations,
		FailedValidations:    vm.FailedValidations,
		AverageValidationTime: vm.AverageValidationTime,
		ValidationsByType:    copyMapInt64(vm.ValidationsByType),
		ErrorsByType:        copyMapInt64(vm.ErrorsByType),
	}
}

// Reset resets all metrics
func (vm *ValidationMetrics) Reset() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	
	vm.TotalValidations = 0
	vm.SuccessfulValidations = 0
	vm.FailedValidations = 0
	vm.AverageValidationTime = 0
	vm.ValidationsByType = make(map[string]int64)
	vm.ErrorsByType = make(map[string]int64)
}

// Global metrics instance
var globalMetrics = NewValidationMetrics()

// GetGlobalMetrics returns the global validation metrics
func GetGlobalMetrics() ValidationMetrics {
	return globalMetrics.GetMetrics()
}

// RecordGlobalValidation records a validation in global metrics
func RecordGlobalValidation(validationType string, success bool, duration time.Duration) {
	globalMetrics.RecordValidation(validationType, success, duration)
}

// ResetGlobalMetrics resets the global validation metrics
func ResetGlobalMetrics() {
	globalMetrics.Reset()
}

// ValidationContext provides context for validation operations
type ValidationContext struct {
	FieldName    string
	FieldValue   interface{}
	DocumentContext map[string]interface{}
	Config       *ValidationConfig
	Metadata     map[string]interface{}
	StartTime    time.Time
}

// NewValidationContext creates a new validation context
func NewValidationContext(fieldName string, fieldValue interface{}) *ValidationContext {
	return &ValidationContext{
		FieldName:    fieldName,
		FieldValue:   fieldValue,
		DocumentContext: make(map[string]interface{}),
		Config:       GetGlobalConfig(),
		Metadata:     make(map[string]interface{}),
		StartTime:    time.Now(),
	}
}

// WithDocumentContext adds document context to the validation context
func (vc *ValidationContext) WithDocumentContext(key string, value interface{}) *ValidationContext {
	vc.DocumentContext[key] = value
	return vc
}

// WithMetadata adds metadata to the validation context
func (vc *ValidationContext) WithMetadata(key string, value interface{}) *ValidationContext {
	vc.Metadata[key] = value
	return vc
}

// Duration returns the time elapsed since validation started
func (vc *ValidationContext) Duration() time.Duration {
	return time.Since(vc.StartTime)
}

// ContextualValidator extends ValidatorInterface with context support
type ContextualValidator interface {
	ValidatorInterface
	ValidateWithContext(ctx *ValidationContext) error
}

// Helper functions

// copyMap creates a copy of a map[string]interface{}
func copyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for key, value := range original {
		copy[key] = value
	}
	return copy
}

// copyMapInt64 creates a copy of a map[string]int64
func copyMapInt64(original map[string]int64) map[string]int64 {
	copy := make(map[string]int64)
	for key, value := range original {
		copy[key] = value
	}
	return copy
}

// ValidateConfigProfile validates a configuration profile
func ValidateConfigProfile(profile ValidationProfile) error {
	validErrorHandling := map[string]bool{
		"fail_fast":   true,
		"collect_all": true,
		"warn_only":   true,
	}
	
	validPerformanceModes := map[string]bool{
		"fast":     true,
		"thorough": true,
		"balanced": true,
	}
	
	if !validErrorHandling[profile.ErrorHandling] {
		return fmt.Errorf("invalid error handling strategy: %s", profile.ErrorHandling)
	}
	
	if !validPerformanceModes[profile.PerformanceMode] {
		return fmt.Errorf("invalid performance mode: %s", profile.PerformanceMode)
	}
	
	return nil
}

// CreateValidationProfileFromConfig creates a validation profile from a config
func CreateValidationProfileFromConfig(config *ValidationConfig) ValidationProfile {
	return ValidationProfile{
		Name:                 config.Profile,
		EnableAllValidations: config.Enabled,
		ErrorHandling:        config.ErrorHandling,
		PerformanceMode:      config.PerformanceMode,
		CustomRules:          copyMap(config.CustomRules),
	}
}