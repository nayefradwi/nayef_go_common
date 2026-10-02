package connectutil

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

type ErrorOptions struct {
	ErrorListener OnErrorListener
}

func ServerErrors(opts ...ErrorOptions) connect.UnaryInterceptorFunc {
	opt := ErrorOptions{ErrorListener: logError}
	if len(opts) > 0 {
		opt = opts[0]
	}

	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			res, err := next(ctx, req)
			if err == nil {
				return res, nil
			}

			opt.ErrorListener(err)
			return res, toServerError(err)
		}
	}
}

func ClientErrors() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			res, err := next(ctx, req)
			if err == nil {
				return res, nil
			}

			connectErr, ok := errors.AsType[*connect.Error](err)
			if !ok {
				return res, err
			}

			return res, ToResultError(connectErr)
		}
	}
}

func toServerError(err error) error {
	if connectErr, ok := errors.AsType[*connect.Error](err); ok {
		return connectErr
	}

	if resultErr, ok := errors.AsType[*ResultError](err); ok {
		return ToConnectError(resultErr)
	}

	return ToConnectError(InternalError("internal server error"))
}

func logError(err error) {
	slog.Error("connect handler error", "error", err.Error())
}
