package auth

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type memApiKeyStore struct {
	mu            sync.Mutex
	rows          []ApiKey
	reads, touchs int
}

func (m *memApiKeyStore) Create(_ context.Context, k ApiKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows = append(m.rows, k)
	return nil
}

func (m *memApiKeyStore) GetByHash(_ context.Context, hash []byte) (ApiKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reads++
	for _, k := range m.rows {
		if bytes.Equal(k.Hash, hash) {
			return k, nil
		}
	}
	return ApiKey{}, ErrApiKeyNotFound
}

func (m *memApiKeyStore) ListByOwner(_ context.Context, owner uuid.UUID) ([]ApiKey, error) {
	return nil, nil
}

func (m *memApiKeyStore) Touch(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.touchs++
	for i := range m.rows {
		if m.rows[i].Id == id {
			m.rows[i].LastUsedAt = time.Now()
		}
	}
	return nil
}

func (m *memApiKeyStore) Delete(_ context.Context, owner, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, k := range m.rows {
		if k.Id == id && k.OwnerId == owner {
			m.rows = append(m.rows[:i], m.rows[i+1:]...)
			return nil
		}
	}
	return ErrApiKeyNotFound
}

func (m *memApiKeyStore) DeleteExpired(context.Context) error { return nil }

func mustApiKeyManager(t *testing.T, store ApiKeyStore) ApiKeyManager {
	t.Helper()
	m, err := NewApiKeyManager(store, "sk_live")
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestApiKey_VerifyRejects(t *testing.T) {
	ctx := context.Background()
	store := &memApiKeyStore{}
	m := mustApiKeyManager(t, store)

	key, _, err := m.Issue(ctx, testOwner, "ci", nil, time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := m.VerifyKey(ctx, "sk_test_"+key[len("sk_live_"):]); err == nil {
		t.Fatal("wrong prefix accepted")
	}
	if store.reads != 0 {
		t.Fatal("wrong prefix should not reach the store")
	}
	if _, err := m.VerifyKey(ctx, key+"x"); err == nil {
		t.Fatal("unknown key accepted")
	}

	expired, _, _ := m.Issue(ctx, testOwner, "old", nil, time.Now().Add(-time.Second))
	if _, err := m.VerifyKey(ctx, expired); err == nil {
		t.Fatal("expired key accepted")
	}
}

func TestApiKey_IdentityAndTouch(t *testing.T) {
	ctx := context.Background()
	store := &memApiKeyStore{}
	m := mustApiKeyManager(t, store)

	key, _, _ := m.Issue(ctx, testOwner, "ci", map[string]any{"scopes": []string{"deploy"}}, time.Time{})
	id, err := m.VerifyKey(ctx, key)
	if err != nil {
		t.Fatal("key without expiry rejected", err)
	}
	if id.Kind != KindKey || id.OwnerId != testOwner || !id.HasClaim("scopes", "deploy") {
		t.Fatalf("bad identity: %+v", id)
	}
	if _, err := m.VerifyKey(ctx, key); err != nil {
		t.Fatal(err)
	}
	if store.touchs != 1 {
		t.Fatalf("touch should run once per window, ran %d", store.touchs)
	}

	store.rows[0].LastUsedAt = time.Now().Add(-2 * apiKeyTouchEvery)
	if _, err := m.VerifyKey(ctx, key); err != nil {
		t.Fatal(err)
	}
	if store.touchs != 2 {
		t.Fatalf("stale last_used_at should be touched, ran %d", store.touchs)
	}
}

func TestApiKey_RevokeIsOwnerScoped(t *testing.T) {
	ctx := context.Background()
	m := mustApiKeyManager(t, &memApiKeyStore{})
	other := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	key, k, _ := m.Issue(ctx, testOwner, "ci", nil, time.Time{})
	if err := m.Revoke(ctx, other, k.Id); err == nil {
		t.Fatal("another owner revoked the key")
	}
	if _, err := m.VerifyKey(ctx, key); err != nil {
		t.Fatal("key should survive a foreign revoke", err)
	}
	if err := m.Revoke(ctx, testOwner, k.Id); err != nil {
		t.Fatal(err)
	}
	if _, err := m.VerifyKey(ctx, key); err == nil {
		t.Fatal("revoked key accepted")
	}
}

func TestNewApiKeyManager_Rejects(t *testing.T) {
	if _, err := NewApiKeyManager(nil, "sk_live"); err == nil {
		t.Error("nil store should be rejected")
	}
	for _, bad := range []string{"", "SK", "sk-live", "1sk", "a234567890123456x"} {
		if _, err := NewApiKeyManager(&memApiKeyStore{}, bad); err == nil {
			t.Error(bad + " should be rejected")
		}
	}
	if _, err := NewApiKeyManager(&memApiKeyStore{}, "a234567890123456"); err != nil {
		t.Error("16-byte prefix should pass", err)
	}
}
