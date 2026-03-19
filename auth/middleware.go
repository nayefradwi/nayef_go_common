package auth

import (
	"net/http"

	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/httputil"
)

type JwtAuthenticationMiddleware struct {
	TokenProvider ITokenProvider
}

type JwtReferenceTokenAuthenticationMiddleware struct {
	ReferenceTokenProvider IReferenceTokenProvider
}

func NewJwtAuthenticationMiddleware(tokenProvider ITokenProvider) JwtAuthenticationMiddleware {
	return JwtAuthenticationMiddleware{
		TokenProvider: tokenProvider,
	}
}

func NewJwtReferenceTokenAuthenticationMiddleware(referenceTokenProvider IReferenceTokenProvider) JwtReferenceTokenAuthenticationMiddleware {
	return JwtReferenceTokenAuthenticationMiddleware{
		ReferenceTokenProvider: referenceTokenProvider,
	}
}

func (m JwtAuthenticationMiddleware) UseAuthentication(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jw := httputil.NewJsonResponseWriter(w)
		token := httputil.GetBearerToken(r)
		if token == "" {
			jw.WriteError(errors.UnauthorizedError("Token not found"))
			return
		}

		accessToken, err := m.TokenProvider.GetClaims(token)
		if err != nil {
			jw.WriteError(errors.UnauthorizedError("Invalid token"))
			return
		}

		ctx := accessToken.WithToken(r.Context())
		r = r.WithContext(ctx)
		f.ServeHTTP(w, r)
	})
}

func (m JwtReferenceTokenAuthenticationMiddleware) UseAuthentication(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jw := httputil.NewJsonResponseWriter(w)
		tokenId := httputil.GetBearerToken(r)
		if tokenId == "" {
			jw.WriteError(errors.UnauthorizedError("Token not found"))
			return
		}

		accessToken, err := m.ReferenceTokenProvider.GetAccessToken(tokenId)
		if err != nil {
			jw.WriteError(errors.UnauthorizedError("Invalid token"))
			return
		}

		ctx := accessToken.WithToken(r.Context())
		r = r.WithContext(ctx)
		f.ServeHTTP(w, r)
	})
}
