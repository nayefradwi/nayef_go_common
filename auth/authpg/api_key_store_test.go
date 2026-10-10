package authpg

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiKeyStore(t *testing.T) {
	pool := mustCreatePostgresConn(t)
	store, err := NewApiKeyStore(pool)
	require.NoError(t, err)
	m, err := auth.NewApiKeyManager(store, "sk_live")
	require.NoError(t, err)
	ctx := context.Background()
	other := uuid.New()

	t.Run("key without expiry or claims round-trips", func(t *testing.T) {
		key, issued, err := m.Issue(ctx, testOwner, "ci", nil, time.Time{})
		require.NoError(t, err)

		got, err := store.GetByHash(ctx, issued.Hash)
		require.NoError(t, err)
		assert.True(t, got.ExpiresAt.IsZero())
		assert.True(t, got.LastUsedAt.IsZero())
		assert.Nil(t, got.Claims)

		id, err := m.VerifyKey(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, auth.KindKey, id.Kind)

		got, err = store.GetByHash(ctx, issued.Hash)
		require.NoError(t, err)
		assert.False(t, got.LastUsedAt.IsZero())
	})

	t.Run("delete is scoped to the owner", func(t *testing.T) {
		key, issued, err := m.Issue(ctx, testOwner, "ci", nil, time.Time{})
		require.NoError(t, err)

		require.ErrorIs(t, store.Delete(ctx, other, issued.Id), auth.ErrApiKeyNotFound)
		_, err = m.VerifyKey(ctx, key)
		require.NoError(t, err)

		require.NoError(t, store.Delete(ctx, testOwner, issued.Id))
		_, err = m.VerifyKey(ctx, key)
		require.Error(t, err)
	})

	t.Run("list returns only the owner's keys, newest first", func(t *testing.T) {
		_, first, err := m.Issue(ctx, other, "first", nil, time.Time{})
		require.NoError(t, err)
		_, second, err := m.Issue(ctx, other, "second", map[string]any{"scopes": []any{"read"}}, time.Time{})
		require.NoError(t, err)

		keys, err := store.ListByOwner(ctx, other)
		require.NoError(t, err)
		require.Len(t, keys, 2)
		assert.Equal(t, second.Id, keys[0].Id)
		assert.Equal(t, first.Id, keys[1].Id)
		assert.Equal(t, []any{"read"}, keys[0].Claims["scopes"])

		none, err := store.ListByOwner(ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, none)
	})

	t.Run("delete expired keeps keys that never expire", func(t *testing.T) {
		owner := uuid.New()
		_, forever, err := m.Issue(ctx, owner, "forever", nil, time.Time{})
		require.NoError(t, err)
		_, dead, err := m.Issue(ctx, owner, "dead", nil, time.Now().Add(-time.Second))
		require.NoError(t, err)

		require.NoError(t, store.DeleteExpired(ctx))
		_, err = store.GetByHash(ctx, forever.Hash)
		require.NoError(t, err)
		_, err = store.GetByHash(ctx, dead.Hash)
		require.ErrorIs(t, err, auth.ErrApiKeyNotFound)
	})
}

func TestApiKeyStore_CustomTable(t *testing.T) {
	pool := mustCreatePostgresConn(t)
	ctx := context.Background()

	sql, err := ApiKeyMigration("partner_keys")
	require.NoError(t, err)
	up, _, found := strings.Cut(sql, "-- +goose Down")
	require.True(t, found)
	_, err = pool.Exec(ctx, up)
	require.NoError(t, err)

	partner, err := NewApiKeyStore(pool, ApiKeyStoreConfig{Table: "partner_keys"})
	require.NoError(t, err)
	m, err := auth.NewApiKeyManager(partner, "pk")
	require.NoError(t, err)

	key, issued, err := m.Issue(ctx, testOwner, "partner", nil, time.Now().Add(time.Hour))
	require.NoError(t, err)
	_, err = m.VerifyKey(ctx, key)
	require.NoError(t, err)
	keys, err := m.List(ctx, testOwner)
	require.NoError(t, err)
	require.Len(t, keys, 1)

	defaultStore, err := NewApiKeyStore(pool)
	require.NoError(t, err)
	_, err = defaultStore.GetByHash(ctx, issued.Hash)
	require.ErrorIs(t, err, auth.ErrApiKeyNotFound)

	require.NoError(t, m.Revoke(ctx, testOwner, issued.Id))
	require.NoError(t, m.DeleteExpired(ctx))
}
