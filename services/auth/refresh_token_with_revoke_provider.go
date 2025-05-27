package auth

import (
	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/modules/auth"
)

type JwtRefreshTokenWithRevokeProvider struct {
	TokenProvider auth.IRefreshTokenProvider
	TokenStore    auth.ITokenStore
}

func NewJwtRefreshTokenWithRevokeProvider(tokenProvider auth.IRefreshTokenProvider, tokenStore auth.ITokenStore) auth.IRefreshTokenProviderWithRevoke {
	return JwtRefreshTokenWithRevokeProvider{
		TokenProvider: tokenProvider,
		TokenStore:    tokenStore,
	}
}

func NewDefaultJwtRefreshTokenWithRevokeProvider(tokenStore auth.ITokenStore) auth.IRefreshTokenProviderWithRevoke {
	return NewJwtRefreshTokenWithRevokeProvider(
		NewDefaultJwtRefreshTokenProvider(),
		tokenStore,
	)
}

func (t JwtRefreshTokenWithRevokeProvider) GenerateId() string {
	id, _ := uuid.NewV7()
	return id.String()
}

func (t JwtRefreshTokenWithRevokeProvider) GenerateToken(ownerId string, claims map[string]interface{}) (auth.TokenDTO, error) {
	TokenPair, err := t.TokenProvider.GenerateToken(ownerId, claims)
	if err != nil {
		return auth.EmptyTokenDTO(), err
	}

	accessJwt, refreshJwt := TokenPair.AccessToken, TokenPair.RefreshToken
	refreshToken, err := t.TokenProvider.GetRefreshToken(refreshJwt)
	if err != nil {
		return auth.EmptyTokenDTO(), err
	}

	refreshToken.Id = t.GenerateId()
	if err := t.TokenStore.StoreToken(refreshToken); err != nil {
		return auth.EmptyTokenDTO(), err
	}

	return auth.NewTokenDTOWithRefresh(accessJwt, refreshToken.Id), nil
}

func (t JwtRefreshTokenWithRevokeProvider) GetAccessToken(accessToken string) (auth.Token, error) {
	return t.TokenProvider.GetAccessToken(accessToken)
}

func (t JwtRefreshTokenWithRevokeProvider) GetRefreshToken(refreshToken string) (auth.Token, error) {
	return t.TokenProvider.GetRefreshToken(refreshToken)
}

func (t JwtRefreshTokenWithRevokeProvider) GetAccessTokenProvider() auth.ITokenProvider {
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
