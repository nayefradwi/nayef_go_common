package auth

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

type JwtRefreshTokenWithRevokeProvider struct {
	TokenProvider IRefreshTokenProvider
	TokenStore    ITokenStore
}

func NewJwtRefreshTokenWithRevokeProvider(tokenProvider IRefreshTokenProvider, tokenStore ITokenStore) IRefreshTokenProviderWithRevoke {
	return JwtRefreshTokenWithRevokeProvider{
		TokenProvider: tokenProvider,
		TokenStore:    tokenStore,
	}
}

func (t JwtRefreshTokenWithRevokeProvider) GenerateId() (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, InternalError("failed to generate token id: " + err.Error())
	}
	return id, nil
}

func (t JwtRefreshTokenWithRevokeProvider) GenerateToken(ownerId uuid.UUID, claims map[string]any) (TokenDTO, error) {
	tokenPair, err := t.TokenProvider.GenerateToken(ownerId, claims)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	refreshToken, err := t.TokenProvider.GetRefreshToken(tokenPair.RefreshToken)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	refreshToken.Id, err = t.GenerateId()
	if err != nil {
		return EmptyTokenDTO(), err
	}

	if err := t.TokenStore.StoreToken(refreshToken); err != nil {
		return EmptyTokenDTO(), err
	}

	return NewTokenDTOWithRefresh(tokenPair.AccessToken, refreshToken.Id.String()), nil
}

func (t JwtRefreshTokenWithRevokeProvider) GetAccessToken(accessToken string) (Token, error) {
	return t.TokenProvider.GetAccessToken(accessToken)
}

func (t JwtRefreshTokenWithRevokeProvider) GetRefreshToken(refreshToken string) (Token, error) {
	return t.TokenProvider.GetRefreshToken(refreshToken)
}

func (t JwtRefreshTokenWithRevokeProvider) GetAccessTokenProvider() ITokenProvider {
	return t.TokenProvider.GetAccessTokenProvider()
}

func (t JwtRefreshTokenWithRevokeProvider) RevokeToken(reference uuid.UUID) error {
	return t.TokenStore.DeleteToken(reference)
}

func (t JwtRefreshTokenWithRevokeProvider) RevokeOwner(ownerId uuid.UUID) error {
	return t.TokenStore.DeleteAllTokensByOwner(ownerId)
}

func (t JwtRefreshTokenWithRevokeProvider) WithTx(tx pgx.Tx) IRefreshTokenProviderWithRevoke {
	t.TokenStore = t.TokenStore.WithTx(tx)
	return t
}
