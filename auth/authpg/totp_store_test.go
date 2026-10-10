package authpg

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTotpStore(t *testing.T) {
	pool := mustCreatePostgresConn(t)
	store, err := NewTotpStore(pool)
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("a step is used once", func(t *testing.T) {
		owner := uuid.New()
		require.NoError(t, store.Save(ctx, owner, []byte("sealed")))
		require.ErrorIs(t, store.UseStep(ctx, owner, 10), auth.ErrTotpNotFound)
		require.NoError(t, store.Confirm(ctx, owner, 10, nil))

		require.ErrorIs(t, store.UseStep(ctx, owner, 10), auth.ErrTotpNotFound)
		require.NoError(t, store.UseStep(ctx, owner, 11))
		require.ErrorIs(t, store.UseStep(ctx, owner, 11), auth.ErrTotpNotFound)
	})

	t.Run("a recovery code is used once", func(t *testing.T) {
		owner := uuid.New()
		require.NoError(t, store.Save(ctx, owner, []byte("sealed")))
		require.ErrorIs(t, store.SetRecoveryCodes(ctx, owner, [][]byte{[]byte("a")}), auth.ErrTotpNotFound)
		require.ErrorIs(t, store.UseRecoveryCode(ctx, owner, []byte("a")), auth.ErrTotpNotFound)

		codes := make([][]byte, 10)
		for i := range codes {
			codes[i] = []byte{byte(i)}
		}
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

		require.NoError(t, store.Delete(ctx, owner))
		_, err = store.Get(ctx, owner)
		require.ErrorIs(t, err, auth.ErrTotpNotFound)
	})
}

func TestTotpStore_CustomTable(t *testing.T) {
	pool := mustCreatePostgresConn(t)
	ctx := context.Background()

	sql, err := TotpMigration("admin_totp")
	require.NoError(t, err)
	up, _, found := strings.Cut(sql, "-- +goose Down")
	require.True(t, found)
	_, err = pool.Exec(ctx, up)
	require.NoError(t, err)

	admin, err := NewTotpStore(pool, TotpStoreConfig{Table: "admin_totp"})
	require.NoError(t, err)
	require.NoError(t, admin.Save(ctx, testOwner, []byte("sealed")))

	defaultStore, err := NewTotpStore(pool)
	require.NoError(t, err)
	_, err = defaultStore.Get(ctx, testOwner)
	require.ErrorIs(t, err, auth.ErrTotpNotFound)
}
