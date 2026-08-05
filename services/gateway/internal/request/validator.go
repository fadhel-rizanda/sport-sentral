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

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
}

func Parse(c *fiber.Ctx, body interface{}) error {
	rawBody := c.Body()
	if len(rawBody) == 0 {
		return response.Error(c, fiber.StatusBadRequest, "request body cannot be empty")
	}

	dec := json.NewDecoder(bytes.NewReader(rawBody))
	dec.DisallowUnknownFields()

	if err := dec.Decode(body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "unknown or malformed fields in request")
	}

	if err := validate.Struct(body); err != nil {
		var fieldErrors []response.FieldError
		for _, e := range err.(validator.ValidationErrors) {
			fieldErrors = append(fieldErrors, response.FieldError{
				Field:   e.Field(),
				Message: messageForTag(e),
			})
		}
		return response.ValidationError(c, fieldErrors)
	}

	return nil
}

func messageForTag(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "email":
		return "invalid email format"
	case "min":
		return e.Field() + " must be at least " + e.Param()
	case "max":
		return e.Field() + " must be at most " + e.Param()
	case "gt":
		return e.Field() + " must be greater than " + e.Param()
	case "gte":
		return e.Field() + " must be greater than or equal to " + e.Param()
	case "lt":
		return e.Field() + " must be less than " + e.Param()
	case "lte":
		return e.Field() + " must be less than or equal to " + e.Param()
	case "uuid":
		return e.Field() + " must be a valid UUID"
	case "oneof":
		return e.Field() + " must be one of: " + e.Param()
	case "numeric":
		return e.Field() + " must be a numeric value"
	case "alphanum":
		return e.Field() + " must contain only alphanumeric characters"
	case "url":
		return e.Field() + " must be a valid URL"
	default:
		return e.Field() + " is invalid"
	}
}
