package authredis

import (
	"context"
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
	store, err := NewSessionStore(mustCreateRedisClient(t))
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
					assert.Equal(t, s.OwnerId, rotated.OwnerId)
					assert.Equal(t, "admin", rotated.Claims["role"])
				case assert.ErrorIs(t, err, auth.ErrSessionNotFound):
					notFound.Add(1)
				}
			})
		}
		wg.Wait()
		assert.EqualValues(t, 1, ok.Load())
		assert.EqualValues(t, n-1, notFound.Load())

		got, err := store.GetByRefresh(ctx, s.RefreshHash)
		require.NoError(t, err)
		assert.False(t, got.RotatedAt.IsZero())
	})

	t.Run("expired refresh is gone without DeleteExpired", func(t *testing.T) {
		s := newTestSession(t, testOwner, time.Now().Add(-time.Second))
		require.NoError(t, store.Create(ctx, s))

		_, err := store.Rotate(ctx, s.RefreshHash, nextSession(t))
		require.ErrorIs(t, err, auth.ErrSessionNotFound)
		_, err = store.GetByRefresh(ctx, s.RefreshHash)
		require.ErrorIs(t, err, auth.ErrSessionNotFound)
		_, err = store.GetByAccess(ctx, s.AccessHash)
		require.ErrorIs(t, err, auth.ErrSessionNotFound)
	})

	t.Run("unknown hash is not found, not internal", func(t *testing.T) {
		_, err := store.GetByRefresh(ctx, auth.HashOpaqueToken("missing"))
		require.ErrorIs(t, err, auth.ErrSessionNotFound)
		_, err = store.GetByAccess(ctx, auth.HashOpaqueToken("missing"))
		require.ErrorIs(t, err, auth.ErrSessionNotFound)
	})

	t.Run("jwt mode session without access round-trips", func(t *testing.T) {
		s := newTestSession(t, testOwner, time.Now().Add(time.Hour))
		s.AccessHash, s.AccessExpiresAt, s.Claims = nil, time.Time{}, nil
		require.NoError(t, store.Create(ctx, s))

		got, err := store.GetByRefresh(ctx, s.RefreshHash)
		require.NoError(t, err)
		assert.Nil(t, got.AccessHash)
		assert.Nil(t, got.Claims)
		assert.True(t, got.AccessExpiresAt.IsZero())

		rotated, err := store.Rotate(ctx, s.RefreshHash, nextSession(t))
		require.NoError(t, err)
		assert.Equal(t, s.FamilyId, rotated.FamilyId)
	})

	t.Run("delete family leaves other families", func(t *testing.T) {
		a := newTestSession(t, testOwner, time.Now().Add(time.Hour))
		b := newTestSession(t, testOwner, time.Now().Add(time.Hour))
		require.NoError(t, store.Create(ctx, a))
		require.NoError(t, store.Create(ctx, b))

		require.NoError(t, store.DeleteFamily(ctx, a.FamilyId))
		_, err := store.GetByRefresh(ctx, a.RefreshHash)
		require.ErrorIs(t, err, auth.ErrSessionNotFound)
		_, err = store.GetByAccess(ctx, a.AccessHash)
		require.ErrorIs(t, err, auth.ErrSessionNotFound)
		_, err = store.Rotate(ctx, a.RefreshHash, nextSession(t))
		require.ErrorIs(t, err, auth.ErrSessionNotFound)

		_, err = store.GetByRefresh(ctx, b.RefreshHash)
		require.NoError(t, err)
	})

	t.Run("delete owner ends every family of that owner only", func(t *testing.T) {
		owner, other := uuid.New(), uuid.New()
		a := newTestSession(t, owner, time.Now().Add(time.Hour))
		b := newTestSession(t, owner, time.Now().Add(time.Hour))
		c := newTestSession(t, other, time.Now().Add(time.Hour))
		for _, s := range []auth.Session{a, b, c} {
			require.NoError(t, store.Create(ctx, s))
		}
		rotated, err := store.Rotate(ctx, a.RefreshHash, nextSession(t))
		require.NoError(t, err)

		require.NoError(t, store.DeleteOwner(ctx, owner))
		for _, hash := range [][]byte{a.RefreshHash, b.RefreshHash, rotated.RefreshHash} {
			_, err := store.GetByRefresh(ctx, hash)
			require.ErrorIs(t, err, auth.ErrSessionNotFound)
		}
		_, err = store.GetByRefresh(ctx, c.RefreshHash)
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
		_, err = m.Refresh(ctx, second.RefreshToken)
		require.Error(t, err)
	})
}

func TestSessionStore_CustomPrefix(t *testing.T) {
	client := mustCreateRedisClient(t)
	ctx := context.Background()

	admin, err := NewSessionStore(client, SessionStoreConfig{Prefix: "admin:sessions:"})
	require.NoError(t, err)
	m, err := auth.NewOpaqueSessionManager(admin, time.Minute, time.Hour)
	require.NoError(t, err)

	first, err := m.Issue(ctx, testOwner, nil)
	require.NoError(t, err)
	second, err := m.Refresh(ctx, first.RefreshToken)
	require.NoError(t, err)

	defaultStore, err := NewSessionStore(client)
	require.NoError(t, err)
	_, err = defaultStore.GetByRefresh(ctx, auth.HashOpaqueToken(second.RefreshToken))
	require.ErrorIs(t, err, auth.ErrSessionNotFound)

	require.NoError(t, m.RevokeOwner(ctx, testOwner))
	_, err = m.VerifyAccess(ctx, second.AccessToken)
	require.Error(t, err)
}
