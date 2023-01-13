package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/nayefradwi/nayef_go_common/baseError"
	"github.com/nayefradwi/nayef_go_common/response"
)

type ClaimsKey struct{}

var secret string = ""

func SetSecret(envSecret string) {
	secret = envSecret
}

func AuthorizeHeaderMiddleware(f http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		tokenSplit := strings.Split(authHeader, "Bearer ")
		if len(tokenSplit) < 2 {
			response.WriteErrorResponse(w, baseError.NewUnAuthorizedError())
			return
		}
		token := tokenSplit[1]
		claims, err := DecodeAccessToken(token, secret)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), ClaimsKey{}, claims)
		rcopy := r.WithContext(ctx)
		f.ServeHTTP(w, rcopy)
	})
	return handler
}
