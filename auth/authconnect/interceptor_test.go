package authconnect

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	unaryProcedure  = "/test.v1.TestService/Call"
	streamProcedure = "/test.v1.TestService/Stream"
	testSecret      = "test-secret-key-at-least-32-bytes"
)

var testOwner = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type stubTokenProvider struct {
	token auth.Token
	err   error
}

func (s stubTokenProvider) GetClaims(string) (auth.Token, error) { return s.token, s.err }
func (s stubTokenProvider) SignClaims(uuid.UUID, map[string]any) (string, error) {
	return "", nil
}

type stubVerifier struct {
	token auth.Token
	err   error
}

func (s stubVerifier) VerifyAccess(context.Context, string) (auth.Token, error) {
	return s.token, s.err
}

func validToken() auth.Token {
	return auth.Token{OwnerId: testOwner, ExpiresAt: time.Now().UTC().Add(time.Hour)}
}

type server struct {
	url   string
	seen  *auth.Identity
	calls *int
}

func newServer(t *testing.T, i ...connect.Interceptor) server {
	t.Helper()
	s := server{seen: &auth.Identity{}, calls: new(int)}

	record := func(ctx context.Context) {
		*s.calls++
		*s.seen, _ = auth.GetIdentity(ctx)
	}

	mux := http.NewServeMux()
	mux.Handle(unaryProcedure, connect.NewUnaryHandler(unaryProcedure,
		func(ctx context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
			record(ctx)
			return connect.NewResponse(&emptypb.Empty{}), nil
		},
		connect.WithInterceptors(i...),
	))
	mux.Handle(streamProcedure, connect.NewServerStreamHandler(streamProcedure,
		func(ctx context.Context, _ *connect.Request[emptypb.Empty], st *connect.ServerStream[emptypb.Empty]) error {
			record(ctx)
			return st.Send(&emptypb.Empty{})
		},
		connect.WithInterceptors(i...),
	))

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	s.url = ts.URL
	return s
}

func callUnary(t *testing.T, s server, authorization string) error {
	t.Helper()
	client := connect.NewClient[emptypb.Empty, emptypb.Empty](http.DefaultClient, s.url+unaryProcedure)
	req := connect.NewRequest(&emptypb.Empty{})
	if authorization != "" {
		req.Header().Set("Authorization", authorization)
	}
	_, err := client.CallUnary(context.Background(), req)
	return err
}

func requireUnauthenticated(t *testing.T, err error) {
	t.Helper()
	var ce *connect.Error
	require.True(t, errors.As(err, &ce), "got %T: %v", err, err)
	require.Equal(t, connect.CodeUnauthenticated, ce.Code())
}

func TestJwtInterceptor_ValidTokenReachesHandler(t *testing.T) {
	cfg, err := auth.NewJwtTokenProviderConfig(testSecret, time.Hour, auth.AccessTokenType)
	require.NoError(t, err)
	provider := auth.NewJwtTokenProvider(cfg)
	signed, err := provider.SignClaims(testOwner, map[string]any{})
	require.NoError(t, err)

	s := newServer(t, NewJwtInterceptor(provider))
	require.NoError(t, callUnary(t, s, "Bearer "+signed))
	require.Equal(t, testOwner, s.seen.OwnerId)
}

func TestJwtInterceptor_Rejects(t *testing.T) {
	cases := []struct {
		name          string
		provider      stubTokenProvider
		authorization string
	}{
		{"missing header", stubTokenProvider{token: validToken()}, ""},
		{"wrong scheme", stubTokenProvider{token: validToken()}, "Basic abc"},
		{"empty bearer", stubTokenProvider{token: validToken()}, "Bearer "},
		{"provider error", stubTokenProvider{err: errors.New("bad")}, "Bearer x"},
		{"expired", stubTokenProvider{token: auth.Token{OwnerId: testOwner, ExpiresAt: time.Now().Add(-time.Hour)}}, "Bearer x"},
		{"zero expiry", stubTokenProvider{token: auth.Token{OwnerId: testOwner}}, "Bearer x"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newServer(t, NewJwtInterceptor(c.provider))
			requireUnauthenticated(t, callUnary(t, s, c.authorization))
			require.Zero(t, *s.calls)
		})
	}
}

func TestOpaqueInterceptor(t *testing.T) {
	denied := newServer(t, NewOpaqueInterceptor(stubVerifier{err: errors.New("not found")}))
	requireUnauthenticated(t, callUnary(t, denied, "Bearer opaque"))
	require.Zero(t, *denied.calls)

	s := newServer(t, NewOpaqueInterceptor(stubVerifier{token: validToken()}))
	require.NoError(t, callUnary(t, s, "Bearer opaque"))
	require.Equal(t, testOwner, s.seen.OwnerId)
}

func TestInterceptor_StreamingHandlerRequiresToken(t *testing.T) {
	s := newServer(t, NewJwtInterceptor(stubTokenProvider{token: validToken()}))
	client := connect.NewClient[emptypb.Empty, emptypb.Empty](http.DefaultClient, s.url+streamProcedure)

	stream, err := client.CallServerStream(context.Background(), connect.NewRequest(&emptypb.Empty{}))
	require.NoError(t, err)
	for stream.Receive() {
	}
	requireUnauthenticated(t, stream.Err())
	require.Zero(t, *s.calls)

	req := connect.NewRequest(&emptypb.Empty{})
	req.Header().Set("Authorization", "Bearer x")
	stream, err = client.CallServerStream(context.Background(), req)
	require.NoError(t, err)
	for stream.Receive() {
	}
	require.NoError(t, stream.Err())
	require.Equal(t, 1, *s.calls)
}

type stubKeyVerifier struct {
	identity auth.Identity
	err      error
}

func (s stubKeyVerifier) VerifyKey(context.Context, string) (auth.Identity, error) {
	return s.identity, s.err
}

func TestApiKeyInterceptor(t *testing.T) {
	denied := newServer(t, NewApiKeyInterceptor(stubKeyVerifier{err: errors.New("not found")}))
	requireUnauthenticated(t, callUnary(t, denied, "Bearer sk_live_x"))
	require.Zero(t, *denied.calls)

	s := newServer(t, NewApiKeyInterceptor(stubKeyVerifier{identity: auth.Identity{OwnerId: testOwner, Kind: auth.KindKey}}))
	require.NoError(t, callUnary(t, s, "Bearer sk_live_x"))
	require.Equal(t, testOwner, s.seen.OwnerId)
	require.Equal(t, auth.KindKey, s.seen.Kind)
}
