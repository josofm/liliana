package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator instance
func New() *Validator {
	v := validator.New()

	// Register custom validators if needed
	// v.RegisterValidation("custom_tag", customValidationFunc)

	return &Validator{
		validate: v,
	}
}

// Validate validates a struct using the validator tags
func (v *Validator) Validate(s interface{}) error {
	return v.validate.Struct(s)
}

// ValidateAndGetErrors returns validation errors in a structured format
func (v *Validator) ValidateAndGetErrors(s interface{}) map[string]string {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			fieldName := v.GetFieldName(fieldError.Field())
			errorMsg := v.formatError(fieldError)
			errors[fieldName] = errorMsg
		}
	}

	return errors
}

// formatError formats a validation error into a user-friendly message
func (v *Validator) formatError(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Minimum %s characters required", fieldError.Param())
	case "max":
		return fmt.Sprintf("Maximum %s characters allowed", fieldError.Param())
	case "oneof":
		return fmt.Sprintf("Value must be one of: %s", fieldError.Param())
	case "gt":
		return fmt.Sprintf("Value must be greater than %s", fieldError.Param())
	case "url":
		return "Invalid URL format"
	default:
		return fmt.Sprintf("Invalid field: %s", fieldError.Tag())
	}
}

// GetFieldName returns the JSON field name from the struct tag
func (v *Validator) GetFieldName(fieldName string) string {
	// This is a simplified version - in a real implementation,
	// you might want to use reflection to get the actual JSON tag
	return strings.ToLower(fieldName)
}
