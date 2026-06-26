package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type revokeTestEnv struct {
	provider IRefreshTokenProviderWithRevoke
	store    ITokenStore
}

func setupRevokeTestEnv(t *testing.T) revokeTestEnv {
	t.Helper()

	pool := mustCreatePostgresConn(t)
	store := NewPostgresTokenStore(pool)

	accessCfg, err := NewJwtTokenProviderConfig("test-access-secret-key", time.Hour)
	require.NoError(t, err)
	accessCfg = accessCfg.SetTokenType(AccessTokenType)

	refreshCfg, err := NewJwtTokenProviderConfig("test-refresh-secret-key", 24*time.Hour)
	require.NoError(t, err)
	refreshCfg = refreshCfg.SetTokenType(RefreshTokenType)

	accessProvider := NewJwtTokenProvider(accessCfg)
	refreshProvider := NewJwtTokenProvider(refreshCfg)
	jwtRefreshProvider := NewJwtRefreshTokenProvider(refreshProvider, accessProvider)
	provider := NewJwtRefreshTokenWithRevokeProvider(jwtRefreshProvider, store)

	return revokeTestEnv{
		provider: provider,
		store:    store,
	}
}

func TestJwtRefreshTokenWithRevoke_GenerateToken(t *testing.T) {
	env := setupRevokeTestEnv(t)

	dto, err := env.provider.GenerateToken(testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	assert.NotEmpty(t, dto.AccessToken, "expected non-empty access token (JWT)")
	assert.NotEmpty(t, dto.RefreshToken, "expected non-empty refresh token (UUID)")

	stored, err := env.store.GetTokenByReference(mustUUID(dto.RefreshToken), RefreshTokenType)
	require.NoError(t, err)
	assert.Equal(t, testOwner, stored.OwnerId)
	assert.Equal(t, RefreshTokenType, stored.Type)
}

func TestJwtRefreshTokenWithRevoke_GetAccessToken(t *testing.T) {
	env := setupRevokeTestEnv(t)

	dto, err := env.provider.GenerateToken(testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	token, err := env.provider.GetAccessToken(dto.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, testOwner, token.OwnerId)
	assert.Equal(t, AccessTokenType, token.Type)
}

func TestJwtRefreshTokenWithRevoke_GetRefreshToken(t *testing.T) {
	env := setupRevokeTestEnv(t)

	dto, err := env.provider.GenerateToken(testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	stored, err := env.store.GetTokenByReference(mustUUID(dto.RefreshToken), RefreshTokenType)
	require.NoError(t, err)

	token, err := env.provider.GetRefreshToken(stored.Value)
	require.NoError(t, err)

	assert.Equal(t, testOwner, token.OwnerId)
	assert.Equal(t, RefreshTokenType, token.Type)
}

func TestJwtRefreshTokenWithRevoke_RevokeToken(t *testing.T) {
	env := setupRevokeTestEnv(t)

	dto, err := env.provider.GenerateToken(testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	err = env.provider.RevokeToken(mustUUID(dto.RefreshToken))
	require.NoError(t, err)

	_, err = env.store.GetTokenByReference(mustUUID(dto.RefreshToken), RefreshTokenType)
	require.Error(t, err)

	token, err := env.provider.GetAccessToken(dto.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, testOwner, token.OwnerId)
}

func TestJwtRefreshTokenWithRevoke_RevokeOwner(t *testing.T) {
	env := setupRevokeTestEnv(t)

	dto1, err := env.provider.GenerateToken(testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	dto2, err := env.provider.GenerateToken(testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	err = env.provider.RevokeOwner(testOwner)
	require.NoError(t, err)

	_, err = env.store.GetTokenByReference(mustUUID(dto1.RefreshToken), RefreshTokenType)
	require.Error(t, err)

	_, err = env.store.GetTokenByReference(mustUUID(dto2.RefreshToken), RefreshTokenType)
	require.Error(t, err)
}
