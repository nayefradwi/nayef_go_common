package auth

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type memSessionStore struct {
	mu   sync.Mutex
	rows []Session
}

func (m *memSessionStore) Create(_ context.Context, s Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows = append(m.rows, s)
	return nil
}

func (m *memSessionStore) Rotate(_ context.Context, hash []byte, next Session) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, r := range m.rows {
		if !bytes.Equal(r.RefreshHash, hash) || !r.RotatedAt.IsZero() || !time.Now().Before(r.ExpiresAt) {
			continue
		}
		m.rows[i].RotatedAt = time.Now()
		next.FamilyId, next.OwnerId, next.Claims = r.FamilyId, r.OwnerId, r.Claims
		m.rows = append(m.rows, next)
		return next, nil
	}
	return Session{}, ErrSessionNotFound
}

func (m *memSessionStore) find(match func(Session) bool) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.rows {
		if match(r) {
			return r, nil
		}
	}
	return Session{}, ErrSessionNotFound
}

func (m *memSessionStore) GetByRefresh(_ context.Context, hash []byte) (Session, error) {
	return m.find(func(r Session) bool { return bytes.Equal(r.RefreshHash, hash) })
}

func (m *memSessionStore) GetByAccess(_ context.Context, hash []byte) (Session, error) {
	return m.find(func(r Session) bool { return r.AccessHash != nil && bytes.Equal(r.AccessHash, hash) })
}

func (m *memSessionStore) delete(match func(Session) bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	kept := m.rows[:0]
	for _, r := range m.rows {
		if !match(r) {
			kept = append(kept, r)
		}
	}
	m.rows = kept
	return nil
}

func (m *memSessionStore) DeleteFamily(_ context.Context, id uuid.UUID) error {
	return m.delete(func(r Session) bool { return r.FamilyId == id })
}

func (m *memSessionStore) DeleteOwner(_ context.Context, id uuid.UUID) error {
	return m.delete(func(r Session) bool { return r.OwnerId == id })
}

func (m *memSessionStore) DeleteExpired(context.Context) error {
	return m.delete(func(r Session) bool { return !time.Now().Before(r.ExpiresAt) })
}

func mustOpaqueSessionManager(t *testing.T, store SessionStore, accessTTL time.Duration) SessionManager {
	t.Helper()
	s, err := NewOpaqueSessionManager(store, accessTTL, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSessions_RefreshReuseRevokesFamily(t *testing.T) {
	ctx := context.Background()
	s := mustOpaqueSessionManager(t, &memSessionStore{}, time.Minute)

	first, err := s.Issue(ctx, testOwner, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Refresh(ctx, first.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.Refresh(ctx, first.RefreshToken); err == nil {
		t.Fatal("reused refresh token should be rejected")
	}
	if _, err := s.Refresh(ctx, second.RefreshToken); err == nil {
		t.Fatal("reuse should revoke the rotated child too")
	}
	if _, err := s.VerifyAccess(ctx, second.AccessToken); err == nil {
		t.Fatal("reuse should revoke opaque access too")
	}
}

func TestSessions_ExpiredRefreshDoesNotRevokeFamily(t *testing.T) {
	ctx := context.Background()
	store := &memSessionStore{}
	s := mustOpaqueSessionManager(t, store, time.Minute)

	dto, err := s.Issue(ctx, testOwner, nil)
	if err != nil {
		t.Fatal(err)
	}
	store.rows[0].ExpiresAt = time.Now().Add(-time.Second)

	if _, err := s.Refresh(ctx, dto.RefreshToken); err == nil {
		t.Fatal("expired refresh token should be rejected")
	}
	if len(store.rows) != 1 {
		t.Fatalf("expired refresh should not delete the family, rows=%d", len(store.rows))
	}
}

func TestSessions_RevokeOnlyHitsTarget(t *testing.T) {
	ctx := context.Background()
	other := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	s := mustOpaqueSessionManager(t, &memSessionStore{}, time.Minute)

	a1, _ := s.Issue(ctx, testOwner, nil)
	a2, _ := s.Issue(ctx, testOwner, nil)
	b, _ := s.Issue(ctx, other, nil)

	if err := s.Revoke(ctx, a1.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := s.VerifyAccess(ctx, a1.AccessToken); err == nil {
		t.Fatal("revoked session access should fail")
	}
	if _, err := s.VerifyAccess(ctx, a2.AccessToken); err != nil {
		t.Fatal("other session of same owner should survive logout", err)
	}

	if err := s.RevokeOwner(ctx, testOwner); err != nil {
		t.Fatal(err)
	}
	if _, err := s.VerifyAccess(ctx, a2.AccessToken); err == nil {
		t.Fatal("logout everywhere should kill owner sessions")
	}
	if _, err := s.VerifyAccess(ctx, b.AccessToken); err != nil {
		t.Fatal("other owner should survive", err)
	}
}

func TestSessions_OldAccessLivesUntilExpiry(t *testing.T) {
	ctx := context.Background()
	store := &memSessionStore{}
	s := mustOpaqueSessionManager(t, store, time.Minute)

	first, _ := s.Issue(ctx, testOwner, nil)
	if _, err := s.Refresh(ctx, first.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := s.VerifyAccess(ctx, first.AccessToken); err != nil {
		t.Fatal("in-flight access token should survive a refresh", err)
	}

	store.rows[0].AccessExpiresAt = time.Now().Add(-time.Second)
	if _, err := s.VerifyAccess(ctx, first.AccessToken); err == nil {
		t.Fatal("expired access token should fail")
	}
}

func TestSessions_JwtRefreshKeepsClaims(t *testing.T) {
	ctx := context.Background()
	jwt := NewJwtTokenProvider(mustConfig(t))
	s, err := NewJwtSessionManager(&memSessionStore{}, jwt, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	first, err := s.Issue(ctx, testOwner, map[string]any{"role": "admin"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Refresh(ctx, first.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}

	tok, err := jwt.GetClaims(second.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if tok.OwnerId != testOwner || tok.Claims["role"] != "admin" {
		t.Fatalf("refreshed jwt lost owner or claims: %+v", tok)
	}
}

func TestNewSessions_Rejects(t *testing.T) {
	if _, err := NewOpaqueSessionManager(nil, time.Minute, time.Hour); err == nil {
		t.Fatal("nil store should be rejected")
	}
	if _, err := NewOpaqueSessionManager(&memSessionStore{}, 0, time.Hour); err == nil {
		t.Fatal("zero access ttl should be rejected")
	}
	refreshCfg, _ := NewJwtTokenProviderConfig(testSecret, time.Hour, RefreshTokenType)
	if _, err := NewJwtSessionManager(&memSessionStore{}, NewJwtTokenProvider(refreshCfg), time.Hour); err == nil {
		t.Fatal("refresh-type jwt provider should be rejected")
	}
}
