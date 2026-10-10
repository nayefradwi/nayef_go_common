package authhttp

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/httputil"
)

type Check func(ctx context.Context, id auth.Identity, r *http.Request) (bool, error)

func Has(check Check) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jw := httputil.NewJsonResponseWriter(w)
			id, ok := auth.GetIdentity(r.Context())
			if !ok {
				jw.WriteError(errors.UnauthorizedError("Identity not found"))
				return
			}

			allowed, err := check(r.Context(), id, r)
			if err != nil {
				slog.ErrorContext(r.Context(), "auth check failed", "error", err)
				jw.WriteError(errors.InternalError("failed to check access"))
				return
			}

			if !allowed {
				jw.WriteError(errors.ForbiddenError("Forbidden"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func HasClaim(name string, values ...string) Check {
	return func(_ context.Context, id auth.Identity, _ *http.Request) (bool, error) {
		return id.HasClaim(name, values...), nil
	}
}

func VerifyClaim(name string, values ...string) func(http.Handler) http.Handler {
	return Has(HasClaim(name, values...))
}

var (
	IsUser Check = isKind(auth.KindUser)
	IsKey  Check = isKind(auth.KindKey)
)

func isKind(kind auth.IdentityKind) Check {
	return func(_ context.Context, id auth.Identity, _ *http.Request) (bool, error) {
		return id.Kind == kind, nil
	}
}

func ByKind(user, key Check) Check {
	return func(ctx context.Context, id auth.Identity, r *http.Request) (bool, error) {
		switch id.Kind {
		case auth.KindUser:
			return user(ctx, id, r)
		case auth.KindKey:
			return key(ctx, id, r)
		}
		return false, nil
	}
}
