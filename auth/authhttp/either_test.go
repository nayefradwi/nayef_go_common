package authhttp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
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
	if err != nil {
		t.Fatal(err)
	}
	key, _, err := keys.Issue(context.Background(), testOwner, "ci", nil, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	store.reads = 0
	return keys, store, key
}

func identityOf(t *testing.T, mw func(http.Handler) http.Handler, bearer string) (auth.Identity, int) {
	t.Helper()
	var got auth.Identity
	h := mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got, _ = auth.GetIdentity(r.Context())
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+bearer)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return got, w.Code
}

func TestEitherMiddleware_RoutesByPrefix(t *testing.T) {
	keys, store, key := newKeys(t)
	userCalls := 0
	mw := NewEitherMiddleware(countingVerifier{token: unexpiredToken(), calls: &userCalls}, keys).UseAuthentication

	id, code := identityOf(t, mw, key)
	if code != http.StatusOK || id.Kind != auth.KindKey || userCalls != 0 {
		t.Fatalf("key: code=%d kind=%q userCalls=%d", code, id.Kind, userCalls)
	}

	reads := store.reads
	id, code = identityOf(t, mw, "opaque-user-token")
	if code != http.StatusOK || id.Kind != auth.KindUser || store.reads != reads {
		t.Fatalf("user: code=%d kind=%q keyReads=%d", code, id.Kind, store.reads-reads)
	}
}

func TestEitherMiddleware_RejectsExpiredUser(t *testing.T) {
	keys, _, _ := newKeys(t)
	calls := 0
	for name, tok := range map[string]auth.Token{
		"expired":     expiredToken(),
		"zero expiry": {OwnerId: testOwner},
	} {
		mw := NewEitherMiddleware(countingVerifier{token: tok, calls: &calls}, keys).UseAuthentication
		if _, code := identityOf(t, mw, "opaque"); code != http.StatusUnauthorized {
			t.Errorf("%s user token got %d", name, code)
		}
	}
}

func TestJwtEitherMiddleware(t *testing.T) {
	keys, _, key := newKeys(t)
	provider := auth.NewJwtTokenProvider(mustConfig(t))
	signed, err := provider.SignClaims(testOwner, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	mw := NewJwtEitherMiddleware(provider, keys).UseAuthentication

	if id, code := identityOf(t, mw, signed); code != http.StatusOK || id.Kind != auth.KindUser {
		t.Fatalf("jwt: code=%d kind=%q", code, id.Kind)
	}
	if id, code := identityOf(t, mw, key); code != http.StatusOK || id.Kind != auth.KindKey {
		t.Fatalf("key: code=%d kind=%q", code, id.Kind)
	}
	if _, code := identityOf(t, mw, key+"x"); code != http.StatusUnauthorized {
		t.Fatalf("unknown key got %d", code)
	}
}

func TestKindChecks(t *testing.T) {
	user := auth.WithIdentity(context.Background(), auth.Identity{Kind: auth.KindUser, Claims: map[string]any{"role": "admin"}})
	key := auth.WithIdentity(context.Background(), auth.Identity{Kind: auth.KindKey, Claims: map[string]any{"scopes": []any{"read"}}})
	deployKey := auth.WithIdentity(context.Background(), auth.Identity{Kind: auth.KindKey, Claims: map[string]any{"scopes": []any{"deploy"}}})
	other := auth.WithIdentity(context.Background(), auth.Identity{Kind: "service"})
	ok := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

	onlyKeys := Has(IsKey)(ok)
	if w := serve(onlyKeys, user); w.Code != http.StatusForbidden {
		t.Errorf("IsKey let a user through: %d", w.Code)
	}
	if w := serve(onlyKeys, key); w.Code != http.StatusOK {
		t.Errorf("IsKey blocked a key: %d", w.Code)
	}

	branch := Has(ByKind(HasClaim("role", "admin"), HasClaim("scopes", "deploy")))(ok)
	if w := serve(branch, user); w.Code != http.StatusOK {
		t.Errorf("admin user blocked: %d", w.Code)
	}
	if w := serve(branch, key); w.Code != http.StatusForbidden {
		t.Errorf("key without deploy scope passed: %d", w.Code)
	}
	if w := serve(branch, deployKey); w.Code != http.StatusOK {
		t.Errorf("key with deploy scope blocked: %d", w.Code)
	}
	if w := serve(branch, other); w.Code != http.StatusForbidden {
		t.Errorf("unknown kind passed: %d", w.Code)
	}
}
