package authpg

import (
	"github.com/nayefradwi/nayef_go_common/auth"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testEnv struct {
	provider auth.IReferenceTokenProvider
	store    auth.ITokenStore
}

func setupTestEnv(t *testing.T) testEnv {
	t.Helper()

	pool := mustCreatePostgresConn(t)
	store := NewTokenStore(pool)

	accessCfg, err := auth.NewJwtTokenProviderConfig("test-access-secret-key-32-bytes-long", time.Hour, auth.AccessTokenType)
	require.NoError(t, err)

	refreshCfg, err := auth.NewJwtTokenProviderConfig("test-refresh-secret-key-32-bytes-long", 24*time.Hour, auth.RefreshTokenType)
	require.NoError(t, err)

	accessProvider := auth.NewJwtTokenProvider(accessCfg)
	refreshProvider := auth.NewJwtTokenProvider(refreshCfg)
	refreshTokenProvider := auth.NewJwtRefreshTokenProvider(refreshProvider, accessProvider)
	referenceProvider := auth.NewJwtReferenceTokenProvider(refreshTokenProvider, store)

	return testEnv{
		provider: referenceProvider,
		store:    store,
	}
}
