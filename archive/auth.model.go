package common

import (
	"time"
)

const defaultAccessTokenDuration = time.Hour * 24 * 7

var defaultRefreshTokenDuration = time.Hour * 24 * 30

type Token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type TokenOptions struct {
	AccessTokenExpiry  time.Time
	RefreshTokenExpiry time.Time
	Secret             string
}

func NewTokenOptions(accessTokenExpiry time.Time, refreshTokenExpiry time.Time) *TokenOptions {
	return &TokenOptions{
		AccessTokenExpiry:  accessTokenExpiry,
		RefreshTokenExpiry: refreshTokenExpiry,
	}
}
func DefaultTokenOptions() TokenOptions {
	return TokenOptions{
		AccessTokenExpiry:  time.Now().Add(defaultAccessTokenDuration),
		RefreshTokenExpiry: time.Now().Add(defaultRefreshTokenDuration),
	}
}
