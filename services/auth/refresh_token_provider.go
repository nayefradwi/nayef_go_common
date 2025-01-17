package auth

import (
	"time"

	"github.com/nayefradwi/nayef_go_common/modules/auth"
)

var defaultExpiresInRefreshToken = 30 * 24 * time.Hour
var DefaultJwtRefreshTokenProviderConfig = auth.ReplaceDefaultJwtExpiresIn(defaultExpiresInRefreshToken)
var DefaultJwtAccesTokenProviderConfig = auth.DefaultJwtTokenProviderConfig

type JwtRefreshTokenProvider[T string | int] struct {
	RefreshTokenProvider auth.JwtTokenProvider[T]
	AccessTokenProvider  auth.JwtTokenProvider[T]
}

func NewJwtRefreshTokenProvider[T string | int](refreshTokenProvider auth.JwtTokenProvider[T], accessTokenProvider auth.JwtTokenProvider[T]) JwtRefreshTokenProvider[T] {
	return JwtRefreshTokenProvider[T]{
		RefreshTokenProvider: refreshTokenProvider,
		AccessTokenProvider:  accessTokenProvider,
	}
}

func NewDefaultJwtRefreshTokenProvider[T string | int]() JwtRefreshTokenProvider[T] {
	return NewJwtRefreshTokenProvider[T](auth.NewJwtTokenProvider[T](DefaultJwtRefreshTokenProviderConfig), auth.NewJwtTokenProvider[T](DefaultJwtAccesTokenProviderConfig))
}

func (t JwtRefreshTokenProvider[T]) GenerateToken(ownerId T, claims map[string]interface{}) (*RefreshToken, error) {
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

func (t JwtRefreshTokenProvider[T]) GetAccessToken(accessToken string) (auth.Token[T], error) {
	return t.AccessTokenProvider.GetClaims(accessToken)
}

func (t JwtRefreshTokenProvider[T]) GetRefreshToken(refreshToken string) (auth.Token[T], error) {
	return t.RefreshTokenProvider.GetClaims(refreshToken)
}
