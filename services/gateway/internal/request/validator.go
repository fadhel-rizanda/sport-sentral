package request

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"microservice-golang/services/gateway/internal/response"
)

// ValidationError represents an error containing multiple field-level validation violations.
type ValidationError struct {
	Errors []response.FieldError
}

func (v *ValidationError) Error() string {
	return "validation failed"
}

// ValidationBag provides a fluent builder to collect field-level validation errors.
type ValidationBag struct {
	errors []response.FieldError
}

// NewValidationBag creates a new empty validation error collector.
func NewValidationBag() *ValidationBag {
	return &ValidationBag{
		errors: make([]response.FieldError, 0),
	}
}

// Add appends a field violation to the bag.
func (b *ValidationBag) Add(field, message string) {
	b.errors = append(b.errors, response.FieldError{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if any field errors have been added.
func (b *ValidationBag) HasErrors() bool {
	return len(b.errors) > 0
}

// Error returns *ValidationError if errors exist, otherwise nil.
func (b *ValidationBag) Error() error {
	if !b.HasErrors() {
		return nil
	}
	return &ValidationError{Errors: b.errors}
}

// NewValidationError creates a ValidationError containing one or more FieldErrors.
func NewValidationError(fieldErrors ...response.FieldError) error {
	return &ValidationError{Errors: fieldErrors}
}

// NewFieldError creates a ValidationError for a single custom field violation.
func NewFieldError(field, message string) error {
	return &ValidationError{
		Errors: []response.FieldError{
			{
				Field:   field,
				Message: message,
			},
		},
	}
}

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			name = strings.SplitN(fld.Tag.Get("query"), ",", 2)[0]
			if name == "-" || name == "" {
				name = strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
				if name == "-" || name == "" {
					return fld.Name
				}
			}
		}
		return name
	})
}

// ValidateStruct validates any struct using validator tags and returns *ValidationError if invalid.
func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		var fieldErrors []response.FieldError
		if valErrs, ok := err.(validator.ValidationErrors); ok {
			for _, e := range valErrs {
				fieldErrors = append(fieldErrors, response.FieldError{
					Field:   e.Field(),
					Message: messageForTag(e),
				})
			}
		} else {
			fieldErrors = append(fieldErrors, response.FieldError{
				Field:   "general",
				Message: err.Error(),
			})
		}
		return &ValidationError{Errors: fieldErrors}
	}
	return nil
}

// Parse parses JSON body and runs struct validation.
func Parse(c *fiber.Ctx, body interface{}) error {
	rawBody := c.Body()
	if len(rawBody) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "request body cannot be empty")
	}

	dec := json.NewDecoder(bytes.NewReader(rawBody))
	dec.DisallowUnknownFields()

	if err := dec.Decode(body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "unknown or malformed fields in request")
	}

	return ValidateStruct(body)
}

// ParseQuery parses query parameters into a struct and runs struct validation.
func ParseQuery(c *fiber.Ctx, query interface{}) error {
	if err := c.QueryParser(query); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "malformed query parameters")
	}
	return ValidateStruct(query)
}

func messageForTag(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email format"
	case "min":
		return "must be at least " + e.Param()
	case "max":
		return "must be at most " + e.Param()
	case "gt":
		return "must be greater than " + e.Param()
	case "gte":
		return "must be greater than or equal to " + e.Param()
	case "lt":
		return "must be less than " + e.Param()
	case "lte":
		return "must be less than or equal to " + e.Param()
	case "uuid":
		return "must be a valid UUID"
	case "oneof":
		return "must be one of: " + e.Param()
	case "numeric":
		return "must be a numeric value"
	case "alphanum":
		return "must contain only alphanumeric characters"
	case "url":
		return "must be a valid URL"
	default:
		if e.Param() != "" {
			return "failed validation on tag '" + e.Tag() + "' with condition '" + e.Param() + "'"
		}
		return "is invalid"
	}
}
