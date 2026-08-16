package handler

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
)

func ErrorHandler(logger *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var valErr *request.ValidationError
		if errors.As(err, &valErr) {
			return response.ValidationError(c, valErr.Errors)
		}

		logger.Error("request error",
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
			zap.Error(err),
		)

		if st, ok := status.FromError(err); ok {
			return handleGrpcError(c, st)
		}

		var e *fiber.Error
		if errors.As(err, &e) {
			return response.Error(c, e.Code, e.Message)
		}

		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}
}

func handleGrpcError(c *fiber.Ctx, st *status.Status) error {
	for _, detail := range st.Details() {
		switch t := detail.(type) {
		case *errdetails.BadRequest:
			var fieldErrors []response.FieldError
			for _, v := range t.GetFieldViolations() {
				fieldErrors = append(fieldErrors, response.FieldError{
					Field:   v.GetField(),
					Message: v.GetDescription(),
				})
			}
			return response.ValidationError(c, fieldErrors)
		}
	}

	var httpStatus int
	switch st.Code() {
	case codes.NotFound:
		httpStatus = fiber.StatusNotFound
	case codes.AlreadyExists:
		httpStatus = fiber.StatusConflict
	case codes.InvalidArgument:
		httpStatus = fiber.StatusBadRequest
	case codes.Unauthenticated:
		httpStatus = fiber.StatusUnauthorized
	case codes.PermissionDenied:
		httpStatus = fiber.StatusForbidden
	default:
		httpStatus = fiber.StatusInternalServerError
	}

	return response.Error(c, httpStatus, st.Message())
}
