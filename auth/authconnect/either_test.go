package authconnect

import (
	"bytes"
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/stretchr/testify/require"
)

type fakeKeyStore struct {
	keys  []auth.ApiKey
	reads int
}

func (s *fakeKeyStore) Create(_ context.Context, k auth.ApiKey) error {
	s.keys = append(s.keys, k)
	return nil
}

func (s *fakeKeyStore) GetByHash(_ context.Context, hash []byte) (auth.ApiKey, error) {
	s.reads++
	for _, k := range s.keys {
		if bytes.Equal(k.Hash, hash) {
			return k, nil
		}
	}
	return auth.ApiKey{}, auth.ErrApiKeyNotFound
}

func (s *fakeKeyStore) ListByOwner(context.Context, uuid.UUID) ([]auth.ApiKey, error) {
	return nil, nil
}
func (s *fakeKeyStore) Touch(context.Context, uuid.UUID) error             { return nil }
func (s *fakeKeyStore) Delete(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (s *fakeKeyStore) DeleteExpired(context.Context) error                { return nil }

type countingVerifier struct {
	token auth.Token
	calls *int
}

func (v countingVerifier) VerifyAccess(context.Context, string) (auth.Token, error) {
	*v.calls++
	return v.token, nil
}

func newKeys(t *testing.T) (auth.ApiKeyManager, *fakeKeyStore, string) {
	t.Helper()
	store := &fakeKeyStore{}
	keys, err := auth.NewApiKeyManager(store, "sk_live")
	require.NoError(t, err)
	key, _, err := keys.Issue(context.Background(), testOwner, "ci", map[string]any{"scopes": []any{"read"}}, time.Time{})
	require.NoError(t, err)
	store.reads = 0
	return keys, store, key
}

func TestEitherInterceptor_RoutesByPrefix(t *testing.T) {
	keys, store, key := newKeys(t)
	userCalls := 0
	s := newServer(t, NewEitherInterceptor(countingVerifier{token: validToken(), calls: &userCalls}, keys))

	require.NoError(t, callUnary(t, s, "Bearer "+key))
	require.Equal(t, auth.KindKey, s.seen.Kind)
	require.Zero(t, userCalls)

	reads := store.reads
	require.NoError(t, callUnary(t, s, "Bearer opaque-user-token"))
	require.Equal(t, auth.KindUser, s.seen.Kind)
	require.Equal(t, reads, store.reads)

	expired := newServer(t, NewEitherInterceptor(countingVerifier{token: auth.Token{OwnerId: testOwner}, calls: &userCalls}, keys))
	requireUnauthenticated(t, callUnary(t, expired, "Bearer opaque"))
	require.Zero(t, *expired.calls)
}

func TestJwtEitherInterceptor(t *testing.T) {
	keys, _, key := newKeys(t)
	cfg, err := auth.NewJwtTokenProviderConfig(testSecret, time.Hour, auth.AccessTokenType)
	require.NoError(t, err)
	provider := auth.NewJwtTokenProvider(cfg)
	signed, err := provider.SignClaims(testOwner, map[string]any{})
	require.NoError(t, err)

	s := newServer(t, NewJwtEitherInterceptor(provider, keys))
	require.NoError(t, callUnary(t, s, "Bearer "+signed))
	require.Equal(t, auth.KindUser, s.seen.Kind)
	require.NoError(t, callUnary(t, s, "Bearer "+key))
	require.Equal(t, auth.KindKey, s.seen.Kind)
	requireUnauthenticated(t, callUnary(t, s, "Bearer "+key+"x"))
}

func TestKindChecks(t *testing.T) {
	keys, _, key := newKeys(t)
	admin := userWith(map[string]any{"role": "admin"})

	onlyKeys := newServer(t, NewJwtEitherInterceptor(admin, keys), Has(IsKey))
	requireCode(t, connect.CodePermissionDenied, callUnary(t, onlyKeys, "Bearer jwt"))
	require.NoError(t, callUnary(t, onlyKeys, "Bearer "+key))

	branch := newServer(t, NewJwtEitherInterceptor(admin, keys), Has(ByKind(HasClaim("role", "admin"), HasClaim("scopes", "read"))))
	require.NoError(t, callUnary(t, branch, "Bearer jwt"))
	require.NoError(t, callUnary(t, branch, "Bearer "+key))

	member := newServer(t, NewJwtEitherInterceptor(userWith(map[string]any{"scopes": []any{"read"}}), keys), Has(ByKind(HasClaim("role", "admin"), HasClaim("scopes", "read"))))
	requireCode(t, connect.CodePermissionDenied, callUnary(t, member, "Bearer jwt"))
}
