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

func (t JwtReferenceTokenProvider) GenerateId() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", InternalError("failed to generate token id: " + err.Error())
	}
	return id.String(), nil
}

func (t JwtReferenceTokenProvider) GenerateToken(ownerId string, claims map[string]any) (TokenDTO, error) {
	tokenPair, err := t.tokenProvider.GenerateToken(ownerId, claims)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	accessToken, err := t.tokenProvider.GetAccessToken(tokenPair.AccessToken)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	refreshToken, err := t.tokenProvider.GetRefreshToken(tokenPair.RefreshToken)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	accessTokenId, err := t.GenerateId()
	if err != nil {
		return EmptyTokenDTO(), err
	}

	refreshTokenId, err := t.GenerateId()
	if err != nil {
		return EmptyTokenDTO(), err
	}

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
