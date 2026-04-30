package response

import "github.com/gofiber/fiber/v2"

type Response struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	Data    interface{}  `json:"data,omitempty"`
	Error   string       `json:"error,omitempty"`
	Errors  []FieldError `json:"errors,omitempty"`
	Meta    *Meta        `json:"meta,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Meta struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

func OK(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Data:    data,
	})
}

func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Response{
		Success: true,
		Data:    data,
	})
}

func OKWithMessage(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Message: message,
	})
}

func OKWithMeta(c *fiber.Ctx, data interface{}, meta Meta) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Data:    data,
		Meta:    &meta,
	})
}

func Error(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(Response{
		Success: false,
		Error:   message,
	})
}

func ValidationError(c *fiber.Ctx, errors []FieldError) error {
	return c.Status(fiber.StatusBadRequest).JSON(Response{
		Success: false,
		Error:   "validation failed",
		Errors:  errors,
	})
}
