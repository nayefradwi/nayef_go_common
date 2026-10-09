package authconnect

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func userWith(claims map[string]any) stubTokenProvider {
	tok := validToken()
	tok.Claims = claims
	return stubTokenProvider{token: tok}
}

func requireCode(t *testing.T, want connect.Code, err error) *connect.Error {
	t.Helper()
	var ce *connect.Error
	require.True(t, errors.As(err, &ce), "got %T: %v", err, err)
	require.Equal(t, want, ce.Code())
	return ce
}

func TestHas(t *testing.T) {
	allow := func(context.Context, auth.Identity, connect.AnyRequest) (bool, error) { return true, nil }
	deny := func(context.Context, auth.Identity, connect.AnyRequest) (bool, error) { return false, nil }
	fail := func(context.Context, auth.Identity, connect.AnyRequest) (bool, error) {
		return false, errors.New("db down")
	}

	t.Run("no identity", func(t *testing.T) {
		s := newServer(t, Has(allow))
		requireCode(t, connect.CodeUnauthenticated, callUnary(t, s, ""))
		require.Zero(t, *s.calls)
	})

	cases := []struct {
		name  string
		check Check
		roles []any
		want  connect.Code
	}{
		{"denied", deny, nil, connect.CodePermissionDenied},
		{"check error", fail, nil, connect.CodeInternal},
		{"claim missing role", HasClaim("roles", "admin"), []any{"member"}, connect.CodePermissionDenied},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newServer(t, NewJwtInterceptor(userWith(map[string]any{"roles": c.roles})), Has(c.check))
			ce := requireCode(t, c.want, callUnary(t, s, "Bearer x"))
			assert.NotContains(t, ce.Message(), "db down")
			require.Zero(t, *s.calls)
		})
	}

	t.Run("allowed by role", func(t *testing.T) {
		s := newServer(t, NewJwtInterceptor(userWith(map[string]any{"roles": []any{"admin"}})), Has(HasClaim("roles", "admin")))
		require.NoError(t, callUnary(t, s, "Bearer x"))
		require.Equal(t, 1, *s.calls)
	})

	t.Run("streaming is rejected", func(t *testing.T) {
		s := newServer(t, NewJwtInterceptor(userWith(nil)), Has(allow))
		client := connect.NewClient[emptypb.Empty, emptypb.Empty](http.DefaultClient, s.url+streamProcedure)
		req := connect.NewRequest(&emptypb.Empty{})
		req.Header().Set("Authorization", "Bearer x")
		stream, err := client.CallServerStream(context.Background(), req)
		require.NoError(t, err)
		for stream.Receive() {
		}
		requireCode(t, connect.CodePermissionDenied, stream.Err())
		require.Zero(t, *s.calls)
	})
}

type memoryAttemptStore struct {
	counts map[string]int
	err    error
}

func (s *memoryAttemptStore) Hit(_ context.Context, key string, window time.Duration) (auth.Attempt, error) {
	if s.err != nil {
		return auth.Attempt{}, s.err
	}
	s.counts[key]++
	return auth.Attempt{Count: s.counts[key], ResetAt: time.Now().Add(window)}, nil
}
func (s *memoryAttemptStore) Reset(context.Context, string) error { return nil }
func (s *memoryAttemptStore) DeleteExpired(context.Context) error { return nil }

func TestRateLimit(t *testing.T) {
	limiter, err := auth.NewLimiter(&memoryAttemptStore{counts: map[string]int{}}, 2, 90*time.Second)
	require.NoError(t, err)
	s := newServer(t, RateLimit(limiter, ByIP("otp")))

	require.NoError(t, callUnary(t, s, ""))
	require.NoError(t, callUnary(t, s, ""))
	ce := requireCode(t, connect.CodeResourceExhausted, callUnary(t, s, ""))
	assert.Equal(t, "90", ce.Meta().Get("Retry-After"))
	require.Equal(t, 2, *s.calls)
}

func TestRateLimit_StoreErrorFailsClosed(t *testing.T) {
	limiter, err := auth.NewLimiter(&memoryAttemptStore{err: errors.New("pg: secret detail")}, 2, time.Minute)
	require.NoError(t, err)
	s := newServer(t, RateLimit(limiter, ByIP("otp")))

	ce := requireCode(t, connect.CodeInternal, callUnary(t, s, ""))
	assert.NotContains(t, ce.Message(), "secret detail")
	require.Zero(t, *s.calls)
}
