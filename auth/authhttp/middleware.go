package authhttp

import (
	"net/http"

	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/httputil"
)

type JwtMiddleware struct {
	TokenProvider auth.ITokenProvider
}

type OpaqueMiddleware struct {
	Verifier auth.AccessVerifier
}

func NewJwtMiddleware(tokenProvider auth.ITokenProvider) JwtMiddleware {
	return JwtMiddleware{
		TokenProvider: tokenProvider,
	}
}

func NewOpaqueMiddleware(verifier auth.AccessVerifier) OpaqueMiddleware {
	return OpaqueMiddleware{
		Verifier: verifier,
	}
}

func (m JwtMiddleware) UseAuthentication(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jw := httputil.NewJsonResponseWriter(w)
		token := httputil.GetBearerToken(r)
		if token == "" {
			jw.WriteError(errors.UnauthorizedError("Token not found"))
			return
		}

		accessToken, err := m.TokenProvider.GetClaims(token)
		if err != nil || accessToken.IsExpired() {
			jw.WriteError(errors.UnauthorizedError("Invalid token"))
			return
		}

		ctx := auth.WithIdentity(r.Context(), auth.IdentityFromToken(accessToken))
		r = r.WithContext(ctx)
		f.ServeHTTP(w, r)
	})
}

func (m OpaqueMiddleware) UseAuthentication(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jw := httputil.NewJsonResponseWriter(w)
		token := httputil.GetBearerToken(r)
		if token == "" {
			jw.WriteError(errors.UnauthorizedError("Token not found"))
			return
		}

		accessToken, err := m.Verifier.VerifyAccess(r.Context(), token)
		if err != nil || accessToken.IsExpired() {
			jw.WriteError(errors.UnauthorizedError("Invalid token"))
			return
		}

		ctx := auth.WithIdentity(r.Context(), auth.IdentityFromToken(accessToken))
		r = r.WithContext(ctx)
		f.ServeHTTP(w, r)
	})
}
