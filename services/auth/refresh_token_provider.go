package auth

import (
	"time"

	"github.com/nayefradwi/nayef_go_common/modules/auth"
)

var defaultExpiresInRefreshToken = 30 * 24 * time.Hour
var DefaultJwtRefreshTokenProviderConfig = auth.ReplaceDefaultJwtExpiresIn(defaultExpiresInRefreshToken)
var DefaultJwtAccesTokenProviderConfig = auth.DefaultJwtTokenProviderConfig

type JwtRefreshTokenProvider struct {
	RefreshTokenProvider auth.JwtTokenProvider
	AccessTokenProvider  auth.JwtTokenProvider
}

func NewJwtRefreshTokenProvider(refreshTokenProvider auth.JwtTokenProvider, accessTokenProvider auth.JwtTokenProvider) JwtRefreshTokenProvider {
	return JwtRefreshTokenProvider{
		RefreshTokenProvider: refreshTokenProvider,
		AccessTokenProvider:  accessTokenProvider,
	}
}

func NewDefaultJwtRefreshTokenProvider() JwtRefreshTokenProvider {
	return NewJwtRefreshTokenProvider(
		auth.NewJwtTokenProvider(DefaultJwtRefreshTokenProviderConfig),
		auth.NewJwtTokenProvider(DefaultJwtAccesTokenProviderConfig),
	)
}

func (t JwtRefreshTokenProvider) GenerateToken(ownerId string, claims map[string]interface{}) (*RefreshToken, error) {
	accessToken, err := t.AccessTokenProvider.SignClaims(ownerId, claims)
	if err != nil {
		return nil, err
	}
	refreshToken, err := t.RefreshTokenProvider.SignClaims(ownerId, claims)
	if err != nil {
		return nil, err
	}
	return &RefreshToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (t JwtRefreshTokenProvider) GetAccessToken(accessToken string) (auth.Token, error) {
	return t.AccessTokenProvider.GetClaims(accessToken)
}

func (t JwtRefreshTokenProvider) GetRefreshToken(refreshToken string) (auth.Token, error) {
	return t.RefreshTokenProvider.GetClaims(refreshToken)
}

func (t JwtRefreshTokenProvider) GetAccessTokenProvider() auth.ITokenProvider {
	return t.AccessTokenProvider
}
