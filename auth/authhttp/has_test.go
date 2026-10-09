package authhttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func serve(h http.Handler, ctx context.Context) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	req.RemoteAddr = "10.0.0.1:5555"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func withUser(claims map[string]any) context.Context {
	return auth.WithIdentity(context.Background(), auth.Identity{OwnerId: testOwner, Kind: auth.KindUser, Claims: claims})
}

func TestHas(t *testing.T) {
	allow := func(context.Context, auth.Identity, *http.Request) (bool, error) { return true, nil }
	deny := func(context.Context, auth.Identity, *http.Request) (bool, error) { return false, nil }
	fail := func(context.Context, auth.Identity, *http.Request) (bool, error) { return false, errors.New("db down") }

	cases := []struct {
		name  string
		check Check
		ctx   context.Context
		want  int
	}{
		{"no identity", allow, context.Background(), http.StatusUnauthorized},
		{"denied", deny, withUser(nil), http.StatusForbidden},
		{"check error", fail, withUser(nil), http.StatusInternalServerError},
		{"allowed", allow, withUser(nil), http.StatusOK},
		{"claim in role list", HasClaim("roles", "admin"), withUser(map[string]any{"roles": []any{"member", "admin"}}), http.StatusOK},
		{"claim missing role", HasClaim("roles", "admin"), withUser(map[string]any{"roles": []any{"member"}}), http.StatusForbidden},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			called := false
			w := serve(Has(c.check)(nextHandler(t, &called)), c.ctx)
			assert.Equal(t, c.want, w.Code)
			assert.Equal(t, c.want == http.StatusOK, called)
		})
	}

	t.Run("check error is not leaked", func(t *testing.T) {
		w := serve(Has(fail)(nextHandler(t, new(bool))), withUser(nil))
		assert.NotContains(t, w.Body.String(), "db down")
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
	h := RateLimit(limiter, ByIP("otp"))(nextHandler(t, new(bool)))

	assert.Equal(t, http.StatusOK, serve(h, context.Background()).Code)
	assert.Equal(t, http.StatusOK, serve(h, context.Background()).Code)

	w := serve(h, context.Background())
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, "90", w.Header().Get("Retry-After"))
}

func TestRateLimit_StoreErrorFailsClosed(t *testing.T) {
	limiter, err := auth.NewLimiter(&memoryAttemptStore{err: errors.New("pg: secret detail")}, 2, time.Minute)
	require.NoError(t, err)
	called := false
	w := serve(RateLimit(limiter, ByIP("otp"))(nextHandler(t, &called)), context.Background())
	assert.False(t, called)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NotContains(t, w.Body.String(), "secret detail")
}

func TestByIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:5555"
	k, err := ByIP("otp")(req)
	require.NoError(t, err)
	assert.Equal(t, "otp:10.0.0.1", k)

	req.RemoteAddr = "[::1]:80"
	k, _ = ByIP("otp")(req)
	assert.Equal(t, "otp:::1", k)
}
