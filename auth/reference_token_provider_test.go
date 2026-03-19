package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReferenceTokenProvider_GenerateToken(t *testing.T) {
	env := setupTestEnv(t)

	dto, err := env.provider.GenerateToken("owner1", map[string]any{"role": "admin"})
	require.NoError(t, err)

	assert.NotEmpty(t, dto.AccessToken, "expected non-empty access token ID")
	assert.NotEmpty(t, dto.RefreshToken, "expected non-empty refresh token ID")

	accessToken, err := env.store.GetTokenByReference(dto.AccessToken, AccessTokenType)
	require.NoError(t, err)
	assert.Equal(t, "owner1", accessToken.OwnerId)

	refreshToken, err := env.store.GetTokenByReference(dto.RefreshToken, RefreshTokenType)
	require.NoError(t, err)
	assert.Equal(t, "owner1", refreshToken.OwnerId)
}

func TestReferenceTokenProvider_GetAccessToken(t *testing.T) {
	env := setupTestEnv(t)

	dto, err := env.provider.GenerateToken("owner1", map[string]any{"role": "admin"})
	require.NoError(t, err)

	token, err := env.provider.GetAccessToken(dto.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, "owner1", token.OwnerId)
	assert.Equal(t, AccessTokenType, token.Type)
}

func TestReferenceTokenProvider_GetRefreshToken(t *testing.T) {
	env := setupTestEnv(t)

	dto, err := env.provider.GenerateToken("owner1", map[string]any{"role": "admin"})
	require.NoError(t, err)

	token, err := env.provider.GetRefreshToken(dto.RefreshToken)
	require.NoError(t, err)

	assert.Equal(t, "owner1", token.OwnerId)
	assert.Equal(t, RefreshTokenType, token.Type)
}

func TestReferenceTokenProvider_RevokeToken(t *testing.T) {
	env := setupTestEnv(t)

	dto, err := env.provider.GenerateToken("owner1", map[string]any{"role": "admin"})
	require.NoError(t, err)

	err = env.provider.RevokeToken(dto.AccessToken)
	require.NoError(t, err)

	_, err = env.provider.GetAccessToken(dto.AccessToken)
	require.Error(t, err)

	_, err = env.provider.GetRefreshToken(dto.RefreshToken)
	require.NoError(t, err)
}

func TestReferenceTokenProvider_RevokeOwner(t *testing.T) {
	env := setupTestEnv(t)

	dto, err := env.provider.GenerateToken("owner1", map[string]any{"role": "admin"})
	require.NoError(t, err)

	err = env.provider.RevokeOwner("owner1")
	require.NoError(t, err)

	_, err = env.provider.GetAccessToken(dto.AccessToken)
	require.Error(t, err)

	_, err = env.provider.GetRefreshToken(dto.RefreshToken)
	require.Error(t, err)
}
