package authconnect

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
)

type authenticateFunc func(ctx context.Context, raw string) (auth.Token, error)

type interceptor struct {
	authenticate authenticateFunc
}

func NewJwtInterceptor(tokenProvider auth.ITokenProvider) connect.Interceptor {
	return interceptor{authenticate: func(_ context.Context, raw string) (auth.Token, error) {
		return tokenProvider.GetClaims(raw)
	}}
}

func NewReferenceTokenInterceptor(referenceTokenProvider auth.IReferenceTokenProvider) connect.Interceptor {
	return interceptor{authenticate: func(ctx context.Context, raw string) (auth.Token, error) {
		id, err := uuid.Parse(raw)
		if err != nil {
			return auth.Token{}, err
		}
		return referenceTokenProvider.GetAccessToken(ctx, id)
	}}
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

	token, err := i.authenticate(ctx, raw)
	if err != nil || token.IsExpired() {
		return ctx, unauthenticated("Invalid token")
	}

	return auth.WithIdentity(ctx, auth.IdentityFromToken(token)), nil
}

func unauthenticated(msg string) error {
	return connect.NewError(connect.CodeUnauthenticated, errors.New(msg))
}
