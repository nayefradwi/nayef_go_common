package grpc

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func RecoverUnary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = handleRecoveredPanic(r)
			}
		}()
		return handler(ctx, req)
	}
}

func handleRecoveredPanic(r any) error {
	if err, ok := r.(error); ok {
		zap.L().Error("internal server error", zap.Any("error", err), zap.Stack("stack trace"))
		return NewGrpcResponseWriter[any]().WriteError(err)
	}

	zap.L().Error("internal server error", zap.Any("error", r), zap.Stack("stack trace"))
	return NewGrpcResponseWriter[any]().WriteError(fmt.Errorf("%v", r))
}
