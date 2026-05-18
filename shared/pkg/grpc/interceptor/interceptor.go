package interceptor

import (
	"buf.build/go/protovalidate"
	"context"
	"errors"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"runtime/debug"
	"time"
)

// UnaryLogger logs every unary RPC with duration and status code.
func UnaryLogger(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err)

		log.Info("grpc",
			zap.String("method", info.FullMethod),
			zap.String("code", code.String()),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)
		return resp, err
	}
}

// UnaryRecovery catches panics and returns an Internal gRPC error.
func UnaryRecovery(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					zap.Any("panic", r),
					zap.String("method", info.FullMethod),
					zap.ByteString("stack", debug.Stack()),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// UnaryValidator validates the request message using protovalidate.
func UnaryValidator(v protovalidate.Validator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if msg, ok := req.(proto.Message); ok {
			if err := v.Validate(msg); err != nil {
				st := status.New(codes.InvalidArgument, "invalid input arguments")
				br := &errdetails.BadRequest{}

				var valErr *protovalidate.ValidationError
				if errors.As(err, &valErr) {
					for _, v := range valErr.Violations {
						fieldName := protovalidate.FieldPathString(v.Proto.GetField())

						br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
							Field:       fieldName,
							Description: v.Proto.GetMessage(),
						})
					}
				}

				if detailedStatus, errDetail := st.WithDetails(br); errDetail == nil {
					return nil, detailedStatus.Err()
				}
				return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
			}
		}
		return handler(ctx, req)
	}
}
