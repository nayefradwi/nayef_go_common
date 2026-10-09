package authpg

import (
	"context"
	"github.com/nayefradwi/nayef_go_common/auth"
	"net/http"
	"testing"

	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReferenceTokenProvider_GenerateToken(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	dto, err := env.provider.GenerateToken(ctx, testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	assert.NotEmpty(t, dto.AccessToken, "expected non-empty access token ID")
	assert.NotEmpty(t, dto.RefreshToken, "expected non-empty refresh token ID")

	accessToken, err := env.store.GetTokenByReference(ctx, mustUUID(dto.AccessToken), auth.AccessTokenType)
	require.NoError(t, err)
	assert.Equal(t, testOwner, accessToken.OwnerId)

	refreshToken, err := env.store.GetTokenByReference(ctx, mustUUID(dto.RefreshToken), auth.RefreshTokenType)
	require.NoError(t, err)
	assert.Equal(t, testOwner, refreshToken.OwnerId)
}

func TestReferenceTokenProvider_GetAccessToken(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	dto, err := env.provider.GenerateToken(ctx, testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	token, err := env.provider.GetAccessToken(ctx, mustUUID(dto.AccessToken))
	require.NoError(t, err)

	assert.Equal(t, testOwner, token.OwnerId)
	assert.Equal(t, auth.AccessTokenType, token.Type)
}

func TestReferenceTokenProvider_GetRefreshToken(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	dto, err := env.provider.GenerateToken(ctx, testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	token, err := env.provider.GetRefreshToken(ctx, mustUUID(dto.RefreshToken))
	require.NoError(t, err)

	assert.Equal(t, testOwner, token.OwnerId)
	assert.Equal(t, auth.RefreshTokenType, token.Type)
}

func TestReferenceTokenProvider_RevokeToken(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	dto, err := env.provider.GenerateToken(ctx, testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	err = env.provider.RevokeToken(ctx, mustUUID(dto.AccessToken))
	require.NoError(t, err)

	_, err = env.provider.GetAccessToken(ctx, mustUUID(dto.AccessToken))
	require.Error(t, err)

	_, err = env.provider.GetRefreshToken(ctx, mustUUID(dto.RefreshToken))
	require.NoError(t, err)
}

func TestReferenceTokenProvider_RevokeOwner(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	dto, err := env.provider.GenerateToken(ctx, testOwner, map[string]any{"role": "admin"})
	require.NoError(t, err)

	err = env.provider.RevokeOwner(ctx, testOwner)
	require.NoError(t, err)

	_, err = env.provider.GetAccessToken(ctx, mustUUID(dto.AccessToken))
	require.Error(t, err)

	_, err = env.provider.GetRefreshToken(ctx, mustUUID(dto.RefreshToken))
	require.Error(t, err)
}

func TestTokenStore_DbErrorIsNotNotFound(t *testing.T) {
	env := setupTestEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := env.store.GetTokenByReference(ctx, testTokenID, auth.AccessTokenType)

	var resultErr *ResultError
	require.ErrorAs(t, err, &resultErr)
	assert.Equal(t, http.StatusInternalServerError, resultErr.Status)
}
