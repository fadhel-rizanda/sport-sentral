package errors

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---- Domain error types ----

type NotFoundError struct{ Resource string }
type ConflictError struct{ Resource string }
type ValidationError struct{ Field, Reason string }
type InvalidArgumentError struct{ Reason string }
type UnauthorizedError struct{ Reason string }
type ForbiddenError struct{ Reason string }
type InternalError struct{ Err error }

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}
func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s already exists", e.Resource)
}
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on field '%s': %s", e.Field, e.Reason)
}
func (e *InvalidArgumentError) Error() string { return e.Reason }
func (e *UnauthorizedError) Error() string    { return e.Reason }
func (e *ForbiddenError) Error() string       { return e.Reason }
func (e *InternalError) Error() string        { return e.Err.Error() }
func (e *InternalError) Unwrap() error        { return e.Err }

// ---- Constructors ----

func NotFound(resource string) error        { return &NotFoundError{Resource: resource} }
func Conflict(resource string) error        { return &ConflictError{Resource: resource} }
func Validation(field, reason string) error { return &ValidationError{Field: field, Reason: reason} }
func InvalidArgument(reason string) error   { return &InvalidArgumentError{Reason: reason} }
func Unauthorized(reason string) error      { return &UnauthorizedError{Reason: reason} }
func Forbidden(reason string) error         { return &ForbiddenError{Reason: reason} }
func Internal(err error) error              { return &InternalError{Err: err} }

// ---- gRPC mapping ----

func ToGRPC(err error) error {
	if err == nil {
		return nil
	}

	var notFound *NotFoundError
	var conflict *ConflictError
	var validation *ValidationError
	var invalidArg *InvalidArgumentError
	var unauthorized *UnauthorizedError
	var forbidden *ForbiddenError
	var internal *InternalError

	switch {
	case errors.As(err, &notFound):
		return status.Errorf(codes.NotFound, err.Error())
	case errors.As(err, &conflict):
		return status.Errorf(codes.AlreadyExists, err.Error())
	case errors.As(err, &validation):
		return status.Errorf(codes.InvalidArgument, err.Error())
	case errors.As(err, &invalidArg):
		return status.Errorf(codes.InvalidArgument, err.Error())
	case errors.As(err, &unauthorized):
		return status.Errorf(codes.Unauthenticated, err.Error())
	case errors.As(err, &forbidden):
		return status.Errorf(codes.PermissionDenied, err.Error())
	case errors.As(err, &internal):
		return status.Errorf(codes.Internal, "internal server error")
	default:
		return status.Errorf(codes.Internal, "internal server error")
	}
}
