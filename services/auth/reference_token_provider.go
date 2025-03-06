package auth

import (
	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/core"
	"github.com/nayefradwi/nayef_go_common/modules/auth"
)

type JwtReferenceTokenProvider struct {
	tokenProvider JwtRefreshTokenProvider
	tokenStore    auth.ITokenStore
}

func NewJwtReferenceTokenProvider(tokenProvider JwtRefreshTokenProvider, tokenStore auth.ITokenStore) JwtReferenceTokenProvider {
	return JwtReferenceTokenProvider{
		tokenProvider: tokenProvider,
		tokenStore:    tokenStore,
	}
}

func NewDefaultJwtReferenceTokenProvider(tokenStore auth.ITokenStore) JwtReferenceTokenProvider {
	return NewJwtReferenceTokenProvider(NewDefaultJwtRefreshTokenProvider(), tokenStore)
}

func (t JwtReferenceTokenProvider) GenerateId() string {
	id, _ := uuid.NewV7()
	return id.String()
}

func (t JwtReferenceTokenProvider) GenerateToken(ownerId string, claims map[string]interface{}) (auth.TokenDTO, error) {
	TokenPair, err := t.tokenProvider.GenerateToken(ownerId, claims)
	if err != nil {
		return auth.EmptyTokenDTO(), err
	}

	accessJwt, refreshJwt := TokenPair.AccessToken, TokenPair.RefreshToken
	accessToken, err := t.tokenProvider.GetAccessToken(accessJwt)
	if err != nil {
		return auth.EmptyTokenDTO(), err
	}

	refreshToken, err := t.tokenProvider.GetRefreshToken(refreshJwt)
	if err != nil {
		return auth.EmptyTokenDTO(), err
	}

	accessTokenId, refreshTokenId := t.GenerateId(), t.GenerateId()
	accessToken.Id, refreshToken.Id = accessTokenId, refreshTokenId
	if err := t.tokenStore.StoreTokens(accessToken, refreshToken); err != nil {
		return auth.EmptyTokenDTO(), err
	}

	return auth.NewTokenDTOWithRefresh(accessTokenId, refreshTokenId), nil
}

func (t JwtReferenceTokenProvider) GetAccessToken(id string) (auth.Token, error) {
	return t.getToken(id, auth.AccessTokenType)
}

func (t JwtReferenceTokenProvider) GetRefreshToken(id string) (auth.Token, error) {
	return t.getToken(id, auth.RefreshTokenType)
}

func (t JwtReferenceTokenProvider) getToken(id string, tokenType int) (auth.Token, error) {
	token, err := t.tokenStore.GetTokenByReference(id, tokenType)
	if err != nil {
		return auth.Token{}, core.UnauthorizedError("Token not found")
	}

	return token, nil
}

func (t JwtReferenceTokenProvider) RevokeToken(id string) error {
	return t.tokenStore.DeleteToken(id)
}

func (t JwtReferenceTokenProvider) RevokeOwner(ownerId string) error {
	return t.tokenStore.DeleteAllTokensByOwner(ownerId)
}

func (t JwtReferenceTokenProvider) GetAccessTokenProvider() auth.ITokenProvider {
	return t.tokenProvider.GetAccessTokenProvider()
}
