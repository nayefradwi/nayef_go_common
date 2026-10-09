package authconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
)

type unaryOnly struct {
	name       string
	streamCode connect.Code
	check      func(ctx context.Context, req connect.AnyRequest) (context.Context, error)
}

func (u unaryOnly) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			return next(ctx, req)
		}

		ctx, err := u.check(ctx, req)
		if err != nil {
			return nil, err
		}

		return next(ctx, req)
	}
}

func (u unaryOnly) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (u unaryOnly) WrapStreamingHandler(connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(context.Context, connect.StreamingHandlerConn) error {
		return connect.NewError(u.streamCode, errors.New(u.name+" supports unary RPCs only"))
	}
}
