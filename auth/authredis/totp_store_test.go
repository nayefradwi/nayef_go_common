package authredis

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTotpStore(t *testing.T) {
	client := mustCreateRedisClient(t)
	store, err := NewTotpStore(client)
	require.NoError(t, err)
	ctx := context.Background()

	codes := make([][]byte, 10)
	for i := range codes {
		codes[i] = []byte{byte(i)}
	}

	t.Run("a step is used once", func(t *testing.T) {
		owner := uuid.New()
		require.NoError(t, store.Save(ctx, owner, []byte("sealed")))
		require.ErrorIs(t, store.UseStep(ctx, owner, 10), auth.ErrTotpNotFound)
		require.NoError(t, store.Confirm(ctx, owner, 10, nil))

		require.ErrorIs(t, store.UseStep(ctx, owner, 10), auth.ErrTotpNotFound)
		var ok atomic.Int32
		var wg sync.WaitGroup
		for range 20 {
			wg.Go(func() {
				if store.UseStep(ctx, owner, 11) == nil {
					ok.Add(1)
				}
			})
		}
		wg.Wait()
		assert.EqualValues(t, 1, ok.Load())
	})

	t.Run("a recovery code is used once", func(t *testing.T) {
		owner := uuid.New()
		require.NoError(t, store.Save(ctx, owner, []byte("sealed")))
		require.ErrorIs(t, store.SetRecoveryCodes(ctx, owner, codes[:1]), auth.ErrTotpNotFound)
		require.ErrorIs(t, store.UseRecoveryCode(ctx, owner, codes[0]), auth.ErrTotpNotFound)

		require.NoError(t, store.Confirm(ctx, owner, 1, codes))
		require.NoError(t, store.UseRecoveryCode(ctx, owner, codes[3]))
		require.ErrorIs(t, store.UseRecoveryCode(ctx, owner, codes[3]), auth.ErrTotpNotFound)

		got, err := store.Get(ctx, owner)
		require.NoError(t, err)
		assert.Len(t, got.RecoveryCodes, 9)
	})

	t.Run("save cannot replace a confirmed secret", func(t *testing.T) {
		owner := uuid.New()
		require.NoError(t, store.Save(ctx, owner, []byte("first")))
		require.NoError(t, store.Save(ctx, owner, []byte("second")))
		require.NoError(t, store.Confirm(ctx, owner, 1, nil))

		require.ErrorIs(t, store.Save(ctx, owner, []byte("attacker")), auth.ErrTotpEnabled)
		got, err := store.Get(ctx, owner)
		require.NoError(t, err)
		assert.Equal(t, []byte("second"), got.Secret)
		assert.False(t, got.ConfirmedAt.IsZero())
	})

	t.Run("delete removes the recovery codes", func(t *testing.T) {
		owner := uuid.New()
		require.NoError(t, store.Save(ctx, owner, []byte("sealed")))
		require.NoError(t, store.Confirm(ctx, owner, 1, codes))
		require.NoError(t, store.Delete(ctx, owner))

		_, err := store.Get(ctx, owner)
		require.ErrorIs(t, err, auth.ErrTotpNotFound)
		require.ErrorIs(t, store.UseRecoveryCode(ctx, owner, codes[0]), auth.ErrTotpNotFound)
	})
}

func TestTotpStore_CustomPrefix(t *testing.T) {
	client := mustCreateRedisClient(t)
	ctx := context.Background()

	admin, err := NewTotpStore(client, TotpStoreConfig{Prefix: "admin:totp:"})
	require.NoError(t, err)
	require.NoError(t, admin.Save(ctx, testOwner, []byte("sealed")))

	defaultStore, err := NewTotpStore(client)
	require.NoError(t, err)
	_, err = defaultStore.Get(ctx, testOwner)
	require.ErrorIs(t, err, auth.ErrTotpNotFound)
}
