package authconnect

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/nayefradwi/nayef_go_common/auth"
)

type Check func(ctx context.Context, id auth.Identity, req connect.AnyRequest) (bool, error)

func Has(check Check) connect.Interceptor {
	return unaryOnly{
		name:       "Has",
		streamCode: connect.CodePermissionDenied,
		check: func(ctx context.Context, req connect.AnyRequest) (context.Context, error) {
			id, ok := auth.GetIdentity(ctx)
			if !ok {
				return ctx, unauthenticated("Identity not found")
			}

			allowed, err := check(ctx, id, req)
			if err != nil {
				slog.ErrorContext(ctx, "auth check failed", "error", err)
				return ctx, connect.NewError(connect.CodeInternal, errors.New("failed to check access"))
			}
			if !allowed {
				return ctx, connect.NewError(connect.CodePermissionDenied, errors.New("Forbidden"))
			}
			return ctx, nil
		},
	}
}

func HasClaim(name string, values ...string) Check {
	return func(_ context.Context, id auth.Identity, _ connect.AnyRequest) (bool, error) {
		return id.HasClaim(name, values...), nil
	}
}

func VerifyClaim(name string, values ...string) connect.Interceptor {
	return Has(HasClaim(name, values...))
}

var (
	IsUser Check = isKind(auth.KindUser)
	IsKey  Check = isKind(auth.KindKey)
)

func isKind(kind auth.IdentityKind) Check {
	return func(_ context.Context, id auth.Identity, _ connect.AnyRequest) (bool, error) {
		return id.Kind == kind, nil
	}
}

func ByKind(user, key Check) Check {
	return func(ctx context.Context, id auth.Identity, req connect.AnyRequest) (bool, error) {
		switch id.Kind {
		case auth.KindUser:
			return user(ctx, id, req)
		case auth.KindKey:
			return key(ctx, id, req)
		}
		return false, nil
	}
}
