package rest

import (
	"net/http"

	"github.com/nayefradwi/nayef_go_common/core"
	"github.com/nayefradwi/nayef_go_common/modules/auth"
	"github.com/nayefradwi/nayef_go_common/modules/rest"
)

type JwtAuthenticationMiddleware struct {
	TokenProvider auth.ITokenProvider
}

type JwtReferenceTokenAuthenicationMiddleware struct {
	ReferenceTokenProvider auth.IReferenceTokenProvider
}

func NewJwtAuthenticationMiddleware(tokenProvider auth.ITokenProvider) JwtAuthenticationMiddleware {
	return JwtAuthenticationMiddleware{
		TokenProvider: tokenProvider,
	}
}

func NewJwtReferenceTokenAuthenicationMiddleware(referenceTokenProvider auth.IReferenceTokenProvider) JwtReferenceTokenAuthenicationMiddleware {
	return JwtReferenceTokenAuthenicationMiddleware{
		ReferenceTokenProvider: referenceTokenProvider,
	}
}

func (m JwtAuthenticationMiddleware) UseAuthenitcation(f http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jw := rest.NewJsonResponseWriter(w)
		token := rest.GetBearerToken(r)
		if token == "" {
			jw.WriteError(core.UnauthorizedError("Token not found"))
			return
		}

		accessToken, err := m.TokenProvider.GetClaims(token)
		if err != nil {
			jw.WriteError(core.UnauthorizedError("Invalid token"))
			return
		}

		ctx := accessToken.WithToken(r.Context())
		r = r.WithContext(ctx)
		f.ServeHTTP(w, r)
	})

	return handler
}

func (m JwtReferenceTokenAuthenicationMiddleware) UseAuthenitcation(f http.Handler) http.Handler {
	hanlder := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jw := rest.NewJsonResponseWriter(w)
		tokenId := rest.GetBearerToken(r)
		if tokenId == "" {
			jw.WriteError(core.UnauthorizedError("Token not found"))
			return
		}

		accessToken, err := m.ReferenceTokenProvider.GetAccessToken(tokenId)
		if err != nil {
			jw.WriteError(core.UnauthorizedError("Invalid token"))
			return
		}

		ctx := accessToken.WithToken(r.Context())
		r = r.WithContext(ctx)

		f.ServeHTTP(w, r)
	})

	return hanlder
}
