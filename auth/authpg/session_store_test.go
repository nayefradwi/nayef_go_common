package authpg

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSession(t *testing.T, owner uuid.UUID, expiresAt time.Time) auth.Session {
	t.Helper()
	_, refreshHash, err := auth.NewOpaqueToken()
	require.NoError(t, err)
	_, accessHash, err := auth.NewOpaqueToken()
	require.NoError(t, err)
	return auth.Session{
		FamilyId:        uuid.New(),
		OwnerId:         owner,
		RefreshHash:     refreshHash,
		AccessHash:      accessHash,
		Claims:          map[string]any{"role": "admin"},
		ExpiresAt:       expiresAt,
		AccessExpiresAt: time.Now().Add(time.Minute),
	}
}

func nextSession(t *testing.T) auth.Session {
	t.Helper()
	_, refreshHash, err := auth.NewOpaqueToken()
	require.NoError(t, err)
	return auth.Session{RefreshHash: refreshHash, ExpiresAt: time.Now().Add(time.Hour)}
}

func TestSessionStore(t *testing.T) {
	pool := mustCreatePostgresConn(t)
	store, err := NewSessionStore(pool)
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("concurrent rotate of one refresh succeeds once", func(t *testing.T) {
		s := newTestSession(t, testOwner, time.Now().Add(time.Hour))
		require.NoError(t, store.Create(ctx, s))

		const n = 20
		var ok, notFound atomic.Int32
		var wg sync.WaitGroup
		for range n {
			wg.Go(func() {
				rotated, err := store.Rotate(ctx, s.RefreshHash, nextSession(t))
				switch {
				case err == nil:
					ok.Add(1)
					assert.Equal(t, s.FamilyId, rotated.FamilyId)
					assert.Equal(t, "admin", rotated.Claims["role"])
				case assert.ErrorIs(t, err, auth.ErrSessionNotFound):
					notFound.Add(1)
				}
			})
		}
		wg.Wait()
		assert.EqualValues(t, 1, ok.Load())
		assert.EqualValues(t, n-1, notFound.Load())
	})

	t.Run("expired refresh does not rotate", func(t *testing.T) {
		s := newTestSession(t, testOwner, time.Now().Add(-time.Second))
		require.NoError(t, store.Create(ctx, s))

		_, err := store.Rotate(ctx, s.RefreshHash, nextSession(t))
		require.ErrorIs(t, err, auth.ErrSessionNotFound)

		got, err := store.GetByRefresh(ctx, s.RefreshHash)
		require.NoError(t, err)
		assert.True(t, got.RotatedAt.IsZero())
	})

	t.Run("unknown hash is not found, not internal", func(t *testing.T) {
		_, err := store.GetByRefresh(ctx, auth.HashOpaqueToken("missing"))
		require.ErrorIs(t, err, auth.ErrSessionNotFound)
		_, err = store.GetByAccess(ctx, auth.HashOpaqueToken("missing"))
		require.ErrorIs(t, err, auth.ErrSessionNotFound)
	})

	t.Run("jwt mode row without access round-trips", func(t *testing.T) {
		s := newTestSession(t, testOwner, time.Now().Add(time.Hour))
		s.AccessHash, s.AccessExpiresAt, s.Claims = nil, time.Time{}, nil
		require.NoError(t, store.Create(ctx, s))

		got, err := store.GetByRefresh(ctx, s.RefreshHash)
		require.NoError(t, err)
		assert.Nil(t, got.AccessHash)
		assert.True(t, got.AccessExpiresAt.IsZero())
	})

	t.Run("delete family leaves other families", func(t *testing.T) {
		a := newTestSession(t, testOwner, time.Now().Add(time.Hour))
		b := newTestSession(t, testOwner, time.Now().Add(time.Hour))
		require.NoError(t, store.Create(ctx, a))
		require.NoError(t, store.Create(ctx, b))

		require.NoError(t, store.DeleteFamily(ctx, a.FamilyId))
		_, err := store.GetByRefresh(ctx, a.RefreshHash)
		require.ErrorIs(t, err, auth.ErrSessionNotFound)
		_, err = store.GetByRefresh(ctx, b.RefreshHash)
		require.NoError(t, err)
	})

	t.Run("delete expired keeps live rows", func(t *testing.T) {
		live := newTestSession(t, testOwner, time.Now().Add(time.Hour))
		dead := newTestSession(t, testOwner, time.Now().Add(-time.Second))
		require.NoError(t, store.Create(ctx, live))
		require.NoError(t, store.Create(ctx, dead))

		require.NoError(t, store.DeleteExpired(ctx))
		_, err := store.GetByRefresh(ctx, dead.RefreshHash)
		require.ErrorIs(t, err, auth.ErrSessionNotFound)
		_, err = store.GetByRefresh(ctx, live.RefreshHash)
		require.NoError(t, err)
	})

	t.Run("manager reuse kills the family", func(t *testing.T) {
		m, err := auth.NewOpaqueSessionManager(store, time.Minute, time.Hour)
		require.NoError(t, err)

		first, err := m.Issue(ctx, testOwner, nil)
		require.NoError(t, err)
		second, err := m.Refresh(ctx, first.RefreshToken)
		require.NoError(t, err)
		_, err = m.VerifyAccess(ctx, second.AccessToken)
		require.NoError(t, err)

		_, err = m.Refresh(ctx, first.RefreshToken)
		require.Error(t, err)
		_, err = m.VerifyAccess(ctx, second.AccessToken)
		require.Error(t, err)
	})
}

func TestSessionStore_CustomTable(t *testing.T) {
	pool := mustCreatePostgresConn(t)
	ctx := context.Background()

	sql, err := SessionMigration("admin_sessions")
	require.NoError(t, err)
	up, _, found := strings.Cut(sql, "-- +goose Down")
	require.True(t, found)
	_, err = pool.Exec(ctx, up)
	require.NoError(t, err)

	admin, err := NewSessionStore(pool, SessionStoreConfig{Table: "admin_sessions"})
	require.NoError(t, err)
	m, err := auth.NewOpaqueSessionManager(admin, time.Minute, time.Hour)
	require.NoError(t, err)

	first, err := m.Issue(ctx, testOwner, nil)
	require.NoError(t, err)
	second, err := m.Refresh(ctx, first.RefreshToken)
	require.NoError(t, err)
	_, err = m.VerifyAccess(ctx, second.AccessToken)
	require.NoError(t, err)

	defaultStore, err := NewSessionStore(pool)
	require.NoError(t, err)
	_, err = defaultStore.GetByRefresh(ctx, auth.HashOpaqueToken(second.RefreshToken))
	require.ErrorIs(t, err, auth.ErrSessionNotFound)

	_, err = m.Refresh(ctx, first.RefreshToken)
	require.Error(t, err)
	_, err = m.VerifyAccess(ctx, second.AccessToken)
	require.Error(t, err)
	require.NoError(t, m.RevokeOwner(ctx, testOwner))
	require.NoError(t, m.DeleteExpired(ctx))
}

func TestValidTable(t *testing.T) {
	for _, bad := range []string{"", "Auth", "x; DROP TABLE users", "1abc", strings.Repeat("a", 49)} {
		assert.Error(t, validTable(bad), bad)
		_, err := SessionMigration(bad)
		assert.Error(t, err, bad)
		_, err = OtpMigration(bad)
		assert.Error(t, err, bad)
		_, err = ApiKeyMigration(bad)
		assert.Error(t, err, bad)
	}
	assert.NoError(t, validTable(strings.Repeat("a", 48)))
}
