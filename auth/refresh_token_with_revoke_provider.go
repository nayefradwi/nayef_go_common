package auth

import (
	"github.com/google/uuid"
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

func NewDefaultJwtRefreshTokenWithRevokeProvider(tokenStore ITokenStore) IRefreshTokenProviderWithRevoke {
	return NewJwtRefreshTokenWithRevokeProvider(
		NewDefaultJwtRefreshTokenProvider(),
		tokenStore,
	)
}

func (t JwtRefreshTokenWithRevokeProvider) GenerateId() string {
	id, _ := uuid.NewV7()
	return id.String()
}

func (t JwtRefreshTokenWithRevokeProvider) GenerateToken(ownerId string, claims map[string]interface{}) (TokenDTO, error) {
	TokenPair, err := t.TokenProvider.GenerateToken(ownerId, claims)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	accessJwt, refreshJwt := TokenPair.AccessToken, TokenPair.RefreshToken
	refreshToken, err := t.TokenProvider.GetRefreshToken(refreshJwt)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	refreshToken.Id = t.GenerateId()
	if err := t.TokenStore.StoreToken(refreshToken); err != nil {
		return EmptyTokenDTO(), err
	}

	return NewTokenDTOWithRefresh(accessJwt, refreshToken.Id), nil
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

func (t JwtRefreshTokenWithRevokeProvider) RevokeToken(reference string) error {
	if err := t.TokenStore.DeleteToken(reference); err != nil {
		return err
	}
	return nil
}

func (t JwtRefreshTokenWithRevokeProvider) RevokeOwner(ownerId string) error {
	if err := t.TokenStore.DeleteAllTokensByOwner(ownerId); err != nil {
		return err
	}
	return nil
}
