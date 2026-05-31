package stdlib

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// Validator provides request validation
type Validator struct {
	errors []ValidationError
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		errors: make([]ValidationError, 0),
	}
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors
func (v *Validator) Errors() []ValidationError {
	return v.errors
}

// ErrorMessages returns all error messages as strings
func (v *Validator) ErrorMessages() []string {
	messages := make([]string, len(v.errors))
	for i, err := range v.errors {
		messages[i] = fmt.Sprintf("%s: %s", err.Field, err.Message)
	}
	return messages
}

// ErrorJSON returns errors as a JSON-serializable map
func (v *Validator) ErrorJSON() map[string]interface{} {
	errors := make(map[string]interface{})
	for _, err := range v.errors {
		errors[err.Field] = err.Message
	}
	return map[string]interface{}{
		"valid":  false,
		"errors": errors,
	}
}

// Clear removes all validation errors
func (v *Validator) Clear() {
	v.errors = make([]ValidationError, 0)
}

// addError adds a validation error
func (v *Validator) addError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// Required validates that a field is not empty
func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.addError(field, "is required")
	}
	return v
}

// MinLength validates minimum string length
func (v *Validator) MinLength(field, value string, min int) *Validator {
	if len(value) < min {
		v.addError(field, fmt.Sprintf("must be at least %d characters", min))
	}
	return v
}

// MaxLength validates maximum string length
func (v *Validator) MaxLength(field, value string, max int) *Validator {
	if len(value) > max {
		v.addError(field, fmt.Sprintf("must be at most %d characters", max))
	}
	return v
}

// Length validates exact string length
func (v *Validator) Length(field, value string, length int) *Validator {
	if len(value) != length {
		v.addError(field, fmt.Sprintf("must be exactly %d characters", length))
	}
	return v
}

// Email validates email format
func (v *Validator) Email(field, value string) *Validator {
	if value == "" {
		return v
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		v.addError(field, "must be a valid email address")
	}
	return v
}

// Pattern validates against a regex pattern
func (v *Validator) Pattern(field, value, pattern string) *Validator {
	if value == "" {
		return v
	}
	regex, err := regexp.Compile(pattern)
	if err != nil {
		v.addError(field, "invalid validation pattern")
		return v
	}
	if !regex.MatchString(value) {
		v.addError(field, "format is invalid")
	}
	return v
}

// Numeric validates that a string is numeric
func (v *Validator) Numeric(field, value string) *Validator {
	if value == "" {
		return v
	}
	_, err := strconv.ParseFloat(value, 64)
	if err != nil {
		v.addError(field, "must be a number")
	}
	return v
}

// Integer validates that a string is an integer
func (v *Validator) Integer(field, value string) *Validator {
	if value == "" {
		return v
	}
	_, err := strconv.Atoi(value)
	if err != nil {
		v.addError(field, "must be an integer")
	}
	return v
}

// Min validates minimum value for integers
func (v *Validator) Min(field, value string, min int) *Validator {
	if value == "" {
		return v
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		v.addError(field, "must be a number")
		return v
	}
	if num < min {
		v.addError(field, fmt.Sprintf("must be at least %d", min))
	}
	return v
}

// Max validates maximum value for integers
func (v *Validator) Max(field, value string, max int) *Validator {
	if value == "" {
		return v
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		v.addError(field, "must be a number")
		return v
	}
	if num > max {
		v.addError(field, fmt.Sprintf("must be at most %d", max))
	}
	return v
}

// Between validates that a value is between min and max
func (v *Validator) Between(field, value string, min, max int) *Validator {
	if value == "" {
		return v
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		v.addError(field, "must be a number")
		return v
	}
	if num < min || num > max {
		v.addError(field, fmt.Sprintf("must be between %d and %d", min, max))
	}
	return v
}

// In validates that a value is in a list of allowed values
func (v *Validator) In(field, value string, allowed []string) *Validator {
	if value == "" {
		return v
	}
	for _, a := range allowed {
		if value == a {
			return v
		}
	}
	v.addError(field, "is not a valid option")
	return v
}

// NotIn validates that a value is not in a list of forbidden values
func (v *Validator) NotIn(field, value string, forbidden []string) *Validator {
	if value == "" {
		return v
	}
	for _, f := range forbidden {
		if value == f {
			v.addError(field, "is not allowed")
			return v
		}
	}
	return v
}

// Alpha validates that a string contains only letters
func (v *Validator) Alpha(field, value string) *Validator {
	if value == "" {
		return v
	}
	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	if !alphaRegex.MatchString(value) {
		v.addError(field, "must contain only letters")
	}
	return v
}

// AlphaNumeric validates that a string contains only letters and numbers
func (v *Validator) AlphaNumeric(field, value string) *Validator {
	if value == "" {
		return v
	}
	alphaNumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphaNumRegex.MatchString(value) {
		v.addError(field, "must contain only letters and numbers")
	}
	return v
}

// URL validates URL format
func (v *Validator) URL(field, value string) *Validator {
	if value == "" {
		return v
	}
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(value) {
		v.addError(field, "must be a valid URL")
	}
	return v
}

// UUID validates UUID format
func (v *Validator) UUID(field, value string) *Validator {
	if value == "" {
		return v
	}
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(value) {
		v.addError(field, "must be a valid UUID")
	}
	return v
}

// Match validates that two fields match
func (v *Validator) Match(field1, value1, field2, value2 string) *Validator {
	if value1 != value2 {
		v.addError(field1, fmt.Sprintf("must match %s", field2))
	}
	return v
}

// Custom adds a custom validation function
func (v *Validator) Custom(field string, isValid bool, message string) *Validator {
	if !isValid {
		v.addError(field, message)
	}
	return v
}

// RequestValidator provides convenient request validation
type RequestValidator struct {
	validator *Validator
	req       *TRequest
}

// NewRequestValidator creates a validator for a request
func NewRequestValidator(req *TRequest) *RequestValidator {
	return &RequestValidator{
		validator: NewValidator(),
		req:       req,
	}
}

// Validate validates the request and returns the validator
func (rv *RequestValidator) Validate() *Validator {
	return rv.validator
}

// Required validates that a request field is present and not empty
func (rv *RequestValidator) Required(fields ...string) *RequestValidator {
	for _, field := range fields {
		value := rv.req.GetField(field)
		rv.validator.Required(field, value)
	}
	return rv
}

// Email validates an email field
func (rv *RequestValidator) Email(field string) *RequestValidator {
	value := rv.req.GetField(field)
	rv.validator.Email(field, value)
	return rv
}

// MinLength validates minimum length
func (rv *RequestValidator) MinLength(field string, min int) *RequestValidator {
	value := rv.req.GetField(field)
	rv.validator.MinLength(field, value, min)
	return rv
}

// MaxLength validates maximum length
func (rv *RequestValidator) MaxLength(field string, max int) *RequestValidator {
	value := rv.req.GetField(field)
	rv.validator.MaxLength(field, value, max)
	return rv
}

// Pattern validates against a pattern
func (rv *RequestValidator) Pattern(field, pattern string) *RequestValidator {
	value := rv.req.GetField(field)
	rv.validator.Pattern(field, value, pattern)
	return rv
}

// In validates against allowed values
func (rv *RequestValidator) In(field string, allowed []string) *RequestValidator {
	value := rv.req.GetField(field)
	rv.validator.In(field, value, allowed)
	return rv
}

// IsValid returns true if validation passed
func (rv *RequestValidator) IsValid() bool {
	return !rv.validator.HasErrors()
}

// Errors returns validation errors
func (rv *RequestValidator) Errors() []ValidationError {
	return rv.validator.Errors()
}
