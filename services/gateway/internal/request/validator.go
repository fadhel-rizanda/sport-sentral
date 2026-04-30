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
	dec := json.NewDecoder(bytes.NewReader(c.Body()))
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
		return e.Field() + " must be at least " + e.Param() + " characters"
	case "max":
		return e.Field() + " must be at most " + e.Param() + " characters"
	case "uuid":
		return e.Field() + " must be a valid UUID"
	case "oneof":
		return e.Field() + " must be one of: " + e.Param()
	default:
		return e.Field() + " is invalid"
	}
}
