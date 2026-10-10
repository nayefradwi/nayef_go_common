package authhttp

import (
	"context"
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

type ApiKeyMiddleware struct {
	Verifier auth.KeyVerifier
}

type EitherMiddleware struct {
	Users auth.AccessVerifier
	Keys  auth.ApiKeyManager
}

type JwtEitherMiddleware struct {
	TokenProvider auth.ITokenProvider
	Keys          auth.ApiKeyManager
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

func NewApiKeyMiddleware(verifier auth.KeyVerifier) ApiKeyMiddleware {
	return ApiKeyMiddleware{
		Verifier: verifier,
	}
}

func NewEitherMiddleware(users auth.AccessVerifier, keys auth.ApiKeyManager) EitherMiddleware {
	return EitherMiddleware{
		Users: users,
		Keys:  keys,
	}
}

func NewJwtEitherMiddleware(tokenProvider auth.ITokenProvider, keys auth.ApiKeyManager) JwtEitherMiddleware {
	return JwtEitherMiddleware{
		TokenProvider: tokenProvider,
		Keys:          keys,
	}
}

func (m JwtMiddleware) UseAuthentication(f http.Handler) http.Handler {
	return authenticateWith(f, func(_ context.Context, raw string) (auth.Identity, error) {
		return userIdentity(m.TokenProvider.GetClaims(raw))
	})
}

func (m OpaqueMiddleware) UseAuthentication(f http.Handler) http.Handler {
	return authenticateWith(f, func(ctx context.Context, raw string) (auth.Identity, error) {
		return userIdentity(m.Verifier.VerifyAccess(ctx, raw))
	})
}

func (m ApiKeyMiddleware) UseAuthentication(f http.Handler) http.Handler {
	return authenticateWith(f, m.Verifier.VerifyKey)
}

func (m EitherMiddleware) UseAuthentication(f http.Handler) http.Handler {
	return authenticateWith(f, func(ctx context.Context, raw string) (auth.Identity, error) {
		if m.Keys.IsKey(raw) {
			return m.Keys.VerifyKey(ctx, raw)
		}
		return userIdentity(m.Users.VerifyAccess(ctx, raw))
	})
}

func (m JwtEitherMiddleware) UseAuthentication(f http.Handler) http.Handler {
	return authenticateWith(f, func(ctx context.Context, raw string) (auth.Identity, error) {
		if m.Keys.IsKey(raw) {
			return m.Keys.VerifyKey(ctx, raw)
		}
		return userIdentity(m.TokenProvider.GetClaims(raw))
	})
}

func authenticateWith(f http.Handler, authenticate func(ctx context.Context, raw string) (auth.Identity, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jw := httputil.NewJsonResponseWriter(w)
		raw := httputil.GetBearerToken(r)
		if raw == "" {
			jw.WriteError(errors.UnauthorizedError("Token not found"))
			return
		}

		identity, err := authenticate(r.Context(), raw)
		if err != nil {
			jw.WriteError(errors.UnauthorizedError("Invalid token"))
			return
		}

		f.ServeHTTP(w, r.WithContext(auth.WithIdentity(r.Context(), identity)))
	})
}

func userIdentity(token auth.Token, err error) (auth.Identity, error) {
	if err != nil {
		return auth.Identity{}, err
	}

	if token.IsExpired() {
		return auth.Identity{}, errors.UnauthorizedError("Invalid token")
	}

	return auth.IdentityFromToken(token), nil
}
