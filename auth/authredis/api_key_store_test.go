package authredis

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiKeyStore(t *testing.T) {
	client := mustCreateRedisClient(t)
	store, err := NewApiKeyStore(client)
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
		assert.Equal(t, issued.Id, got.Id)
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
		require.ErrorIs(t, store.Delete(ctx, testOwner, issued.Id), auth.ErrApiKeyNotFound)
		n, err := client.Exists(ctx, defaultApiKeyPrefix+"h:"+string(issued.Hash)).Result()
		require.NoError(t, err)
		assert.Zero(t, n)
	})

	t.Run("touch does not bring back a deleted key", func(t *testing.T) {
		_, issued, err := m.Issue(ctx, testOwner, "ci", nil, time.Time{})
		require.NoError(t, err)
		require.NoError(t, store.Delete(ctx, testOwner, issued.Id))

		require.NoError(t, store.Touch(ctx, issued.Id))
		n, err := client.Exists(ctx, defaultApiKeyPrefix+"k:"+issued.Id.String()).Result()
		require.NoError(t, err)
		assert.Zero(t, n)
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

	t.Run("expired key is gone, keys that never expire stay", func(t *testing.T) {
		owner := uuid.New()
		_, forever, err := m.Issue(ctx, owner, "forever", nil, time.Time{})
		require.NoError(t, err)
		_, dead, err := m.Issue(ctx, owner, "dead", nil, time.Now().Add(-time.Second))
		require.NoError(t, err)

		_, err = store.GetByHash(ctx, forever.Hash)
		require.NoError(t, err)
		_, err = store.GetByHash(ctx, dead.Hash)
		require.ErrorIs(t, err, auth.ErrApiKeyNotFound)

		keys, err := store.ListByOwner(ctx, owner)
		require.NoError(t, err)
		require.Len(t, keys, 1)
		assert.Equal(t, forever.Id, keys[0].Id)
	})

	t.Run("owner index expires with its longest-lived key", func(t *testing.T) {
		owner := uuid.New()
		index := defaultApiKeyPrefix + "o:" + owner.String()
		_, short, err := m.Issue(ctx, owner, "short", nil, time.Now().Add(300*time.Millisecond))
		require.NoError(t, err)
		_, _, err = m.Issue(ctx, owner, "shorter", nil, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)

		at, err := client.PExpireTime(ctx, index).Result()
		require.NoError(t, err)
		assert.Equal(t, short.ExpiresAt.UnixMilli(), at.Milliseconds())

		time.Sleep(400 * time.Millisecond)
		n, err := client.Exists(ctx, index).Result()
		require.NoError(t, err)
		assert.Zero(t, n)

		_, _, err = m.Issue(ctx, owner, "expiring", nil, time.Now().Add(time.Hour))
		require.NoError(t, err)
		_, _, err = m.Issue(ctx, owner, "forever", nil, time.Time{})
		require.NoError(t, err)
		ttl, err := client.PTTL(ctx, index).Result()
		require.NoError(t, err)
		assert.Equal(t, time.Duration(-1), ttl)
	})
}

func TestApiKeyStore_CustomPrefix(t *testing.T) {
	client := mustCreateRedisClient(t)
	ctx := context.Background()

	partner, err := NewApiKeyStore(client, ApiKeyStoreConfig{Prefix: "partner:keys:"})
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

	defaultStore, err := NewApiKeyStore(client)
	require.NoError(t, err)
	_, err = defaultStore.GetByHash(ctx, issued.Hash)
	require.ErrorIs(t, err, auth.ErrApiKeyNotFound)

	require.NoError(t, m.Revoke(ctx, testOwner, issued.Id))
	require.NoError(t, m.DeleteExpired(ctx))
}
