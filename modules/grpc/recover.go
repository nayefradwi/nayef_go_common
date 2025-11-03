package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func RecoverUnary() grpc.UnaryServerInterceptor {
	return recoverInterceptor
}

func recoverInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	defer func() { err = recoverError() }()
	return handler(ctx, req)
}

func recoverError() error {
	recovered := recover()
	if err, ok := recovered.(error); ok {
		zap.L().Error("internal server error", zap.Any("error", err), zap.Stack("stack trace"))
		return NewGrpcResponseWriter[any]().WriteError(err)
	}

	return nil
}
