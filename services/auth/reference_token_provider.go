package auth

import (
	"github.com/google/uuid"
	"github.com/nayefradwi/nayef_go_common/core"
	"github.com/nayefradwi/nayef_go_common/modules/auth"
)

type JwtReferenceTokenProvider struct {
	tokenProvider            JwtRefreshTokenProvider
	referenceTokenRepository IReferenceTokenRepository[string]
}

func NewJwtReferenceTokenProvider(tokenProvider JwtRefreshTokenProvider, repo IReferenceTokenRepository[string]) JwtReferenceTokenProvider {
	return JwtReferenceTokenProvider{
		tokenProvider:            tokenProvider,
		referenceTokenRepository: repo,
	}
}

func NewDefaultJwtReferenceTokenProvider(repo IReferenceTokenRepository[string]) JwtReferenceTokenProvider {
	return NewJwtReferenceTokenProvider(NewDefaultJwtRefreshTokenProvider(), repo)
}

func (t JwtReferenceTokenProvider) GenerateId() string {
	id, _ := uuid.NewV7()
	return id.String()
}

func (t JwtReferenceTokenProvider) GenerateToken(ownerId string, claims map[string]interface{}) (ReferenceToken, error) {
	TokenPair, err := t.tokenProvider.GenerateToken(ownerId, claims)
	if err != nil {
		return ReferenceToken{}, err
	}

	accessJwt, refreshJwt := TokenPair.AccessToken, TokenPair.RefreshToken
	accessToken, err := t.tokenProvider.GetAccessToken(accessJwt)
	if err != nil {
		return ReferenceToken{}, err
	}

	refreshToken, err := t.tokenProvider.GetRefreshToken(refreshJwt)
	if err != nil {
		return ReferenceToken{}, err
	}

	accessTokenId, refreshTokenId := t.GenerateId(), t.GenerateId()
	accessToken.Id, refreshToken.Id = accessTokenId, refreshTokenId
	return t.referenceTokenRepository.StoreToken(accessToken, refreshToken)
}

func (t JwtReferenceTokenProvider) GetAccessToken(id string) (auth.Token[string], error) {
	return t.getToken(id, AccessTokenType)
}

func (t JwtReferenceTokenProvider) GetRefreshToken(id string) (auth.Token[string], error) {
	return t.getToken(id, RefreshTokenType)
}

func (t JwtReferenceTokenProvider) getToken(id string, tokenType int) (auth.Token[string], error) {
	token, err := t.referenceTokenRepository.GetTokenByReference(id, tokenType)
	if err != nil {
		return auth.Token[string]{}, core.UnauthorizedError("Token not found")
	}

	return token, nil
}

func (t JwtReferenceTokenProvider) RevokeToken(id string) error {
	return t.referenceTokenRepository.DeleteToken(id)
}

func (t JwtReferenceTokenProvider) RevokeOwner(ownerId string) error {
	return t.referenceTokenRepository.DeleteAllTokensByOwner(ownerId)
}
