package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testEnv struct {
	provider IReferenceTokenProvider
	store    ITokenStore
}

func setupTestEnv(t *testing.T) testEnv {
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
	refreshTokenProvider := NewJwtRefreshTokenProvider(refreshProvider, accessProvider)
	referenceProvider := NewJwtReferenceTokenProvider(refreshTokenProvider, store)

	return testEnv{
		provider: referenceProvider,
		store:    store,
	}
}
