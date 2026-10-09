package auth

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/stretchr/testify/require"
)

type fakeTokenStore struct {
	token Token
}

func (s fakeTokenStore) StoreToken(_ context.Context, _ Token) error     { return nil }
func (s fakeTokenStore) StoreTokens(_ context.Context, _ ...Token) error { return nil }
func (s fakeTokenStore) GetTokenByReference(_ context.Context, _ uuid.UUID, _ int) (Token, error) {
	return s.token, nil
}
func (s fakeTokenStore) GetTokenByOwner(_ context.Context, _ uuid.UUID, _ int) (Token, error) {
	return s.token, nil
}
func (s fakeTokenStore) DeleteToken(_ context.Context, _ uuid.UUID) error            { return nil }
func (s fakeTokenStore) DeleteAllTokensByOwner(_ context.Context, _ uuid.UUID) error { return nil }

func referenceProviderWith(token Token) IReferenceTokenProvider {
	return NewJwtReferenceTokenProvider(nil, fakeTokenStore{token: token})
}

func storedToken(tokenType int, expiresAt time.Time) Token {
	return Token{
		Id:        testTokenID,
		OwnerId:   testOwner,
		Type:      tokenType,
		ExpiresAt: expiresAt,
		Claims:    map[string]any{},
	}
}

// The refresh flow calls GetRefreshToken directly without passing through the
// authentication middleware, so expiry has to be enforced by the provider.
func TestReferenceTokenProvider_GetRefreshToken_Expired(t *testing.T) {
	provider := referenceProviderWith(storedToken(RefreshTokenType, time.Now().UTC().Add(-time.Minute)))

	_, err := provider.GetRefreshToken(context.Background(), testTokenID)

	requireUnauthorized(t, err)
}

func TestReferenceTokenProvider_GetAccessToken_Expired(t *testing.T) {
	provider := referenceProviderWith(storedToken(AccessTokenType, time.Now().UTC().Add(-time.Minute)))

	_, err := provider.GetAccessToken(context.Background(), testTokenID)

	requireUnauthorized(t, err)
}

func TestReferenceTokenProvider_ZeroExpiryIsRejected(t *testing.T) {
	provider := referenceProviderWith(storedToken(AccessTokenType, time.Time{}))

	_, err := provider.GetAccessToken(context.Background(), testTokenID)

	requireUnauthorized(t, err)
}

func TestReferenceTokenProvider_GetAccessToken_NotExpired(t *testing.T) {
	provider := referenceProviderWith(storedToken(AccessTokenType, time.Now().UTC().Add(time.Hour)))

	token, err := provider.GetAccessToken(context.Background(), testTokenID)

	require.NoError(t, err)
	require.Equal(t, testOwner, token.OwnerId)
}

func requireUnauthorized(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var resultErr *ResultError
	require.ErrorAs(t, err, &resultErr)
	require.Equal(t, http.StatusUnauthorized, resultErr.Status)
}
