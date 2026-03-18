package auth

import (
	"github.com/google/uuid"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

type JwtReferenceTokenProvider struct {
	tokenProvider IRefreshTokenProvider
	tokenStore    ITokenStore
}

func NewJwtReferenceTokenProvider(tokenProvider IRefreshTokenProvider, tokenStore ITokenStore) IReferenceTokenProvider {
	return JwtReferenceTokenProvider{
		tokenProvider: tokenProvider,
		tokenStore:    tokenStore,
	}
}

func NewDefaultJwtReferenceTokenProvider(tokenStore ITokenStore) IReferenceTokenProvider {
	return NewJwtReferenceTokenProvider(NewDefaultJwtRefreshTokenProvider(), tokenStore)
}

func (t JwtReferenceTokenProvider) GenerateId() string {
	id, _ := uuid.NewV7()
	return id.String()
}

func (t JwtReferenceTokenProvider) GenerateToken(ownerId string, claims map[string]interface{}) (TokenDTO, error) {
	TokenPair, err := t.tokenProvider.GenerateToken(ownerId, claims)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	accessJwt, refreshJwt := TokenPair.AccessToken, TokenPair.RefreshToken
	accessToken, err := t.tokenProvider.GetAccessToken(accessJwt)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	refreshToken, err := t.tokenProvider.GetRefreshToken(refreshJwt)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	accessTokenId, refreshTokenId := t.GenerateId(), t.GenerateId()
	accessToken.Id, refreshToken.Id = accessTokenId, refreshTokenId
	if err := t.tokenStore.StoreTokens(accessToken, refreshToken); err != nil {
		return EmptyTokenDTO(), err
	}

	return NewTokenDTOWithRefresh(accessTokenId, refreshTokenId), nil
}

func (t JwtReferenceTokenProvider) GetAccessToken(id string) (Token, error) {
	return t.getToken(id, AccessTokenType)
}

func (t JwtReferenceTokenProvider) GetRefreshToken(id string) (Token, error) {
	return t.getToken(id, RefreshTokenType)
}

func (t JwtReferenceTokenProvider) getToken(id string, tokenType int) (Token, error) {
	token, err := t.tokenStore.GetTokenByReference(id, tokenType)
	if err != nil {
		return Token{}, UnauthorizedError("Token not found")
	}

	return token, nil
}

func (t JwtReferenceTokenProvider) RevokeToken(id string) error {
	return t.tokenStore.DeleteToken(id)
}

func (t JwtReferenceTokenProvider) RevokeOwner(ownerId string) error {
	return t.tokenStore.DeleteAllTokensByOwner(ownerId)
}

func (t JwtReferenceTokenProvider) GetAccessTokenProvider() ITokenProvider {
	return t.tokenProvider.GetAccessTokenProvider()
}
