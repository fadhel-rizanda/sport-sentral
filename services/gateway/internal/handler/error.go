package handler

import (
	"github.com/gofiber/fiber/v2"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"microservice-golang/services/gateway/internal/response"
)

func grpcError(c *fiber.Ctx, err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return response.Error(c, 500, "internal error")
	}

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

	switch st.Code() {
	case codes.NotFound:
		return response.Error(c, fiber.StatusNotFound, st.Message())
	case codes.AlreadyExists:
		return response.Error(c, fiber.StatusConflict, st.Message())
	case codes.InvalidArgument, codes.Unknown:
		return response.Error(c, fiber.StatusBadRequest, st.Message())
	case codes.Unauthenticated:
		return response.Error(c, fiber.StatusUnauthorized, st.Message())
	case codes.PermissionDenied:
		return response.Error(c, fiber.StatusForbidden, st.Message())
	default:
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}
}
