package authconnect

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net"
	"strconv"

	"connectrpc.com/connect"
	"github.com/nayefradwi/nayef_go_common/auth"
)

type KeyFunc func(ctx context.Context, req connect.AnyRequest) (string, error)

func RateLimit(limiter auth.Limiter, key KeyFunc) connect.Interceptor {
	return unaryOnly{
		name:       "RateLimit",
		streamCode: connect.CodeUnimplemented,
		check: func(ctx context.Context, req connect.AnyRequest) (context.Context, error) {
			k, err := key(ctx, req)
			if err != nil {
				return ctx, err
			}

			allowed, retryAfter, err := limiter.Allow(ctx, k)
			if err != nil {
				slog.ErrorContext(ctx, "rate limit check failed", "error", err)
				return ctx, connect.NewError(connect.CodeInternal, errors.New("failed to check rate limit"))
			}

			if !allowed {
				ce := connect.NewError(connect.CodeResourceExhausted, errors.New("Too many requests"))
				ce.Meta().Set("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
				return ctx, ce
			}

			return ctx, nil
		},
	}
}

func ByIP(prefix string) KeyFunc {
	return func(_ context.Context, req connect.AnyRequest) (string, error) {
		addr := req.Peer().Addr
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
		}
		return prefix + ":" + host, nil
	}
}
