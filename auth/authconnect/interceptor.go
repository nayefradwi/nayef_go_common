package authconnect

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/nayefradwi/nayef_go_common/auth"
)

type authenticateFunc func(ctx context.Context, raw string) (auth.Identity, error)

type interceptor struct {
	authenticate authenticateFunc
}

func NewJwtInterceptor(tokenProvider auth.ITokenProvider) connect.Interceptor {
	return interceptor{authenticate: func(_ context.Context, raw string) (auth.Identity, error) {
		return userIdentity(tokenProvider.GetClaims(raw))
	}}
}

func NewOpaqueInterceptor(verifier auth.AccessVerifier) connect.Interceptor {
	return interceptor{authenticate: func(ctx context.Context, raw string) (auth.Identity, error) {
		return userIdentity(verifier.VerifyAccess(ctx, raw))
	}}
}

func NewApiKeyInterceptor(verifier auth.KeyVerifier) connect.Interceptor {
	return interceptor{authenticate: verifier.VerifyKey}
}

func NewEitherInterceptor(users auth.AccessVerifier, keys auth.ApiKeyManager) connect.Interceptor {
	return interceptor{authenticate: func(ctx context.Context, raw string) (auth.Identity, error) {
		if keys.IsKey(raw) {
			return keys.VerifyKey(ctx, raw)
		}
		return userIdentity(users.VerifyAccess(ctx, raw))
	}}
}

func NewJwtEitherInterceptor(tokenProvider auth.ITokenProvider, keys auth.ApiKeyManager) connect.Interceptor {
	return interceptor{authenticate: func(ctx context.Context, raw string) (auth.Identity, error) {
		if keys.IsKey(raw) {
			return keys.VerifyKey(ctx, raw)
		}
		return userIdentity(tokenProvider.GetClaims(raw))
	}}
}

func userIdentity(token auth.Token, err error) (auth.Identity, error) {
	if err != nil {
		return auth.Identity{}, err
	}

	if token.IsExpired() {
		return auth.Identity{}, errors.New("token expired")
	}

	return auth.IdentityFromToken(token), nil
}

func (i interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			return next(ctx, req)
		}

		ctx, err := i.withIdentity(ctx, req.Header())
		if err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

func (i interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, err := i.withIdentity(ctx, conn.RequestHeader())
		if err != nil {
			return err
		}
		return next(ctx, conn)
	}
}

func (i interceptor) withIdentity(ctx context.Context, header http.Header) (context.Context, error) {
	raw, ok := strings.CutPrefix(header.Get("Authorization"), "Bearer ")
	if !ok || raw == "" {
		return ctx, unauthenticated("Token not found")
	}

	identity, err := i.authenticate(ctx, raw)
	if err != nil {
		return ctx, unauthenticated("Invalid token")
	}

	return auth.WithIdentity(ctx, identity), nil
}

func unauthenticated(msg string) error {
	return connect.NewError(connect.CodeUnauthenticated, errors.New(msg))
}
