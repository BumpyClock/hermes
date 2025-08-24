// ABOUTME: Type-specific validator implementations for string, URL, date, image, and number validation
// ABOUTME: Provides concrete validator types that implement the ValidatorInterface

package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	
	"github.com/BumpyClock/hermes/internal/utils/security"
)

// StringValidator validates string fields
type StringValidator struct {
	BaseValidator
	options StringOptions
	pattern *regexp.Regexp
}

// NewStringValidator creates a new string validator
func NewStringValidator(options StringOptions) *StringValidator {
	sv := &StringValidator{
		BaseValidator: NewBaseValidator("string", "string"),
		options:       options,
	}
	
	if options.Pattern != "" {
		if pattern, err := regexp.Compile(options.Pattern); err == nil {
			sv.pattern = pattern
		}
	}
	
	return sv
}

// Validate validates a string value
func (sv *StringValidator) Validate(value interface{}) error {
	if !sv.IsEnabled() {
		return nil
	}
	
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}
	
	if sv.options.TrimSpaces {
		str = strings.TrimSpace(str)
	}
	
	// Check required constraint
	if sv.options.Required && str == "" {
		return fmt.Errorf("field is required but is empty")
	}
	
	// Check empty constraint
	if !sv.options.AllowEmpty && str == "" {
		return fmt.Errorf("empty strings are not allowed")
	}
	
	// Check length constraints
	if sv.options.MinLength > 0 && len(str) < sv.options.MinLength {
		return fmt.Errorf("string length %d is below minimum %d", len(str), sv.options.MinLength)
	}
	
	if sv.options.MaxLength > 0 && len(str) > sv.options.MaxLength {
		return fmt.Errorf("string length %d exceeds maximum %d", len(str), sv.options.MaxLength)
	}
	
	// Check pattern constraint
	if sv.pattern != nil && !sv.pattern.MatchString(str) {
		return fmt.Errorf("string does not match required pattern")
	}
	
	return nil
}

// URLValidator validates URL fields
type URLValidator struct {
	BaseValidator
	options URLOptions
}

// NewURLValidator creates a new URL validator
func NewURLValidator(options URLOptions) *URLValidator {
	return &URLValidator{
		BaseValidator: NewBaseValidator("url", "url"),
		options:       options,
	}
}

// Validate validates a URL value
func (uv *URLValidator) Validate(value interface{}) error {
	if !uv.IsEnabled() {
		return nil
	}
	
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string URL, got %T", value)
	}
	
	if str == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	
	// Parse the URL
	parsedURL, err := url.Parse(str)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	
	// Use security validator for additional checks
	if err := security.ValidateURL(str); err != nil {
		return fmt.Errorf("URL security validation failed: %w", err)
	}
	
	// Check HTTPS requirement
	if uv.options.RequireHTTPS && parsedURL.Scheme != "https" {
		return fmt.Errorf("HTTPS is required, got %s", parsedURL.Scheme)
	}
	
	// Check query parameter constraint
	if !uv.options.AllowQuery && parsedURL.RawQuery != "" {
		return fmt.Errorf("query parameters are not allowed")
	}
	
	// Check fragment constraint
	if !uv.options.AllowFragment && parsedURL.Fragment != "" {
		return fmt.Errorf("URL fragments are not allowed")
	}
	
	// Check domain allowlist
	if len(uv.options.AllowedDomains) > 0 {
		domainAllowed := false
		for _, domain := range uv.options.AllowedDomains {
			if strings.HasSuffix(parsedURL.Host, domain) {
				domainAllowed = true
				break
			}
		}
		if !domainAllowed {
			return fmt.Errorf("domain %s is not in allowlist", parsedURL.Host)
		}
	}
	
	// Check domain blocklist
	for _, domain := range uv.options.BlockedDomains {
		if strings.HasSuffix(parsedURL.Host, domain) {
			return fmt.Errorf("domain %s is blocked", parsedURL.Host)
		}
	}
	
	return nil
}

// DateValidator validates date fields
type DateValidator struct {
	BaseValidator
	options DateOptions
}

// NewDateValidator creates a new date validator
func NewDateValidator(options DateOptions) *DateValidator {
	return &DateValidator{
		BaseValidator: NewBaseValidator("date", "date"),
		options:       options,
	}
}

// Validate validates a date value
func (dv *DateValidator) Validate(value interface{}) error {
	if !dv.IsEnabled() {
		return nil
	}
	
	var parsedTime time.Time
	var err error
	
	switch v := value.(type) {
	case string:
		parsedTime, err = dv.parseDate(v)
		if err != nil {
			return fmt.Errorf("failed to parse date string: %w", err)
		}
	case time.Time:
		parsedTime = v
	default:
		return fmt.Errorf("expected string or time.Time, got %T", value)
	}
	
	now := time.Now()
	
	// Check future requirement
	if dv.options.RequireFuture && parsedTime.Before(now) {
		return fmt.Errorf("date must be in the future")
	}
	
	// Check past requirement
	if dv.options.RequirePast && parsedTime.After(now) {
		return fmt.Errorf("date must be in the past")
	}
	
	// Check minimum age
	if dv.options.MinAge > 0 {
		minTime := now.Add(-dv.options.MinAge)
		if parsedTime.After(minTime) {
			return fmt.Errorf("date is too recent (minimum age: %v)", dv.options.MinAge)
		}
	}
	
	// Check maximum age
	if dv.options.MaxAge > 0 {
		maxTime := now.Add(-dv.options.MaxAge)
		if parsedTime.Before(maxTime) {
			return fmt.Errorf("date is too old (maximum age: %v)", dv.options.MaxAge)
		}
	}
	
	return nil
}

// parseDate attempts to parse a date string using various formats
func (dv *DateValidator) parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"January 2, 2006",
		"Jan 2, 2006",
		"2006/01/02",
		"01/02/2006",
		"02-01-2006",
		"2006-01-02T15:04:05.000Z",
	}
	
	// Add custom formats if specified
	if len(dv.options.AllowedFormats) > 0 {
		formats = append(dv.options.AllowedFormats, formats...)
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date with any known format")
}

// ImageValidator validates image URL fields
type ImageValidator struct {
	BaseValidator
	options ImageOptions
}

// NewImageValidator creates a new image validator
func NewImageValidator(options ImageOptions) *ImageValidator {
	return &ImageValidator{
		BaseValidator: NewBaseValidator("image", "image"),
		options:       options,
	}
}

// Validate validates an image URL
func (iv *ImageValidator) Validate(value interface{}) error {
	if !iv.IsEnabled() {
		return nil
	}
	
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string URL, got %T", value)
	}
	
	if str == "" {
		return fmt.Errorf("image URL cannot be empty")
	}
	
	// Parse the URL
	parsedURL, err := url.Parse(str)
	if err != nil {
		return fmt.Errorf("invalid image URL format: %w", err)
	}
	
	// Check HTTPS requirement
	if iv.options.RequireHTTPS && parsedURL.Scheme != "https" {
		return fmt.Errorf("HTTPS is required for images, got %s", parsedURL.Scheme)
	}
	
	// Check file extension for format validation
	if len(iv.options.AllowedFormats) > 0 {
		format := iv.getImageFormat(parsedURL.Path)
		if format == "" {
			return fmt.Errorf("cannot determine image format from URL")
		}
		
		formatAllowed := false
		for _, allowedFormat := range iv.options.AllowedFormats {
			if strings.EqualFold(format, allowedFormat) {
				formatAllowed = true
				break
			}
		}
		
		if !formatAllowed {
			return fmt.Errorf("image format %s is not allowed", format)
		}
	}
	
	// Note: File size, width, and height validation would require downloading
	// the image, which is beyond the scope of URL validation. These could be
	// implemented as separate validators that work with actual image data.
	
	return nil
}

// getImageFormat extracts the image format from a URL path
func (iv *ImageValidator) getImageFormat(path string) string {
	if lastDot := strings.LastIndex(path, "."); lastDot != -1 && lastDot < len(path)-1 {
		return strings.ToLower(path[lastDot+1:])
	}
	return ""
}

// NumberValidator validates numeric fields
type NumberValidator struct {
	BaseValidator
	options NumberOptions
}

// NewNumberValidator creates a new number validator
func NewNumberValidator(options NumberOptions) *NumberValidator {
	return &NumberValidator{
		BaseValidator: NewBaseValidator("number", "number"),
		options:       options,
	}
}

// Validate validates a numeric value
func (nv *NumberValidator) Validate(value interface{}) error {
	if !nv.IsEnabled() {
		return nil
	}
	
	var num float64
	var err error
	
	switch v := value.(type) {
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case int32:
		num = float64(v)
	case float32:
		num = float64(v)
	case float64:
		num = v
	case string:
		num, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("failed to parse number from string: %w", err)
		}
	default:
		return fmt.Errorf("expected numeric type, got %T", value)
	}
	
	// Check integer requirement
	if nv.options.RequireInt && num != float64(int64(num)) {
		return fmt.Errorf("integer value required, got float")
	}
	
	// Check negative value constraint
	if !nv.options.AllowNegative && num < 0 {
		return fmt.Errorf("negative values are not allowed")
	}
	
	// Check minimum value
	if num < nv.options.Min {
		return fmt.Errorf("value %.2f is below minimum %.2f", num, nv.options.Min)
	}
	
	// Check maximum value
	if num > nv.options.Max {
		return fmt.Errorf("value %.2f exceeds maximum %.2f", num, nv.options.Max)
	}
	
	return nil
}

// CustomValidator allows for domain-specific validation rules
type CustomValidator struct {
	BaseValidator
	validationFunc func(interface{}) error
}

// NewCustomValidator creates a validator with a custom validation function
func NewCustomValidator(name, vType string, validationFunc func(interface{}) error) *CustomValidator {
	return &CustomValidator{
		BaseValidator:  NewBaseValidator(name, vType),
		validationFunc: validationFunc,
	}
}

// Validate validates using the custom validation function
func (cv *CustomValidator) Validate(value interface{}) error {
	if !cv.IsEnabled() {
		return nil
	}
	
	return cv.validationFunc(value)
}