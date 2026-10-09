package authpg

import (
	"context"
	"github.com/nayefradwi/nayef_go_common/auth"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type revokeTestEnv struct {
	provider auth.IRefreshTokenProviderWithRevoke
	store    auth.ITokenStore
}

func setupRevokeTestEnv(t *testing.T) revokeTestEnv {
	t.Helper()

	pool := mustCreatePostgresConn(t)
	store := NewTokenStore(pool)

	accessCfg, err := auth.NewJwtTokenProviderConfig("test-access-secret-key-32-bytes-long", time.Hour, auth.AccessTokenType)
	require.NoError(t, err)

	refreshCfg, err := auth.NewJwtTokenProviderConfig("test-refresh-secret-key-32-bytes-long", 24*time.Hour, auth.RefreshTokenType)
	require.NoError(t, err)

	accessProvider := auth.NewJwtTokenProvider(accessCfg)
	refreshProvider := auth.NewJwtTokenProvider(refreshCfg)
	jwtRefreshProvider := auth.NewJwtRefreshTokenProvider(refreshProvider, accessProvider)
	provider := auth.NewJwtRefreshTokenWithRevokeProvider(jwtRefreshProvider, store)

	return revokeTestEnv{
		provider: provider,
		store:    store,
	}
}

func TestJwtRefreshTokenWithRevoke_GenerateToken(t *testing.T) {
	env := setupRevokeTestEnv(t)
	ctx := context.Background()

	dto, err := env.provider.GenerateToken(ctx, testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	assert.NotEmpty(t, dto.AccessToken, "expected non-empty access token (JWT)")
	assert.NotEmpty(t, dto.RefreshToken, "expected non-empty refresh token (UUID)")

	stored, err := env.store.GetTokenByReference(ctx, mustUUID(dto.RefreshToken), auth.RefreshTokenType)
	require.NoError(t, err)
	assert.Equal(t, testOwner, stored.OwnerId)
	assert.Equal(t, auth.RefreshTokenType, stored.Type)
}

func TestJwtRefreshTokenWithRevoke_GetAccessToken(t *testing.T) {
	env := setupRevokeTestEnv(t)
	ctx := context.Background()

	dto, err := env.provider.GenerateToken(ctx, testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	token, err := env.provider.GetAccessToken(dto.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, testOwner, token.OwnerId)
	assert.Equal(t, auth.AccessTokenType, token.Type)
}

func TestJwtRefreshTokenWithRevoke_GetRefreshToken(t *testing.T) {
	env := setupRevokeTestEnv(t)
	ctx := context.Background()

	dto, err := env.provider.GenerateToken(ctx, testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	stored, err := env.store.GetTokenByReference(ctx, mustUUID(dto.RefreshToken), auth.RefreshTokenType)
	require.NoError(t, err)

	token, err := env.provider.GetRefreshToken(stored.Value)
	require.NoError(t, err)

	assert.Equal(t, testOwner, token.OwnerId)
	assert.Equal(t, auth.RefreshTokenType, token.Type)
}

func TestJwtRefreshTokenWithRevoke_RevokeToken(t *testing.T) {
	env := setupRevokeTestEnv(t)
	ctx := context.Background()

	dto, err := env.provider.GenerateToken(ctx, testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	err = env.provider.RevokeToken(ctx, mustUUID(dto.RefreshToken))
	require.NoError(t, err)

	_, err = env.store.GetTokenByReference(ctx, mustUUID(dto.RefreshToken), auth.RefreshTokenType)
	require.Error(t, err)

	token, err := env.provider.GetAccessToken(dto.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, testOwner, token.OwnerId)
}

func TestJwtRefreshTokenWithRevoke_RevokeOwner(t *testing.T) {
	env := setupRevokeTestEnv(t)
	ctx := context.Background()

	dto1, err := env.provider.GenerateToken(ctx, testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	dto2, err := env.provider.GenerateToken(ctx, testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	err = env.provider.RevokeOwner(ctx, testOwner)
	require.NoError(t, err)

	_, err = env.store.GetTokenByReference(ctx, mustUUID(dto1.RefreshToken), auth.RefreshTokenType)
	require.Error(t, err)

	_, err = env.store.GetTokenByReference(ctx, mustUUID(dto2.RefreshToken), auth.RefreshTokenType)
	require.Error(t, err)
}
