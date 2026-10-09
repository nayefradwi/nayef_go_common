package authhttp

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/httputil"
)

type JwtMiddleware struct {
	TokenProvider auth.ITokenProvider
}

type ReferenceTokenMiddleware struct {
	ReferenceTokenProvider auth.IReferenceTokenProvider
}

func NewJwtMiddleware(tokenProvider auth.ITokenProvider) JwtMiddleware {
	return JwtMiddleware{
		TokenProvider: tokenProvider,
	}
}

func NewReferenceTokenMiddleware(referenceTokenProvider auth.IReferenceTokenProvider) ReferenceTokenMiddleware {
	return ReferenceTokenMiddleware{
		ReferenceTokenProvider: referenceTokenProvider,
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

		ctx := accessToken.WithToken(r.Context())
		r = r.WithContext(ctx)
		f.ServeHTTP(w, r)
	})
}

func (m ReferenceTokenMiddleware) UseAuthentication(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jw := httputil.NewJsonResponseWriter(w)
		tokenId := httputil.GetBearerToken(r)
		if tokenId == "" {
			jw.WriteError(errors.UnauthorizedError("Token not found"))
			return
		}

		id, err := uuid.Parse(tokenId)
		if err != nil {
			jw.WriteError(errors.UnauthorizedError("Invalid token"))
			return
		}

		accessToken, err := m.ReferenceTokenProvider.GetAccessToken(r.Context(), id)
		if err != nil || accessToken.IsExpired() {
			jw.WriteError(errors.UnauthorizedError("Invalid token"))
			return
		}

		ctx := accessToken.WithToken(r.Context())
		r = r.WithContext(ctx)
		f.ServeHTTP(w, r)
	})
}
